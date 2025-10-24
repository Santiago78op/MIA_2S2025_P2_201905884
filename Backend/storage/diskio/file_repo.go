package diskio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Backend/core/models"
	"Backend/core/ports"
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

	// Buscar en particiones primarias
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed && p.Type == models.PartTypePrimary {
			n := strings.TrimRight(string(p.Name[:]), "\x00")
			if n == name {
				return nil
			}
		}
	}

	// Buscar en particiones lógicas (dentro de la extendida)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("partición no encontrada: %s", name)
	}
	defer f.Close()

	// Encontrar la partición extendida
	var extStart int64 = -1
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed && p.Type == models.PartTypeExtend {
			extStart = p.Start
			break
		}
	}

	// Si hay partición extendida, buscar en sus particiones lógicas
	if extStart != -1 {
		cur := extStart
		for {
			var e models.EBR
			if err := readEBR(f, &e, cur); err != nil {
				break
			}

			// Si es el primer EBR vacío, terminar
			if e.Status == models.PartStatusFree && e.Size == 0 {
				break
			}

			// Si está en uso, verificar el nombre
			if e.Status == models.PartStatusUsed {
				n := strings.TrimRight(string(e.Name[:]), "\x00")
				if n == name {
					return nil
				}
			}

			// Si no hay siguiente, terminar
			if e.Next == -1 || e.Next == 0 {
				break
			}

			cur = e.Next
		}
	}

	return fmt.Errorf("partición no encontrada: %s", name)
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
	// Encuentra un hueco según el algoritmo de fit
	start, ok := findSpaceFor(&m, part.Size, part.Fit)
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
	start, ok := findSpaceFor(&m, part.Size, part.Fit)
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

	// Leer MBR para obtener el tamaño de la partición extendida
	var m models.MBR
	if err := readMBR(f, &m); err != nil {
		return err
	}

	// Encontrar la partición extendida
	var extPart models.Partition
	found := false
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed && p.Type == models.PartTypeExtend {
			extPart = p
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("no se encontró partición extendida")
	}

	// Recolectar todos los EBRs existentes
	var ebrs []ebrInfo

	cur := extStart
	for {
		var e models.EBR
		if err := readEBR(f, &e, cur); err != nil {
			return err
		}

		// Si es el primer EBR vacío, la lista está vacía
		if e.Status == models.PartStatusFree && e.Size == 0 && e.Start == extStart {
			break
		}

		// Agregar EBR a la lista
		ebrs = append(ebrs, ebrInfo{ebr: e, pos: cur})

		// Si no hay siguiente, terminamos
		if e.Next == -1 {
			break
		}

		cur = e.Next
	}

	// Encontrar espacio libre para el nuevo EBR
	// newEBRPos es donde se escribirá el header EBR en el disco
	newEBRPos, ok := findSpaceForLogical(extPart, ebrs, ebrNew.Size, ebrNew.Fit)
	if !ok {
		return fmt.Errorf("no hay espacio suficiente en la partición extendida")
	}

	// Calcular el tamaño del header EBR
	ebrHeaderSize := int64(binary.Size(models.EBR{}))

	// EBR.Start debe apuntar a los datos (después del header)
	ebrNew.Start = newEBRPos + ebrHeaderSize
	ebrNew.Status = models.PartStatusUsed
	ebrNew.Next = -1

	// Si no hay EBRs, este es el primero
	if len(ebrs) == 0 {
		// Escribir el primer EBR en el inicio de la extendida
		if err := writeEBR(f, &ebrNew, extStart); err != nil {
			return err
		}
		return nil
	}

	// Insertar el nuevo EBR en la posición correcta (ordenado por Start)
	// Encontrar dónde insertarlo
	insertIdx := -1
	for i, info := range ebrs {
		if ebrNew.Start < info.ebr.Start {
			insertIdx = i
			break
		}
	}

	if insertIdx == -1 {
		// Insertar al final
		lastIdx := len(ebrs) - 1
		prev := ebrs[lastIdx].ebr

		// Escribir nuevo EBR en su posición calculada
		if err := writeEBR(f, &ebrNew, newEBRPos); err != nil {
			return err
		}

		// Actualizar el anterior para que apunte a la posición del nuevo EBR
		prev.Next = newEBRPos
		if err := writeEBR(f, &prev, ebrs[lastIdx].pos); err != nil {
			return err
		}
	} else if insertIdx == 0 {
		// Insertar al principio
		// El nuevo EBR apunta a la posición del EBR que estaba primero
		ebrNew.Next = ebrs[0].pos

		// Escribir el nuevo EBR al inicio de la extendida
		if err := writeEBR(f, &ebrNew, extStart); err != nil {
			return err
		}
	} else {
		// Insertar en medio
		prev := ebrs[insertIdx-1].ebr

		// El nuevo apunta a la posición del siguiente EBR
		ebrNew.Next = ebrs[insertIdx].pos

		// Escribir nuevo EBR en su posición calculada
		if err := writeEBR(f, &ebrNew, newEBRPos); err != nil {
			return err
		}

		// El anterior apunta a la posición del nuevo EBR
		prev.Next = newEBRPos
		if err := writeEBR(f, &prev, ebrs[insertIdx-1].pos); err != nil {
			return err
		}
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

// Encuentra un hueco libre para 'size' bytes según el algoritmo de fit (FF, BF, WF)
func findSpaceFor(m *models.MBR, size int64, fit byte) (int64, bool) {
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

	// Recolectar todos los huecos disponibles
	type gap struct{ start, size int64 }
	var gaps []gap
	var cur int64 = 0
	for _, u := range used {
		if u.s-cur >= size {
			gaps = append(gaps, gap{cur, u.s - cur})
		}
		if u.e > cur {
			cur = u.e
		}
	}
	// hueco al final
	if m.Size-cur >= size {
		gaps = append(gaps, gap{cur, m.Size - cur})
	}

	if len(gaps) == 0 {
		return 0, false
	}

	// Aplicar algoritmo de fit
	switch fit {
	case 'F': // First Fit
		return gaps[0].start, true
	case 'B': // Best Fit
		bestIdx := 0
		for i := 1; i < len(gaps); i++ {
			if gaps[i].size < gaps[bestIdx].size {
				bestIdx = i
			}
		}
		return gaps[bestIdx].start, true
	case 'W': // Worst Fit
		worstIdx := 0
		for i := 1; i < len(gaps); i++ {
			if gaps[i].size > gaps[worstIdx].size {
				worstIdx = i
			}
		}
		return gaps[worstIdx].start, true
	default:
		// Por defecto First Fit
		return gaps[0].start, true
	}
}

// ebrInfo contiene información sobre un EBR existente
type ebrInfo struct {
	ebr models.EBR
	pos int64
}

// findSpaceForLogical encuentra espacio libre dentro de una partición extendida para una partición lógica
func findSpaceForLogical(extPart models.Partition, ebrs []ebrInfo, size int64, fit byte) (int64, bool) {
	type seg struct{ s, e int64 }

	// Calcular el tamaño de un EBR
	ebrSize := int64(binary.Size(models.EBR{}))

	// Áreas ocupadas dentro de la extendida
	var used []seg

	// Agregar todas las particiones lógicas existentes
	for _, info := range ebrs {
		// Cada partición lógica ocupa: EBR (en info.pos) + datos (en info.ebr.Start)
		// El EBR está en info.pos, y los datos van desde info.ebr.Start hasta info.ebr.Start + info.ebr.Size
		start := info.pos
		end := info.ebr.Start + info.ebr.Size
		used = append(used, seg{start, end})
	}

	// Ordenar por inicio
	for i := 0; i < len(used); i++ {
		for j := i + 1; j < len(used); j++ {
			if used[j].s < used[i].s {
				used[i], used[j] = used[j], used[i]
			}
		}
	}

	// Recolectar huecos disponibles
	type gap struct{ start, size int64 }
	var gaps []gap

	cur := extPart.Start
	extEnd := extPart.Start + extPart.Size

	for _, u := range used {
		if u.s-cur >= size+ebrSize {
			gaps = append(gaps, gap{cur, u.s - cur})
		}
		if u.e > cur {
			cur = u.e
		}
	}

	// Hueco al final
	if extEnd-cur >= size+ebrSize {
		gaps = append(gaps, gap{cur, extEnd - cur})
	}

	if len(gaps) == 0 {
		return 0, false
	}

	// Aplicar algoritmo de fit
	switch fit {
	case 'F': // First Fit
		return gaps[0].start, true
	case 'B': // Best Fit
		bestIdx := 0
		for i := 1; i < len(gaps); i++ {
			if gaps[i].size < gaps[bestIdx].size {
				bestIdx = i
			}
		}
		return gaps[bestIdx].start, true
	case 'W': // Worst Fit
		worstIdx := 0
		for i := 1; i < len(gaps); i++ {
			if gaps[i].size > gaps[worstIdx].size {
				worstIdx = i
			}
		}
		return gaps[worstIdx].start, true
	default:
		// Por defecto First Fit
		return gaps[0].start, true
	}
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

	// Buscar en particiones primarias
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed && trimName(p.Name) == ext.Name {
			return ext.Path, Region{Start: p.Start, Size: p.Size}, nil
		}
	}

	// Buscar en particiones lógicas (dentro de la extendida)
	f, err := os.Open(ext.Path)
	if err != nil {
		return "", Region{}, fmt.Errorf("partición no encontrada: %s", ext.Name)
	}
	defer f.Close()

	// Encontrar la partición extendida
	var extStart int64 = -1
	for _, p := range m.Partitions {
		if p.Status == models.PartStatusUsed && p.Type == models.PartTypeExtend {
			extStart = p.Start
			break
		}
	}

	// Si hay partición extendida, buscar en sus particiones lógicas
	if extStart != -1 {
		cur := extStart
		for {
			var e models.EBR
			if err := readEBR(f, &e, cur); err != nil {
				break
			}

			// Si es el primer EBR vacío, terminar
			if e.Status == models.PartStatusFree && e.Size == 0 {
				break
			}

			// Si está en uso, verificar el nombre
			if e.Status == models.PartStatusUsed {
				n := strings.TrimRight(string(e.Name[:]), "\x00")
				if n == ext.Name {
					return ext.Path, Region{Start: e.Start, Size: e.Size}, nil
				}
			}

			// Si no hay siguiente, terminar
			if e.Next == -1 || e.Next == 0 {
				break
			}

			cur = e.Next
		}
	}

	return "", Region{}, fmt.Errorf("partición no encontrada: %s", ext.Name)
}

