package diskio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend/core/models"
	"backend/core/ports"
)

// =====================================
// FileDiskRepository (MBR/Particiones)
// =====================================

type FileDiskRepository struct{}

func NewFileDiskRepository() *FileDiskRepository { return &FileDiskRepository{} }

func (*FileDiskRepository) NormalizePath(p string) string { return filepath.Clean(p) }

func (*FileDiskRepository) CreateDisk(path string, sizeBytes int64, fit rune) error {
	if sizeBytes <= 0 {
		return fmt.Errorf("tamaño inválido")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Rellena con 0s para tamaño exacto
	if err := f.Truncate(sizeBytes); err != nil {
		return err
	}

	// Escribe MBR al inicio
	mbr := models.MBR{
		Size:      sizeBytes,
		Timestamp: time.Now().Unix(),
		Signature: rand.Int64(),
		Fit:       byte(strings.ToUpper(string(fit))[0]),
	}
	if err := writeMBR(f, &mbr); err != nil {
		return err
	}
	return nil
}

func (*FileDiskRepository) RemoveDisk(path string) error {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("disco no existe: %s", path)
		}
		return err
	}
	return os.Remove(path)
}

func (*FileDiskRepository) ReadMBR(path string) (models.MBR, error) {
	var m models.MBR
	f, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		return m, err
	}
	defer f.Close()
	if err := readMBR(f, &m); err != nil {
		return m, err
	}
	return m, nil
}

func (*FileDiskRepository) WriteMBR(path string, m models.MBR) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	return writeMBR(f, &m)
}

func (*FileDiskRepository) DiskSignature(path string) (string, error) {
	m, err := (*FileDiskRepository)(nil).ReadMBR(path) // reuse receiver-less
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", m.Signature), nil
}

func (*FileDiskRepository) ValidatePrimaryForMount(path, name string) error {
	m, err := (*FileDiskRepository)(nil).ReadMBR(path)
	if err != nil {
		return err
	}
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed && p.Type == models.PartTypePrimary {
			n := strings.TrimRight(string(p.Name[:]), "\x00")
			if n == name {
				return nil
			}
		}
	}
	return fmt.Errorf("partición no encontrada o no es primaria: %s", name)
}

func (*FileDiskRepository) CreatePrimary(path string, part models.Partition) error {
	m, err := (*FileDiskRepository)(nil).ReadMBR(path)
	if err != nil {
		return err
	}
	// Reglas simples: nombre único y espacio disponible
	if nameExistsInMBR(&m, part.Name) {
		return fmt.Errorf("ya existe una partición con ese nombre")
	}
	slot := -1
	for i := 0; i < 4; i++ {
		if m.Partitions[i].Status == models.PartStatusFree {
			slot = i
			break
		}
	}
	if slot == -1 {
		return fmt.Errorf("MBR sin slots libres")
	}
	// Encuentra un hueco simple (después del MBR, ignorando EBRs para P simple)
	start, ok := findSpaceFor(&m, part.Size)
	if !ok {
		return fmt.Errorf("no hay espacio suficiente")
	}
	part.Status = models.PartStatusUsed
	part.Type = models.PartTypePrimary
	part.Start = start

	m.Partitions[slot] = part
	return (*FileDiskRepository)(nil).WriteMBR(path, m)
}

func (*FileDiskRepository) CreateExtended(path string, part models.Partition) error {
	m, err := (*FileDiskRepository)(nil).ReadMBR(path)
	if err != nil {
		return err
	}
	if nameExistsInMBR(&m, part.Name) {
		return fmt.Errorf("ya existe una partición con ese nombre")
	}
	// Solo una extendida
	for i := 0; i < 4; i++ {
		if m.Partitions[i].Status == models.PartStatusUsed && m.Partitions[i].Type == models.PartTypeExtend {
			return fmt.Errorf("ya existe una extendida")
		}
	}
	slot := -1
	for i := 0; i < 4; i++ {
		if m.Partitions[i].Status == models.PartStatusFree {
			slot = i
			break
		}
	}
	if slot == -1 {
		return fmt.Errorf("MBR sin slots libres")
	}
	start, ok := findSpaceFor(&m, part.Size)
	if !ok {
		return fmt.Errorf("no hay espacio suficiente")
	}
	part.Status = models.PartStatusUsed
	part.Type = models.PartTypeExtend
	part.Start = start
	m.Partitions[slot] = part

	// Inicializa un EBR “vacío” al inicio del área extendida
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	ebr := models.EBR{
		Status: models.PartStatusFree,
		Fit:    part.Fit,
		Start:  part.Start,
		Size:   0,
		Next:   -1,
	}
	if err := writeEBR(f, &ebr, ebr.Start); err != nil {
		return err
	}

	return (*FileDiskRepository)(nil).WriteMBR(path, m)
}