// -------- API de alto nivel --------

func (r *FileFsRepository) Mkfs(id string, formatType string) error {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	// Validar tipo de formateo (por ahora solo soportamos "full")
	if formatType != "full" && formatType != "" {
		return fmt.Errorf("tipo de formateo no soportado: %s", formatType)
	}

	// Calcular n usando utils.ComputeLayout
	const inodeSize = 64
	const blockSize = models.BlockSizeFile // Usar constante actualizada (128)

	// Importar utils para usar ComputeLayout
	sb := computeLayout(region.Start, region.Size, inodeSize, blockSize)

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Escribir SuperBlock
	if err := writeSuperBlock(f, region.Start, &sb); err != nil {
		return fmt.Errorf("error escribiendo superblock: %w", err)
	}

	// Inicializar bitmaps a 0
	bmInodeSize := int(sb.SInodesCount)
	bmBlockSize := int(sb.SBlocksCount)

	bmInode := make([]byte, bmInodeSize)
	bmBlock := make([]byte, bmBlockSize)

	if _, err := SafeWriteAt(f, bmInode, sb.SBmInodeStart, region); err != nil {
		return fmt.Errorf("error escribiendo bitmap inodos: %w", err)
	}
	if _, err := SafeWriteAt(f, bmBlock, sb.SBmBlockStart, region); err != nil {
		return fmt.Errorf("error escribiendo bitmap bloques: %w", err)
	}

	// Inicializar tabla de inodos a -1 (libres)
	inodesCount := int(sb.SInodesCount)
	for i := 0; i < inodesCount; i++ {
		ino := models.Inode{
			IUid:   -1,
			IGid:   -1,
			ISize:  -1,
			IAtime: 0,
			ICtime: 0,
			IMtime: 0,
			IType:  255,
			IPerm:  [3]byte{'0', '0', '0'},
		}
		for j := range ino.IBlock {
			ino.IBlock[j] = -1
		}
		if err := writeInodeAt(f, sb.SInodeStart, i, &ino, region); err != nil {
			return fmt.Errorf("error inicializando inodo %d: %w", i, err)
		}
	}

	// Crear directorio raíz (inodo 0)
	now := time.Now().Unix()
	rootInode := models.Inode{
		IUid:   1,   // root
		IGid:   1,   // root
		ISize:  0,   // se actualizará
		IAtime: now,
		ICtime: now,
		IMtime: now,
		IType:  models.FileTypeFolder,
		IPerm:  [3]byte{'7', '5', '5'},
	}
	for j := range rootInode.IBlock {
		rootInode.IBlock[j] = -1
	}

	// Reservar bloque 0 para directorio raíz
	rootInode.IBlock[0] = 0
	rootBlock := models.FolderBlock{}
	// Inicializar entradas vacías
	for i := range rootBlock.Content {
		rootBlock.Content[i].BInodo = -1
	}

	// Entrada . (self)
	copy(rootBlock.Content[0].BName[:], ".")
	rootBlock.Content[0].BInodo = 0

	// Entrada .. (parent, también 0 para raíz)
	copy(rootBlock.Content[1].BName[:], "..")
	rootBlock.Content[1].BInodo = 0

	// Reservar inodo 1 para users.txt
	copy(rootBlock.Content[2].BName[:], "users.txt")
	rootBlock.Content[2].BInodo = 1

	rootInode.ISize = int32(binary.Size(rootBlock))

	// Escribir inodo raíz
	if err := writeInodeAt(f, sb.SInodeStart, 0, &rootInode, region); err != nil {
		return fmt.Errorf("error escribiendo inodo raíz: %w", err)
	}

	// Escribir bloque raíz
	if err := writeBlockAt(f, sb.SBlockStart, 0, &rootBlock, region); err != nil {
		return fmt.Errorf("error escribiendo bloque raíz: %w", err)
	}

	// Marcar inodo 0 y bloque 0 como usados en bitmaps
	bmInode[0] = 1
	bmBlock[0] = 1

	// Crear users.txt (inodo 1)
	usersContent := "1,G,root\n1,U,root,root,123\n"
	usersInode := models.Inode{
		IUid:   1,
		IGid:   1,
		ISize:  int32(len(usersContent)),
		IAtime: now,
		ICtime: now,
		IMtime: now,
		IType:  models.FileTypeRegular,
		IPerm:  [3]byte{'6', '6', '4'},
	}
	for j := range usersInode.IBlock {
		usersInode.IBlock[j] = -1
	}

	// Reservar bloque 1 para users.txt
	usersInode.IBlock[0] = 1
	usersBlock := models.FileBlock{}
	copy(usersBlock.Content[:], usersContent)

	if err := writeInodeAt(f, sb.SInodeStart, 1, &usersInode, region); err != nil {
		return fmt.Errorf("error escribiendo inodo users.txt: %w", err)
	}

	if err := writeBlockAt(f, sb.SBlockStart, 1, &usersBlock, region); err != nil {
		return fmt.Errorf("error escribiendo bloque users.txt: %w", err)
	}

	// Marcar inodo 1 y bloque 1 como usados
	bmInode[1] = 1
	bmBlock[1] = 1

	// Actualizar superblock con contadores
	sb.SFreeInodesCount = sb.SInodesCount - 2 // usados: 0 (/) y 1 (users.txt)
	sb.SFreeBlocksCount = sb.SBlocksCount - 2 // usados: 0 y 1
	sb.SMtime = now

	// Escribir bitmaps actualizados
	if _, err := SafeWriteAt(f, bmInode, sb.SBmInodeStart, region); err != nil {
		return fmt.Errorf("error actualizando bitmap inodos: %w", err)
	}
	if _, err := SafeWriteAt(f, bmBlock, sb.SBmBlockStart, region); err != nil {
		return fmt.Errorf("error actualizando bitmap bloques: %w", err)
	}

	// Escribir superblock actualizado
	if err := writeSuperBlock(f, region.Start, &sb); err != nil {
		return fmt.Errorf("error actualizando superblock: %w", err)
	}

	return nil
}

func (r *FileFsRepository) Mkdir(id string, absPath []string, parents bool, uid int, gid int) error {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Detectar tipo de filesystem y leer superblock apropiado
	sbInterface, inodeStart, blockStart, err := r.readSuperBlockUniversal(f, region.Start)
	if err != nil {
		return fmt.Errorf("error leyendo superblock: %w", err)
	}

	// Extraer campos comunes del superblock
	var bmInodeStart, bmBlockStart int64
	var inodesCount, blocksCount int32

	switch sb := sbInterface.(type) {
	case models.SuperBlock:
		// EXT2
		bmInodeStart = sb.SBmInodeStart
		bmBlockStart = sb.SBmBlockStart
		inodesCount = sb.SInodesCount
		blocksCount = sb.SBlocksCount
	case models.SuperBlockExt3:
		// EXT3
		bmInodeStart = sb.BmInodeStart
		bmBlockStart = sb.BmBlockStart
		inodesCount = sb.InodesCount
		blocksCount = sb.BlocksCount
	default:
		return fmt.Errorf("tipo de superblock desconocido")
	}

	// Leer bitmaps
	bmInode, err := readBitmap(f, bmInodeStart, int(inodesCount), region)
	if err != nil {
		return fmt.Errorf("error leyendo bitmap inodos: %w", err)
	}
	bmBlock, err := readBitmap(f, bmBlockStart, int(blocksCount), region)
	if err != nil {
		return fmt.Errorf("error leyendo bitmap bloques: %w", err)
	}

	// Navegar desde la raíz (inodo 0)
	currentInodeIdx := int32(0)

	for i, name := range absPath {
		isLast := i == len(absPath)-1

		// Leer inodo actual
		currentInode, err := readInodeAt(f, inodeStart, int(currentInodeIdx), region)
		if err != nil {
			return fmt.Errorf("error leyendo inodo %d: %w", currentInodeIdx, err)
		}

		// Verificar que es un directorio
		if currentInode.IType != models.FileTypeFolder {
			return fmt.Errorf("'%s' no es un directorio", strings.Join(absPath[:i], "/"))
		}

		// Verificar permisos de escritura en el directorio padre (antes de crear)
		if !hasWritePermission(currentInode, int32(uid), int32(gid)) {
			parentPath := "/"
			if i > 0 {
				parentPath = "/" + strings.Join(absPath[:i], "/")
			}
			return fmt.Errorf("no tiene permiso de escritura en: %s", parentPath)
		}

		// Buscar entrada 'name' en el directorio actual
		foundIdx, err := findEntryInDir(f, &currentInode, blockStart, name, region)
		if err != nil {
			return fmt.Errorf("error buscando entrada: %w", err)
		}

		if foundIdx == -1 {
			// No existe
			if !isLast && !parents {
				return fmt.Errorf("directorio padre no existe y -p no está activo")
			}

			// Crear nuevo directorio
			newInodeIdx, newBlockIdx, err := allocateInodeAndBlock(bmInode, bmBlock, int(inodesCount), int(blocksCount))
			if err != nil {
				return fmt.Errorf("error reservando inodo/bloque: %w", err)
			}

			// Crear inodo para el nuevo directorio
			now := time.Now().Unix()
			newInode := models.Inode{
				IUid:   int32(uid), // UID del usuario que crea el directorio
				IGid:   int32(gid), // GID del usuario que crea el directorio
				ISize:  0,
				IAtime: now,
				ICtime: now,
				IMtime: now,
				IType:  models.FileTypeFolder,
				IPerm:  [3]byte{'6', '6', '4'}, // Permisos 664 según la guía
			}
			for j := range newInode.IBlock {
				newInode.IBlock[j] = -1
			}
			newInode.IBlock[0] = int32(newBlockIdx)

			// Crear bloque de directorio con . y ..
			newDirBlock := models.FolderBlock{}
			for j := range newDirBlock.Content {
				newDirBlock.Content[j].BInodo = -1
			}
			copy(newDirBlock.Content[0].BName[:], ".")
			newDirBlock.Content[0].BInodo = int32(newInodeIdx)
			copy(newDirBlock.Content[1].BName[:], "..")
			newDirBlock.Content[1].BInodo = currentInodeIdx

			newInode.ISize = int32(binary.Size(newDirBlock))

			// Escribir inodo y bloque
			if err := writeInodeAt(f, inodeStart, newInodeIdx, &newInode, region); err != nil {
				return fmt.Errorf("error escribiendo inodo nuevo: %w", err)
			}
			if err := writeBlockAt(f, blockStart, newBlockIdx, &newDirBlock, region); err != nil {
				return fmt.Errorf("error escribiendo bloque nuevo: %w", err)
			}

			// Agregar entrada al directorio padre
			if err := addEntryToDir(f, &currentInode, blockStart, name, int32(newInodeIdx), bmBlock, int(blocksCount), region); err != nil {
				return fmt.Errorf("error agregando entrada al directorio padre: %w", err)
			}

			// Actualizar inodo padre
			currentInode.IMtime = now
			if err := writeInodeAt(f, inodeStart, int(currentInodeIdx), &currentInode, region); err != nil {
				return fmt.Errorf("error actualizando inodo padre: %w", err)
			}

			// Actualizar superblock según el tipo
			now = time.Now().Unix()
			switch sb := sbInterface.(type) {
			case models.SuperBlock:
				sb.SFreeInodesCount--
				sb.SFreeBlocksCount--
				sb.SMtime = now
				if err := writeSuperBlock(f, region.Start, &sb); err != nil {
					return fmt.Errorf("error actualizando superblock: %w", err)
				}
			case models.SuperBlockExt3:
				sb.FreeInodes--
				sb.FreeBlocks--
				sb.MTime = now
				if err := r.writeSuperExt3(f, sb); err != nil {
					return fmt.Errorf("error actualizando superblock: %w", err)
				}
			}

			// Continuar con el nuevo directorio
			currentInodeIdx = int32(newInodeIdx)
		} else {
			// Ya existe
			if isLast {
				return fmt.Errorf("el directorio ya existe: %s", name)
			}
			currentInodeIdx = foundIdx
		}
	}

	// Escribir bitmaps actualizados
	if _, err := SafeWriteAt(f, bmInode, bmInodeStart, region); err != nil {
		return fmt.Errorf("error actualizando bitmap inodos: %w", err)
	}
	if _, err := SafeWriteAt(f, bmBlock, bmBlockStart, region); err != nil {
		return fmt.Errorf("error actualizando bitmap bloques: %w", err)
	}

	// Registrar en journal (ignorar errores para no romper la operación)
	fullPath := "/" + strings.Join(absPath, "/")
	_ = r.JournalAppendPublic(id, "MKDIR", fullPath, "", time.Now().Unix())

	return nil
}