func (*FileDiskRepository) CreateLogical(path string, extStart int64, ebrNew models.EBR) error {
	// Encadena EBRs dentro del área extendida
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Recorre lista de EBRs
	cur := extStart
	var prev models.EBR
	for {
		var e models.EBR
		if err := readEBR(f, &e, cur); err != nil {
			return err
		}
		if e.Status == models.PartStatusFree && e.Size == 0 && e.Start == extStart {
			// Lista vacía → el primer EBR es el nuevo
			ebrNew.Next = -1
			if err := writeEBR(f, &ebrNew, extStart); err != nil {
				return err
			}
			break
		}
		if e.Next == -1 {
			// Insertar al final
			prev = e
			break
		}
		cur = e.Next
	}

	// prev es el último EBR “real”
	// Escribe nuevo ebr al final (al offset ebrNew.Start)
	if err := writeEBR(f, &ebrNew, ebrNew.Start); err != nil {
		return err
	}
	prev.Next = ebrNew.Start
	// Actualiza “prev” en disco
	if err := writeEBR(f, &prev, prev.Start); err != nil {
		return err
	}
	return nil
}

func (*FileDiskRepository) FindExtended(path string) (models.Partition, bool, error) {
	m, err := (*FileDiskRepository)(nil).ReadMBR(path)
	if err != nil {
		return models.Partition{}, false, err
	}
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed && p.Type == models.PartTypeExtend {
			return p, true, nil
		}
	}
	return models.Partition{}, false, nil
}

// --- helpers binarios ---

func writeMBR(f *os.File, m *models.MBR) error {
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	return binary.Write(f, binary.LittleEndian, m)
}
func readMBR(f *os.File, m *models.MBR) error {
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	return binary.Read(f, binary.LittleEndian, m)
}

func writeEBR(f *os.File, e *models.EBR, at int64) error {
	if _, err := f.Seek(at, 0); err != nil {
		return err
	}
	return binary.Write(f, binary.LittleEndian, e)
}
func readEBR(f *os.File, e *models.EBR, at int64) error {
	if _, err := f.Seek(at, 0); err != nil {
		return err
	}
	return binary.Read(f, binary.LittleEndian, e)
}

func trimName(b [16]byte) string {
	return strings.TrimRight(string(b[:]), "\x00")
}
func nameExistsInMBR(m *models.MBR, name [16]byte) bool {
	n := trimName(name)
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed && trimName(p.Name) == n {
			return true
		}
	}
	return false
}

// Encuentra un hueco libre simple para 'size' bytes (naïve: después de los usados, ordenando por Start)
func findSpaceFor(m *models.MBR, size int64) (int64, bool) {
	type seg struct{ s, e int64 }
	used := []seg{{0, int64(binary.Size(*m))}} // MBR reservado
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed {
			used = append(used, seg{p.Start, p.Start + p.Size})
		}
	}
	// ordenar por inicio
	for i := 0; i < len(used); i++ {
		for j := i + 1; j < len(used); j++ {
			if used[j].s < used[i].s {
				used[i], used[j] = used[j], used[i]
			}
		}
	}
	// buscar huecos entre segmentos
	var cur int64 = 0
	for _, u := range used {
		if u.s-cur >= size {
			return cur, true
		}
		if u.e > cur {
			cur = u.e
		}
	}
	// hueco al final
	if m.Size-cur >= size {
		return cur, true
	}
	return 0, false
}

// =====================================
// FileFsRepository (esqueleto EXT2)
// =====================================