func (r *FileFsRepository) Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool, uid int, gid int) error {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Detectar tipo de filesystem y leer superblock apropiado
	sbInterface, inodeStart, blockStart, err := r.readSuperBlockUniversal(f, region.Start)
	if err != nil {
		return fmt.Errorf("error leyendo superblock: %w", err)
	}

	// Extraer campos comunes del superblock
	var bmInodeStart, bmBlockStart int64
	var inodesCount, blocksCount int32

	switch sb := sbInterface.(type) {
	case models.SuperBlock:
		// EXT2
		bmInodeStart = sb.SBmInodeStart
		bmBlockStart = sb.SBmBlockStart
		inodesCount = sb.SInodesCount
		blocksCount = sb.SBlocksCount
	case models.SuperBlockExt3:
		// EXT3
		bmInodeStart = sb.BmInodeStart
		bmBlockStart = sb.BmBlockStart
		inodesCount = sb.InodesCount
		blocksCount = sb.BlocksCount
	default:
		return fmt.Errorf("tipo de superblock desconocido")
	}

	// Leer bitmaps
	bmInode, err := readBitmap(f, bmInodeStart, int(inodesCount), region)
	if err != nil {
		return fmt.Errorf("error leyendo bitmap inodos: %w", err)
	}
	bmBlock, err := readBitmap(f, bmBlockStart, int(blocksCount), region)
	if err != nil {
		return fmt.Errorf("error leyendo bitmap bloques: %w", err)
	}

	// Navegar al directorio padre, creándolo si recursive=true
	if len(absPath) == 0 {
		return fmt.Errorf("path vacío")
	}

	parentPath := absPath[:len(absPath)-1]
	fileName := absPath[len(absPath)-1]

	// Navegar al directorio padre
	currentInodeIdx := int32(0) // raíz

	for i, name := range parentPath {
		// Leer inodo actual
		currentInode, err := readInodeAt(f, inodeStart, int(currentInodeIdx), region)
		if err != nil {
			return fmt.Errorf("error leyendo inodo %d: %w", currentInodeIdx, err)
		}

		// Verificar que es un directorio
		if currentInode.IType != models.FileTypeFolder {
			return fmt.Errorf("'%s' no es un directorio", strings.Join(parentPath[:i], "/"))
		}

		// Verificar permisos de escritura en el directorio
		if !hasWritePermission(currentInode, int32(uid), int32(gid)) {
			dirPath := "/"
			if i > 0 {
				dirPath = "/" + strings.Join(parentPath[:i], "/")
			}
			return fmt.Errorf("no tiene permiso de escritura en: %s", dirPath)
		}

		// Buscar entrada
		foundIdx, err := findEntryInDir(f, &currentInode, blockStart, name, region)
		if err != nil {
			return fmt.Errorf("error buscando entrada: %w", err)
		}

		if foundIdx == -1 {
			// No existe
			if !recursive {
				return fmt.Errorf("directorio padre no existe y -p no está activo")
			}

			// Crear directorio padre (similar a Mkdir)
			newInodeIdx, newBlockIdx, err := allocateInodeAndBlock(bmInode, bmBlock, int(inodesCount), int(blocksCount))
			if err != nil {
				return fmt.Errorf("error reservando inodo/bloque: %w", err)
			}

			now := time.Now().Unix()
			newInode := models.Inode{
				IUid:   int32(uid), // UID del usuario que crea el directorio
				IGid:   int32(gid), // GID del usuario que crea el directorio
				ISize:  0,
				IAtime: now,
				ICtime: now,
				IMtime: now,
				IType:  models.FileTypeFolder,
				IPerm:  [3]byte{'6', '6', '4'}, // Permisos 664 según la guía
			}
			for j := range newInode.IBlock {
				newInode.IBlock[j] = -1
			}
			newInode.IBlock[0] = int32(newBlockIdx)

			newDirBlock := models.FolderBlock{}
			for j := range newDirBlock.Content {
				newDirBlock.Content[j].BInodo = -1
			}
			copy(newDirBlock.Content[0].BName[:], ".")
			newDirBlock.Content[0].BInodo = int32(newInodeIdx)
			copy(newDirBlock.Content[1].BName[:], "..")
			newDirBlock.Content[1].BInodo = currentInodeIdx

			newInode.ISize = int32(binary.Size(newDirBlock))

			if err := writeInodeAt(f, inodeStart, newInodeIdx, &newInode, region); err != nil {
				return fmt.Errorf("error escribiendo inodo nuevo: %w", err)
			}
			if err := writeBlockAt(f, blockStart, newBlockIdx, &newDirBlock, region); err != nil {
				return fmt.Errorf("error escribiendo bloque nuevo: %w", err)
			}

			if err := addEntryToDir(f, &currentInode, blockStart, name, int32(newInodeIdx), bmBlock, int(blocksCount), region); err != nil {
				return fmt.Errorf("error agregando entrada al directorio padre: %w", err)
			}

			currentInode.IMtime = now
			if err := writeInodeAt(f, inodeStart, int(currentInodeIdx), &currentInode, region); err != nil {
				return fmt.Errorf("error actualizando inodo padre: %w", err)
			}

			// Nota: Los contadores de superblock se actualizan al final

			currentInodeIdx = int32(newInodeIdx)
		} else {
			currentInodeIdx = foundIdx
		}
	}

	// Ahora crear el archivo en el directorio padre
	parentInode, err := readInodeAt(f, inodeStart, int(currentInodeIdx), region)
	if err != nil {
		return fmt.Errorf("error leyendo inodo padre: %w", err)
	}

	// Verificar que el archivo no existe
	existingIdx, err := findEntryInDir(f, &parentInode, blockStart, fileName, region)
	if err != nil {
		return fmt.Errorf("error buscando archivo: %w", err)
	}
	if existingIdx != -1 {
		return fmt.Errorf("el archivo ya existe: %s", fileName)
	}

	// Leer contenido desde archivo host si se especifica (prioridad sobre -size)
	var content []byte
	if contentHostPath != "" {
		// Intentar leer como archivo primero
		content, err = os.ReadFile(contentHostPath)
		if err != nil {
			// Si falla, usar el valor como contenido directo
			content = []byte(contentHostPath)
		}
	} else {
		// Generar contenido con números 0-9 repetidos según el tamaño
		// Según la guía: "El contenido serán números del 0 al 9 cuantas veces sea necesario"
		content = make([]byte, size)
		for i := 0; i < size; i++ {
			content[i] = byte('0' + (i % 10))
		}
	}

	// Calcular cuántos bloques necesitamos
	blocksNeeded := (len(content) + models.BlockSizeFile - 1) / models.BlockSizeFile // redondear hacia arriba

	// Reservar inodo para el archivo
	fileInodeIdx := -1
	for i := 0; i < int(inodesCount); i++ {
		if bmInode[i] == 0 {
			bmInode[i] = 1
			fileInodeIdx = i
			break
		}
	}
	if fileInodeIdx == -1 {
		return fmt.Errorf("no hay inodos libres")
	}

	// Reservar bloques para el contenido
	blockIndices := make([]int32, 0, blocksNeeded)
	for i := 0; i < blocksNeeded && len(blockIndices) < blocksNeeded; i++ {
		for j := 0; j < int(blocksCount); j++ {
			if bmBlock[j] == 0 {
				bmBlock[j] = 1
				blockIndices = append(blockIndices, int32(j))
				break
			}
		}
	}
	if len(blockIndices) < blocksNeeded {
		// Revertir asignaciones
		bmInode[fileInodeIdx] = 0
		for _, blkIdx := range blockIndices {
			bmBlock[blkIdx] = 0
		}
		return fmt.Errorf("no hay suficientes bloques libres")
	}

	// Crear inodo del archivo
	now := time.Now().Unix()
	fileInode := models.Inode{
		IUid:   int32(uid), // UID del usuario que crea el archivo
		IGid:   int32(gid), // GID del usuario que crea el archivo
		ISize:  int32(len(content)),
		IAtime: now,
		ICtime: now,
		IMtime: now,
		IType:  models.FileTypeRegular,
		IPerm:  [3]byte{'6', '6', '4'}, // Permisos 664 según la guía
	}
	for j := range fileInode.IBlock {
		fileInode.IBlock[j] = -1
	}

	// Asignar bloques directos (hasta 12)
	for i := 0; i < len(blockIndices) && i < 12; i++ {
		fileInode.IBlock[i] = blockIndices[i]
	}

	// Escribir contenido en bloques
	for i, blkIdx := range blockIndices {
		start := i * models.BlockSizeFile
		end := start + models.BlockSizeFile
		if end > len(content) {
			end = len(content)
		}

		var fileBlock models.FileBlock
		copy(fileBlock.Content[:], content[start:end])

		if err := writeBlockAt(f, blockStart, int(blkIdx), &fileBlock, region); err != nil {
			return fmt.Errorf("error escribiendo bloque de archivo: %w", err)
		}
	}

	// Escribir inodo del archivo
	if err := writeInodeAt(f, inodeStart, fileInodeIdx, &fileInode, region); err != nil {
		return fmt.Errorf("error escribiendo inodo de archivo: %w", err)
	}

	// Agregar entrada al directorio padre
	if err := addEntryToDir(f, &parentInode, blockStart, fileName, int32(fileInodeIdx), bmBlock, int(blocksCount), region); err != nil {
		return fmt.Errorf("error agregando entrada al directorio padre: %w", err)
	}

	// Actualizar inodo padre
	parentInode.IMtime = now
	if err := writeInodeAt(f, inodeStart, int(currentInodeIdx), &parentInode, region); err != nil {
		return fmt.Errorf("error actualizando inodo padre: %w", err)
	}

	// Actualizar superblock según el tipo
	switch sb := sbInterface.(type) {
	case models.SuperBlock:
		sb.SFreeInodesCount -= int32(1)
		sb.SFreeBlocksCount -= int32(len(blockIndices))
		sb.SMtime = now
		if err := writeSuperBlock(f, region.Start, &sb); err != nil {
			return fmt.Errorf("error actualizando superblock: %w", err)
		}
	case models.SuperBlockExt3:
		sb.FreeInodes -= int32(1)
		sb.FreeBlocks -= int32(len(blockIndices))
		sb.MTime = now
		if err := r.writeSuperExt3(f, sb); err != nil {
			return fmt.Errorf("error actualizando superblock: %w", err)
		}
	}

	// Escribir bitmaps actualizados
	if _, err := SafeWriteAt(f, bmInode, bmInodeStart, region); err != nil {
		return fmt.Errorf("error actualizando bitmap inodos: %w", err)
	}
	if _, err := SafeWriteAt(f, bmBlock, bmBlockStart, region); err != nil {
		return fmt.Errorf("error actualizando bitmap bloques: %w", err)
	}

	// Registrar en journal (ignorar errores para no romper la operación)
	fullPath := "/" + strings.Join(absPath, "/")
	contentInfo := fmt.Sprintf("size=%d", size)
	_ = r.JournalAppendPublic(id, "MKFILE", fullPath, contentInfo, time.Now().Unix())

	return nil
}

func (r *FileFsRepository) Cat(id string, files [][]string, uid int, gid int) (string, error) {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Detectar tipo de filesystem y leer superblock apropiado
	_, inodeStart, blockStart, err := r.readSuperBlockUniversal(f, region.Start)
	if err != nil {
		return "", fmt.Errorf("error leyendo superblock: %w", err)
	}

	var result strings.Builder

	for _, filePath := range files {
		if len(filePath) == 0 {
			continue
		}

		// Navegar hasta el archivo
		currentInodeIdx := int32(0) // raíz

		for i, name := range filePath {
			isLast := i == len(filePath)-1

			// Leer inodo actual
			currentInode, err := readInodeAt(f, inodeStart, int(currentInodeIdx), region)
			if err != nil {
				return "", fmt.Errorf("error leyendo inodo %d: %w", currentInodeIdx, err)
			}

			if !isLast {
				// Debe ser un directorio
				if currentInode.IType != models.FileTypeFolder {
					return "", fmt.Errorf("'%s' no es un directorio", strings.Join(filePath[:i], "/"))
				}

				// Buscar siguiente componente
				foundIdx, err := findEntryInDir(f, &currentInode, blockStart, name, region)
				if err != nil {
					return "", fmt.Errorf("error buscando entrada: %w", err)
				}
				if foundIdx == -1 {
					return "", fmt.Errorf("no existe: %s", strings.Join(filePath[:i+1], "/"))
				}
				currentInodeIdx = foundIdx
			} else {
				// Es el último componente, buscar el archivo
				foundIdx, err := findEntryInDir(f, &currentInode, blockStart, name, region)
				if err != nil {
					return "", fmt.Errorf("error buscando archivo: %w", err)
				}
				if foundIdx == -1 {
					return "", fmt.Errorf("archivo no existe: %s", strings.Join(filePath, "/"))
				}

				// Leer inodo del archivo
				fileInode, err := readInodeAt(f, inodeStart, int(foundIdx), region)
				if err != nil {
					return "", fmt.Errorf("error leyendo inodo de archivo: %w", err)
				}

				// Verificar que es un archivo
				if fileInode.IType != models.FileTypeRegular {
					return "", fmt.Errorf("no es un archivo regular: %s", strings.Join(filePath, "/"))
				}

				// Verificar permisos de lectura
				if !hasReadPermission(fileInode, int32(uid), int32(gid)) {
					return "", fmt.Errorf("no tiene permiso de lectura para el archivo: %s", strings.Join(filePath, "/"))
				}

				// Leer contenido de los bloques
				content := make([]byte, 0, fileInode.ISize)
				for i := 0; i < 12; i++ {
					blkIdx := fileInode.IBlock[i]
					if blkIdx == -1 {
						break
					}

					// Leer bloque
					var fileBlock models.FileBlock
					offset := blockStart + int64(blkIdx*models.BlockSizeFile)
					if _, err := f.Seek(offset, 0); err != nil {
						return "", fmt.Errorf("error buscando bloque: %w", err)
					}
					if err := binary.Read(f, binary.LittleEndian, &fileBlock); err != nil {
						return "", fmt.Errorf("error leyendo bloque: %w", err)
					}

					// Agregar contenido
					blockContent := fileBlock.Content[:]
					if len(content)+len(blockContent) > int(fileInode.ISize) {
						// Último bloque parcial
						remaining := int(fileInode.ISize) - len(content)
						content = append(content, blockContent[:remaining]...)
						break
					}
					content = append(content, blockContent...)
				}

				// Agregar al resultado
				// Si es users.txt, filtrar líneas con ID=0 (eliminadas)
				fileName := strings.Join(filePath, "/")
				if fileName == "users.txt" || fileName == "/users.txt" {
					filteredContent := filterDeletedEntries(string(content))
					result.WriteString(filteredContent)
				} else {
					result.Write(content)
				}
				result.WriteString("\n")
			}
		}
	}

	return result.String(), nil
}

// filterDeletedEntries filtra las líneas de users.txt que tienen ID=0 (eliminadas)
func filterDeletedEntries(content string) string {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	var filtered []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extraer el primer campo (ID)
		parts := strings.Split(line, ",")
		if len(parts) >= 2 {
			id := strings.TrimSpace(parts[0])
			// Solo incluir si el ID no es "0"
			if id != "0" {
				filtered = append(filtered, line)
			}
		}
	}

	return strings.Join(filtered, "\n")
}