type FileFsRepository struct {
	mounts ports.MountStore
	disk   *FileDiskRepository
}

func NewFileFsRepository(mounts ports.MountStore, disk *FileDiskRepository) *FileFsRepository {
	return &FileFsRepository{mounts: mounts, disk: disk}
}

func (r *FileFsRepository) resolve(id string) (diskPath string, region Region, err error) {
	// Busca el montaje por ID en la store y resuelve el Part.Start/Size leyendo MBR
	var ext ports.MountedEntry
	found := false
	for _, it := range r.mounts.List() {
		if it.ID == id {
			ext = it
			found = true
			break
		}
	}
	if !found {
		return "", Region{}, fmt.Errorf("id no montado: %s", id)
	}
	m, err := r.disk.ReadMBR(ext.Path)
	if err != nil {
		return "", Region{}, err
	}
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed && trimName(p.Name) == ext.Name {
			return ext.Path, Region{Start: p.Start, Size: p.Size}, nil
		}
	}
	return "", Region{}, fmt.Errorf("partición no encontrada: %s", ext.Name)
}

// -------- API de alto nivel (TODO: conecta tu EXT2) --------

func (r *FileFsRepository) Mkfs(id string) error {
	// TODO: implementa:
	// - calcular n (inodos y 3n bloques) según fórmula del enunciado
	// - escribir SuperBlock con offsets absolutos
	// - inicializar bitmaps en 0 y tabla de inodos/bloques
	// - crear / y users.txt con líneas iniciales
	return fmt.Errorf("Mkfs: TODO implementar")
}

func (r *FileFsRepository) Mkdir(id string, absPath []string, parents bool) error {
	// TODO: caminar desde / creando directorios, reservar inodo/bloques, actualizar bitmaps y SB
	return fmt.Errorf("Mkdir: TODO implementar")
}

func (r *FileFsRepository) Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool) error {
	// TODO: crear padres si recursive, reservar inodo y bloques para contenido, actualizar bitmaps y SB
	return fmt.Errorf("Mkfile: TODO implementar")
}

func (r *FileFsRepository) Cat(id string, files [][]string) (string, error) {
	// TODO: resolver inodo del archivo y concatenar contenidos de sus bloques
	return "", fmt.Errorf("Cat: TODO implementar")
}

// -------- Lecturas/escrituras finas (SB/bitmaps/inodos/bloques) --------

func (r *FileFsRepository) ReadSuper(id string) (models.SuperBlock, Region, error) {
	// TODO: lee SB desde region.Start (o desde offset guardado)
	return models.SuperBlock{}, Region{}, fmt.Errorf("ReadSuper: TODO implementar")
}

func (r *FileFsRepository) WriteSuper(id string, sb models.SuperBlock) error {
	// TODO: escribe SB
	return fmt.Errorf("WriteSuper: TODO implementar")
}

func (r *FileFsRepository) ReadBitmapInode(id string) ([]byte, error) {
	return nil, fmt.Errorf("ReadBitmapInode: TODO implementar")
}
func (r *FileFsRepository) ReadBitmapBlock(id string) ([]byte, error) {
	return nil, fmt.Errorf("ReadBitmapBlock: TODO implementar")
}
func (r *FileFsRepository) WriteBitmapInode(id string, data []byte) error {
	return fmt.Errorf("WriteBitmapInode: TODO implementar")
}
func (r *FileFsRepository) WriteBitmapBlock(id string, data []byte) error {
	return fmt.Errorf("WriteBitmapBlock: TODO implementar")
}

func (r *FileFsRepository) ReadInode(id string, index int) (models.Inode, error) {
	return models.Inode{}, fmt.Errorf("ReadInode: TODO implementar")
}
func (r *FileFsRepository) WriteInode(id string, index int, ino models.Inode) error {
	return fmt.Errorf("WriteInode: TODO implementar")
}

func (r *FileFsRepository) ReadBlock(id string, index int) ([]byte, error) {
	return nil, fmt.Errorf("ReadBlock: TODO implementar")
}
func (r *FileFsRepository) WriteBlock(id string, index int, data []byte) error {
	return fmt.Errorf("WriteBlock: TODO implementar")
}