// UpdateFile actualiza el contenido de un archivo existente
func (r *FileFsRepository) UpdateFile(id string, absPath []string, newContent string) error {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Detectar tipo de filesystem y leer superblock apropiado
	sbInterface, inodeStart, blockStart, err := r.readSuperBlockUniversal(f, region.Start)
	if err != nil {
		return fmt.Errorf("error leyendo superblock: %w", err)
	}

	// Extraer campos comunes del superblock
	var bmBlockStart int64
	var blocksCount int32

	switch sb := sbInterface.(type) {
	case models.SuperBlock:
		// EXT2
		bmBlockStart = sb.SBmBlockStart
		blocksCount = sb.SBlocksCount
	case models.SuperBlockExt3:
		// EXT3
		bmBlockStart = sb.BmBlockStart
		blocksCount = sb.BlocksCount
	default:
		return fmt.Errorf("tipo de superblock desconocido")
	}

	// Leer bitmaps
	bmBlock, err := readBitmap(f, bmBlockStart, int(blocksCount), region)
	if err != nil {
		return fmt.Errorf("error leyendo bitmap bloques: %w", err)
	}

	// Navegar hasta el archivo
	currentInodeIdx := int32(0) // raíz

	for i, name := range absPath {
		isLast := i == len(absPath)-1

		// Leer inodo actual
		currentInode, err := readInodeAt(f, inodeStart, int(currentInodeIdx), region)
		if err != nil {
			return fmt.Errorf("error leyendo inodo %d: %w", currentInodeIdx, err)
		}

		if !isLast {
			// Debe ser un directorio
			if currentInode.IType != models.FileTypeFolder {
				return fmt.Errorf("'%s' no es un directorio", strings.Join(absPath[:i], "/"))
			}

			// Buscar siguiente componente
			foundIdx, err := findEntryInDir(f, &currentInode, blockStart, name, region)
			if err != nil {
				return fmt.Errorf("error buscando entrada: %w", err)
			}
			if foundIdx == -1 {
				return fmt.Errorf("no existe: %s", strings.Join(absPath[:i+1], "/"))
			}
			currentInodeIdx = foundIdx
		} else {
			// Es el último componente, buscar el archivo
			foundIdx, err := findEntryInDir(f, &currentInode, blockStart, name, region)
			if err != nil {
				return fmt.Errorf("error buscando archivo: %w", err)
			}
			if foundIdx == -1 {
				return fmt.Errorf("archivo no existe: %s", strings.Join(absPath, "/"))
			}

			// Leer inodo del archivo
			fileInode, err := readInodeAt(f, inodeStart, int(foundIdx), region)
			if err != nil {
				return fmt.Errorf("error leyendo inodo de archivo: %w", err)
			}

			// Verificar que es un archivo
			if fileInode.IType != models.FileTypeRegular {
				return fmt.Errorf("no es un archivo regular: %s", strings.Join(absPath, "/"))
			}

			// Convertir contenido a bytes
			content := []byte(newContent)
			blocksNeeded := (len(content) + models.BlockSizeFile - 1) / models.BlockSizeFile

			// Liberar bloques antiguos si los hay
			for i := 0; i < 12; i++ {
				if fileInode.IBlock[i] != -1 {
					bmBlock[fileInode.IBlock[i]] = 0
					fileInode.IBlock[i] = -1
				}
			}

			// Asignar nuevos bloques
			blockIndices := make([]int32, 0, blocksNeeded)
			for i := 0; i < blocksNeeded; i++ {
				for j := 0; j < int(blocksCount); j++ {
					if bmBlock[j] == 0 {
						bmBlock[j] = 1
						blockIndices = append(blockIndices, int32(j))
						break
					}
				}
			}

			if len(blockIndices) < blocksNeeded {
				return fmt.Errorf("no hay suficientes bloques libres")
			}

			// Actualizar inodo
			fileInode.ISize = int32(len(content))
			fileInode.IMtime = time.Now().Unix()

			// Asignar bloques al inodo
			for i := 0; i < len(blockIndices) && i < 12; i++ {
				fileInode.IBlock[i] = blockIndices[i]
			}

			// Escribir contenido en bloques
			for i, blkIdx := range blockIndices {
				start := i * models.BlockSizeFile
				end := start + models.BlockSizeFile
				if end > len(content) {
					end = len(content)
				}

				var fileBlock models.FileBlock
				copy(fileBlock.Content[:], content[start:end])

				if err := writeBlockAt(f, blockStart, int(blkIdx), &fileBlock, region); err != nil {
					return fmt.Errorf("error escribiendo bloque de archivo: %w", err)
				}
			}

			// Escribir inodo actualizado
			if err := writeInodeAt(f, inodeStart, int(foundIdx), &fileInode, region); err != nil {
				return fmt.Errorf("error escribiendo inodo actualizado: %w", err)
			}

			// Actualizar superblock según el tipo
			now := time.Now().Unix()
			switch sb := sbInterface.(type) {
			case models.SuperBlock:
				sb.SMtime = now
				if err := writeSuperBlock(f, region.Start, &sb); err != nil {
					return fmt.Errorf("error actualizando superblock: %w", err)
				}
			case models.SuperBlockExt3:
				sb.MTime = now
				if err := r.writeSuperExt3(f, sb); err != nil {
					return fmt.Errorf("error actualizando superblock: %w", err)
				}
			}
		}
	}

	// Escribir bitmap actualizado
	if _, err := SafeWriteAt(f, bmBlock, bmBlockStart, region); err != nil {
		return fmt.Errorf("error actualizando bitmap bloques: %w", err)
	}

	return nil
}

// -------- Lecturas/escrituras finas (SB/bitmaps/inodos/bloques) --------

func (r *FileFsRepository) ReadSuper(id string) (models.SuperBlock, Region, error) {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return models.SuperBlock{}, Region{}, err
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return models.SuperBlock{}, Region{}, err
	}
	defer f.Close()

	sb, err := readSuperBlock(f, region.Start)
	if err != nil {
		return models.SuperBlock{}, Region{}, err
	}

	return sb, region, nil
}

func (r *FileFsRepository) WriteSuper(id string, sb models.SuperBlock) error {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	return writeSuperBlock(f, region.Start, &sb)
}

func (r *FileFsRepository) ReadBitmapInode(id string) ([]byte, error) {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sb, err := readSuperBlock(f, region.Start)
	if err != nil {
		return nil, err
	}

	return readBitmap(f, sb.SBmInodeStart, int(sb.SInodesCount), region)
}

func (r *FileFsRepository) ReadBitmapBlock(id string) ([]byte, error) {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sb, err := readSuperBlock(f, region.Start)
	if err != nil {
		return nil, err
	}

	return readBitmap(f, sb.SBmBlockStart, int(sb.SBlocksCount), region)
}

func (r *FileFsRepository) WriteBitmapInode(id string, data []byte) error {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	sb, err := readSuperBlock(f, region.Start)
	if err != nil {
		return err
	}

	_, err = SafeWriteAt(f, data, sb.SBmInodeStart, region)
	return err
}

func (r *FileFsRepository) WriteBitmapBlock(id string, data []byte) error {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	sb, err := readSuperBlock(f, region.Start)
	if err != nil {
		return err
	}

	_, err = SafeWriteAt(f, data, sb.SBmBlockStart, region)
	return err
}

func (r *FileFsRepository) ReadInode(id string, index int) (models.Inode, error) {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return models.Inode{}, err
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return models.Inode{}, err
	}
	defer f.Close()

	sb, err := readSuperBlock(f, region.Start)
	if err != nil {
		return models.Inode{}, err
	}

	return readInodeAt(f, sb.SInodeStart, index, region)
}

func (r *FileFsRepository) WriteInode(id string, index int, ino models.Inode) error {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	sb, err := readSuperBlock(f, region.Start)
	if err != nil {
		return err
	}

	return writeInodeAt(f, sb.SInodeStart, index, &ino, region)
}

func (r *FileFsRepository) ReadBlock(id string, index int) ([]byte, error) {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sb, err := readSuperBlock(f, region.Start)
	if err != nil {
		return nil, err
	}

	block := make([]byte, models.BlockSizeFile)
	offset := sb.SBlockStart + int64(index*models.BlockSizeFile)
	if _, err := f.Seek(offset, 0); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(f, block); err != nil {
		return nil, err
	}

	return block, nil
}

func (r *FileFsRepository) WriteBlock(id string, index int, data []byte) error {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	sb, err := readSuperBlock(f, region.Start)
	if err != nil {
		return err
	}

	if len(data) > models.BlockSizeFile {
		return fmt.Errorf("bloque demasiado grande: %d bytes (máx %d)", len(data), models.BlockSizeFile)
	}

	// Pad con zeros si es necesario
	block := make([]byte, models.BlockSizeFile)
	copy(block, data)

	offset := sb.SBlockStart + int64(index*models.BlockSizeFile)
	_, err = SafeWriteAt(f, block, offset, region)
	return err
}

// -------- Funciones auxiliares internas --------

// computeLayout calcula el layout del filesystem usando la fórmula del enunciado
func computeLayout(partStart, partSize int64, inodeSize, blockSize int) models.SuperBlock {
	// Fórmula: partSize = sizeof(SB) + n + 3n + n*inodeSize + 3n*blockSize
	// Resolviendo para n:
	// n = (partSize - sizeof(SB)) / (1 + 3 + inodeSize + 3*blockSize)

	const sbSize = 80 // tamaño aproximado del SuperBlock (ajustar si es necesario)

	numerator := float64(partSize - sbSize)
	denominator := float64(1 + 3 + inodeSize + 3*blockSize)

	n := int32(numerator / denominator)
	if n < 1 {
		n = 1
	}

	inodes := n
	blocks := 3 * n

	// Calcular offsets absolutos
	bmInodeSize := int64(n)
	bmBlockSize := int64(3 * n)
	inodeTableSize := int64(n) * int64(inodeSize)

	sbStart := partStart
	bmInodeStart := sbStart + sbSize
	bmBlockStart := bmInodeStart + bmInodeSize
	inodeStart := bmBlockStart + bmBlockSize
	blockStart := inodeStart + inodeTableSize

	return models.SuperBlock{
		SFilesystemType:  models.FSProjectType,
		SInodesCount:     inodes,
		SBlocksCount:     blocks,
		SFreeInodesCount: inodes,
		SFreeBlocksCount: blocks,
		SMtime:           0,
		SUmtime:          0,
		SMagic:           models.FSMagic,
		SBmInodeStart:    bmInodeStart,
		SBmBlockStart:    bmBlockStart,
		SInodeStart:      inodeStart,
		SBlockStart:      blockStart,
	}
}

// writeSuperBlock escribe el superblock al inicio de la partición
func writeSuperBlock(f *os.File, offset int64, sb *models.SuperBlock) error {
	if _, err := f.Seek(offset, 0); err != nil {
		return err
	}
	return binary.Write(f, binary.LittleEndian, sb)
}

// readSuperBlock lee el superblock desde el inicio de la partición
func readSuperBlock(f *os.File, offset int64) (models.SuperBlock, error) {
	var sb models.SuperBlock
	if _, err := f.Seek(offset, 0); err != nil {
		return sb, err
	}
	if err := binary.Read(f, binary.LittleEndian, &sb); err != nil {
		return sb, err
	}
	return sb, nil
}

// readSuperBlockUniversal lee el superblock y detecta si es EXT2 o EXT3
// Retorna: superblock (solo para magic check), inodeStart, blockStart, error
func (r *FileFsRepository) readSuperBlockUniversal(f *os.File, offset int64) (interface{}, int64, int64, error) {
	// Primero leer como EXT2 para detectar el tipo
	sb2, err := readSuperBlock(f, offset)
	if err != nil {
		return nil, 0, 0, err
	}

	// Si SMagic no es válido, probar como EXT3
	if sb2.SMagic != models.FSMagic {
		// Intentar leer como EXT3
		sb3, err := r.readSuperExt3(f, offset)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("formato inválido (ni EXT2 ni EXT3)")
		}
		// EXT3
		return sb3, sb3.InodeStart, sb3.BlockStart, nil
	}

	// Es EXT2
	return sb2, sb2.SInodeStart, sb2.SBlockStart, nil
}

// writeInodeAt escribe un inodo en la tabla de inodos
func writeInodeAt(f *os.File, inodeTableStart int64, index int, ino *models.Inode, region Region) error {
	offset := inodeTableStart + int64(index*binary.Size(*ino))
	fmt.Printf("DEBUG writeInodeAt: index=%d, offset=%d, IUid=%d, ISize=%d\n", index, offset, ino.IUid, ino.ISize)
	if _, err := f.Seek(offset, 0); err != nil {
		return fmt.Errorf("error seeking to %d: %w", offset, err)
	}
	if err := binary.Write(f, binary.LittleEndian, ino); err != nil {
		return fmt.Errorf("error writing inode: %w", err)
	}
	fmt.Printf("DEBUG writeInodeAt: inodo %d escrito exitosamente\n", index)
	return nil
}

// readInodeAt lee un inodo de la tabla de inodos
func readInodeAt(f *os.File, inodeTableStart int64, index int, region Region) (models.Inode, error) {
	var ino models.Inode
	inodeSize := binary.Size(ino)
	offset := inodeTableStart + int64(index*inodeSize)
	if index == 0 || index == 1 {
		fmt.Printf("DEBUG readInodeAt: index=%d, inodeSize=%d, offset=%d, inodeTableStart=%d\n", index, inodeSize, offset, inodeTableStart)
	}
	if _, err := f.Seek(offset, 0); err != nil {
		return ino, err
	}
	if err := binary.Read(f, binary.LittleEndian, &ino); err != nil {
		return ino, err
	}
	if index == 0 || index == 1 {
		fmt.Printf("DEBUG readInodeAt: inodo %d leído: IType=%d, ISize=%d, IUid=%d\n", index, ino.IType, ino.ISize, ino.IUid)
	}
	return ino, nil
}

// writeBlockAt escribe un bloque (folder, file o pointer)
func writeBlockAt(f *os.File, blockAreaStart int64, index int, data interface{}, region Region) error {
	var buf []byte
	switch v := data.(type) {
	case *models.FolderBlock:
		buf = make([]byte, binary.Size(*v))
		w := bytes.NewBuffer(buf[:0])
		if err := binary.Write(w, binary.LittleEndian, v); err != nil {
			return err
		}
		buf = w.Bytes()
	case *models.FileBlock:
		buf = v.Content[:]
	case *models.PointerBlock:
		buf = make([]byte, binary.Size(*v))
		w := bytes.NewBuffer(buf[:0])
		if err := binary.Write(w, binary.LittleEndian, v); err != nil {
			return err
		}
		buf = w.Bytes()
	default:
		return fmt.Errorf("tipo de bloque no soportado")
	}

	offset := blockAreaStart + int64(index*models.BlockSizeFile)
	if _, err := f.Seek(offset, 0); err != nil {
		return err
	}
	_, err := f.Write(buf)
	return err
}

// readBitmap lee un bitmap completo desde disco
func readBitmap(f *os.File, offset int64, size int, region Region) ([]byte, error) {
	bm := make([]byte, size)
	if _, err := SafeReadAt(f, bm, offset, region); err != nil {
		return nil, err
	}
	return bm, nil
}

// allocateInodeAndBlock busca el primer inodo y bloque libres en los bitmaps y los marca como usados
func allocateInodeAndBlock(bmInode, bmBlock []byte, maxInodes, maxBlocks int) (inodeIdx, blockIdx int, err error) {
	inodeIdx = -1
	blockIdx = -1

	// Buscar primer inodo libre
	for i := 0; i < maxInodes; i++ {
		if bmInode[i] == 0 {
			bmInode[i] = 1
			inodeIdx = i
			break
		}
	}
	if inodeIdx == -1 {
		return -1, -1, fmt.Errorf("no hay inodos libres")
	}

	// Buscar primer bloque libre
	for i := 0; i < maxBlocks; i++ {
		if bmBlock[i] == 0 {
			bmBlock[i] = 1
			blockIdx = i
			break
		}
	}
	if blockIdx == -1 {
		// Revertir asignación de inodo
		bmInode[inodeIdx] = 0
		return -1, -1, fmt.Errorf("no hay bloques libres")
	}

	return inodeIdx, blockIdx, nil
}

// findEntryInDir busca una entrada por nombre en un directorio, retorna el índice de inodo o -1 si no existe
func findEntryInDir(f *os.File, dirInode *models.Inode, blockAreaStart int64, name string, region Region) (int32, error) {
	// Solo soportar bloques directos por ahora (primeros 12 apuntadores)
	for i := 0; i < 12; i++ {
		blkIdx := dirInode.IBlock[i]
		if blkIdx == -1 {
			continue
		}

		// Leer el bloque de directorio
		var dirBlk models.FolderBlock
		offset := blockAreaStart + int64(blkIdx*models.BlockSizeFile)
		if _, err := f.Seek(offset, 0); err != nil {
			return -1, err
		}
		if err := binary.Read(f, binary.LittleEndian, &dirBlk); err != nil {
			return -1, err
		}

		// Buscar entrada
		for _, entry := range dirBlk.Content {
			if entry.BInodo == -1 {
				continue
			}
			entryName := strings.TrimRight(string(entry.BName[:]), "\x00")
			if entryName == name {
				return entry.BInodo, nil
			}
		}
	}

	return -1, nil
}

// addEntryToDir agrega una nueva entrada a un directorio, reservando bloques adicionales si es necesario
func addEntryToDir(f *os.File, dirInode *models.Inode, blockAreaStart int64, name string, inodeIdx int32, bmBlock []byte, maxBlocks int, region Region) error {
	// Buscar slot libre en bloques existentes
	for i := 0; i < 12; i++ {
		blkIdx := dirInode.IBlock[i]
		if blkIdx == -1 {
			// Necesitamos reservar un nuevo bloque
			newBlkIdx := -1
			for j := 0; j < maxBlocks; j++ {
				if bmBlock[j] == 0 {
					bmBlock[j] = 1
					newBlkIdx = j
					break
				}
			}
			if newBlkIdx == -1 {
				return fmt.Errorf("no hay bloques libres para expandir directorio")
			}

			// Crear nuevo bloque de directorio vacío
			var dirBlk models.FolderBlock
			for k := range dirBlk.Content {
				dirBlk.Content[k].BInodo = -1
			}
			// Agregar entrada en el primer slot
			copy(dirBlk.Content[0].BName[:], name)
			dirBlk.Content[0].BInodo = inodeIdx

			// Escribir bloque
			if err := writeBlockAt(f, blockAreaStart, newBlkIdx, &dirBlk, region); err != nil {
				return err
			}

			// Actualizar inodo del directorio
			dirInode.IBlock[i] = int32(newBlkIdx)
			return nil
		}

		// Leer bloque existente
		var dirBlk models.FolderBlock
		offset := blockAreaStart + int64(blkIdx*models.BlockSizeFile)
		if _, err := f.Seek(offset, 0); err != nil {
			return err
		}
		if err := binary.Read(f, binary.LittleEndian, &dirBlk); err != nil {
			return err
		}

		// Buscar slot libre
		for j := range dirBlk.Content {
			if dirBlk.Content[j].BInodo == -1 {
				copy(dirBlk.Content[j].BName[:], name)
				dirBlk.Content[j].BInodo = inodeIdx

				// Escribir bloque actualizado
				if err := writeBlockAt(f, blockAreaStart, int(blkIdx), &dirBlk, region); err != nil {
					return err
				}
				return nil
			}
		}
	}

	return fmt.Errorf("directorio lleno (no hay slots libres)")
}

// hasReadPermission verifica si el usuario tiene permiso de lectura sobre un archivo
// basándose en los permisos Unix (owner-group-others)
func hasReadPermission(fileInode models.Inode, uid int32, gid int32) bool {
	// Los permisos están en formato UGO: IPerm[0] = owner, IPerm[1] = group, IPerm[2] = others
	// Cada byte es un carácter ASCII ('0'-'7')

	// Root siempre tiene todos los permisos (bypass)
	if uid == 1 {
		return true
	}

	// Si el usuario es el dueño del archivo
	if fileInode.IUid == uid {
		// Verificar bit de lectura del owner (bit 2: 4 = read)
		ownerPerm := int(fileInode.IPerm[0] - '0')
		return (ownerPerm & 4) != 0 // bit de lectura en owner
	}

	// Si el usuario pertenece al grupo del archivo
	if fileInode.IGid == gid {
		// Verificar bit de lectura del group (bit 2: 4 = read)
		groupPerm := int(fileInode.IPerm[1] - '0')
		return (groupPerm & 4) != 0 // bit de lectura en group
	}

	// Otherwise, verificar permisos de others
	othersPerm := int(fileInode.IPerm[2] - '0')
	return (othersPerm & 4) != 0 // bit de lectura en others
}

func hasWritePermission(dirInode models.Inode, uid int32, gid int32) bool {
	// Los permisos están en formato UGO: IPerm[0] = owner, IPerm[1] = group, IPerm[2] = others
	// Cada byte es un carácter ASCII ('0'-'7')

	// Root siempre tiene todos los permisos (bypass)
	if uid == 1 {
		return true
	}

	// Si el usuario es el dueño del directorio
	if dirInode.IUid == uid {
		// Verificar bit de escritura del owner (bit 1: 2 = write)
		ownerPerm := int(dirInode.IPerm[0] - '0')
		return (ownerPerm & 2) != 0 // bit de escritura en owner
	}

	// Si el usuario pertenece al grupo del directorio
	if dirInode.IGid == gid {
		// Verificar bit de escritura del group (bit 1: 2 = write)
		groupPerm := int(dirInode.IPerm[1] - '0')
		return (groupPerm & 2) != 0 // bit de escritura en group
	}

	// Otherwise, verificar permisos de others
	othersPerm := int(dirInode.IPerm[2] - '0')
	return (othersPerm & 2) != 0 // bit de escritura en others
}

// ============================================================
// EXT3 + JOURNAL - Métodos públicos del repositorio
// ============================================================

// MkfsWithTypeExt3Wrapper formatea partición con EXT3 (wrapper público)
func (r *FileFsRepository) MkfsExt3(id string) error {
	return r.MkfsWithTypeExt3(id, JournalConst50)
}

// MkfsWithType formatea la partición con EXT2 o EXT3 (unificado P1/P2)
func (r *FileFsRepository) MkfsWithType(id string, fsType int32, full bool) error {
	switch fsType {
	case 2:
		// EXT2: usar la implementación existente de P1
		return r.Mkfs(id, "full")
	case 3:
		// EXT3: usar la implementación nueva con journal
		return r.MkfsWithTypeExt3(id, JournalConst50)
	default:
		return fmt.Errorf("fsType inválido (2|3)")
	}
}

// JournalAppendPublic agrega una entrada al journal (interfaz pública)
func (r *FileFsRepository) JournalAppendPublic(id string, op, path, content string, ts int64) error {
	diskPath, partStart, _, err := r.getMountInfo(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	sb, err := r.readSuperExt3(f, partStart)
	if err != nil {
		return err
	}

	if sb.FsType != 3 {
		return errors.New("operación no soportada en EXT2")
	}

	var inf models.Information
	inf.Set(op, path, content, ts)
	j := models.Journal{Count: 0, Content: inf}

	return r.AppendJournal(f, sb, j)
}

// JournalListPublic devuelve todas las entradas del journal
func (r *FileFsRepository) JournalListPublic(id string) ([]models.Journal, error) {
	diskPath, partStart, _, err := r.getMountInfo(id)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sb, err := r.readSuperExt3(f, partStart)
	if err != nil {
		return nil, err
	}

	if sb.FsType != 3 {
		return nil, errors.New("operación no soportada en EXT2")
	}

	return r.ListJournal(f, sb)
}

// JournalClearPublic limpia todas las entradas del journal
func (r *FileFsRepository) JournalClearPublic(id string) error {
	diskPath, partStart, _, err := r.getMountInfo(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	sb, err := r.readSuperExt3(f, partStart)
	if err != nil {
		return err
	}

	if sb.FsType != 3 {
		return errors.New("operación no soportada en EXT2")
	}

	return r.ClearJournal(f, sb)
}
