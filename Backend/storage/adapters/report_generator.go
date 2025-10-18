package adapters

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Backend/command/reports"
	"Backend/core/models"
	"Backend/core/ports"
	"Backend/storage/diskio"
	"Backend/storage/graphviz"
)

// ReportGenerator implementa reports.ReportGenerator
type ReportGenerator struct {
	gv         *graphviz.Dot
	fsRepo     *diskio.FileFsRepository
	diskRepo   *diskio.FileDiskRepository
	mountStore ports.MountStore
}

func NewReportGenerator(
	gv *graphviz.Dot,
	fsRepo *diskio.FileFsRepository,
	diskRepo *diskio.FileDiskRepository,
	mountStore ports.MountStore,
) reports.ReportGenerator {
	return &ReportGenerator{
		gv:         gv,
		fsRepo:     fsRepo,
		diskRepo:   diskRepo,
		mountStore: mountStore,
	}
}

func (g *ReportGenerator) ensureDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0o755)
}

func (g *ReportGenerator) MBR(id, out string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Buscar la partición montada
	var diskPath string
	for _, mount := range g.mountStore.List() {
		if mount.ID == id {
			diskPath = mount.Path
			break
		}
	}
	if diskPath == "" {
		return "", fmt.Errorf("id no montado: %s", id)
	}

	// Leer MBR
	mbr, err := g.diskRepo.ReadMBR(diskPath)
	if err != nil {
		return "", fmt.Errorf("error leyendo MBR: %w", err)
	}

	// Convertir timestamp a fecha legible
	timestamp := time.Unix(mbr.Timestamp, 0)
	fechaFormateada := timestamp.Format("02/01/2006 15:04:05")

	// Generar DOT para Graphviz con formato HTML table
	var dotContent strings.Builder
	dotContent.WriteString("digraph {\n")
	dotContent.WriteString("  node [shape=plaintext]\n\n")
	dotContent.WriteString("  TablaReportNodo [\n")
	dotContent.WriteString("    label=<\n")
	dotContent.WriteString("      <table border=\"1\" cellborder=\"1\" cellspacing=\"0\">\n")
	dotContent.WriteString("        <tr>\n")
	dotContent.WriteString("          <td bgcolor=\"SlateBlue\" colspan=\"2\"><b>Reporte MBR</b></td>\n")
	dotContent.WriteString("        </tr>\n")

	// Información del MBR
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"Azure\">mbr_tamano</td>\n          <td bgcolor=\"Azure\">%d</td>\n        </tr>\n", mbr.Size))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#AFA1D1\">mbr_fecha_creacion</td>\n          <td bgcolor=\"#AFA1D1\">%s</td>\n        </tr>\n", fechaFormateada))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"Azure\">mbr_disk_signature</td>\n          <td bgcolor=\"Azure\">%d</td>\n        </tr>\n", mbr.Signature))

	// Abrir archivo del disco para leer EBRs si es necesario
	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("error abriendo disco: %w", err)
	}
	defer f.Close()

	// Procesar particiones primarias y extendidas
	partitionCount := 0
	logicalPartitionCount := 0

	for i, part := range mbr.Partitions {
		if part.Status == 0 || part.Status != models.PartStatusUsed {
			continue
		}

		partitionCount++

		// Limpiar nombre
		name := string(part.Name[:])
		for j, c := range name {
			if c == 0 {
				name = name[:j]
				break
			}
		}

		// Calcular porcentaje del disco
		porcentaje := float64(part.Size) / float64(mbr.Size) * 100.0

		// Si es partición extendida, procesar primero sus lógicas
		if part.Type == models.PartTypeExtend {
			// Primero el header de la partición extendida
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightCoral\" colspan=\"2\"><b>Partición %d</b></td>\n        </tr>\n", i+1))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"MistyRose\">part_status</td>\n          <td bgcolor=\"MistyRose\">Activo</td>\n        </tr>\n"))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#FFB6C1\">part_type</td>\n          <td bgcolor=\"#FFB6C1\">%c</td>\n        </tr>\n", part.Type))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"MistyRose\">part_fit</td>\n          <td bgcolor=\"MistyRose\">%c</td>\n        </tr>\n", part.Fit))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#FFB6C1\">part_start</td>\n          <td bgcolor=\"#FFB6C1\">%d</td>\n        </tr>\n", part.Start))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"MistyRose\">part_size</td>\n          <td bgcolor=\"MistyRose\">%d</td>\n        </tr>\n", part.Size))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#FFB6C1\">part_name</td>\n          <td bgcolor=\"#FFB6C1\">%s</td>\n        </tr>\n", name))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"Azure\">porcentaje_disco</td>\n          <td bgcolor=\"Azure\">%.2f%%</td>\n        </tr>\n", porcentaje))

			// Leer y procesar EBRs (particiones lógicas)
			currentPos := part.Start
			localLogicalCount := 0

			for currentPos != -1 {
				var ebr models.EBR
				if _, err := f.Seek(currentPos, 0); err != nil {
					break
				}
				if err := binary.Read(f, binary.LittleEndian, &ebr); err != nil {
					break
				}

				// Verificar si es un EBR válido
				if ebr.Status == 0 && ebr.Size == 0 {
					break
				}

				if ebr.Status == models.PartStatusUsed && ebr.Size > 0 {
					localLogicalCount++
					logicalPartitionCount++

					ebrName := string(ebr.Name[:])
					for j, c := range ebrName {
						if c == 0 {
							ebrName = ebrName[:j]
							break
						}
					}

					ebrPorcentaje := float64(ebr.Size) / float64(mbr.Size) * 100.0

					dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightBlue\" colspan=\"2\"><b>Partición Lógica #%d</b></td>\n        </tr>\n", localLogicalCount+1))
					dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightCyan\">part_status</td>\n          <td bgcolor=\"LightCyan\">Activo</td>\n        </tr>\n"))
					dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#B0E0E6\">part_type</td>\n          <td bgcolor=\"#B0E0E6\">L</td>\n        </tr>\n"))
					dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightCyan\">part_start</td>\n          <td bgcolor=\"LightCyan\">%d</td>\n        </tr>\n", ebr.Start))
					dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#B0E0E6\">part_size</td>\n          <td bgcolor=\"#B0E0E6\">%d</td>\n        </tr>\n", ebr.Size))
					dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightCyan\">part_name</td>\n          <td bgcolor=\"LightCyan\">%s</td>\n        </tr>\n", ebrName))
					dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"Azure\">porcentaje_disco</td>\n          <td bgcolor=\"Azure\">%.2f%%</td>\n        </tr>\n", ebrPorcentaje))
				}

				currentPos = ebr.Next
			}
		} else {
			// Partición primaria
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightCoral\" colspan=\"2\"><b>Partición %d</b></td>\n        </tr>\n", i+1))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"MistyRose\">part_status</td>\n          <td bgcolor=\"MistyRose\">Activo</td>\n        </tr>\n"))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#FFB6C1\">part_type</td>\n          <td bgcolor=\"#FFB6C1\">%c</td>\n        </tr>\n", part.Type))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"MistyRose\">part_fit</td>\n          <td bgcolor=\"MistyRose\">%c</td>\n        </tr>\n", part.Fit))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#FFB6C1\">part_start</td>\n          <td bgcolor=\"#FFB6C1\">%d</td>\n        </tr>\n", part.Start))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"MistyRose\">part_size</td>\n          <td bgcolor=\"MistyRose\">%d</td>\n        </tr>\n", part.Size))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#FFB6C1\">part_name</td>\n          <td bgcolor=\"#FFB6C1\">%s</td>\n        </tr>\n", name))
			dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"Azure\">porcentaje_disco</td>\n          <td bgcolor=\"Azure\">%.2f%%</td>\n        </tr>\n", porcentaje))
		}
	}

	// Agregar particiones libres
	for i := partitionCount; i < 4; i++ {
		dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightGray\" colspan=\"2\">Partición %d - Libre</td>\n        </tr>\n", i+1))
	}

	dotContent.WriteString("      </table>\n")
	dotContent.WriteString("    >\n")
	dotContent.WriteString("  ]\n")
	dotContent.WriteString("}")

	// Generar imagen usando Graphviz
	return g.gv.Generate(dotContent.String(), out)
}

func (g *ReportGenerator) Disk(id, out string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Buscar la partición montada
	var diskPath string
	var diskName string
	for _, mount := range g.mountStore.List() {
		if mount.ID == id {
			diskPath = mount.Path
			// Extraer nombre del disco
			parts := strings.Split(diskPath, "/")
			diskName = strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
			break
		}
	}
	if diskPath == "" {
		return "", fmt.Errorf("id no montado: %s", id)
	}

	// Leer MBR
	mbr, err := g.diskRepo.ReadMBR(diskPath)
	if err != nil {
		return "", fmt.Errorf("error leyendo MBR: %w", err)
	}

	// Abrir archivo del disco para leer EBRs si es necesario
	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("error abriendo disco: %w", err)
	}
	defer f.Close()

	// Generar DOT para Graphviz con visualización de bloques
	var dotContent strings.Builder
	dotContent.WriteString("digraph G {\n")
	dotContent.WriteString("  rankdir=LR;\n")
	dotContent.WriteString("  node [shape=none];\n")
	dotContent.WriteString("  labelloc=\"t\";\n")
	dotContent.WriteString(fmt.Sprintf("  label=\"Reporte de Disco: %s\";\n", diskName))
	dotContent.WriteString("  diskStructure [label=<\n")
	dotContent.WriteString("    <table border=\"0\" cellborder=\"1\" cellspacing=\"0\" width=\"1000\">\n")
	dotContent.WriteString("      <tr>\n")

	// Calcular tamaño del MBR (típicamente los primeros bytes)
	sizeMBR := int64(binary.Size(mbr))
	porcentajeMBR := float64(sizeMBR) / float64(mbr.Size) * 100
	anchoMBR := int(porcentajeMBR * 10)
	if anchoMBR < 1 {
		anchoMBR = 1
	}
	dotContent.WriteString(fmt.Sprintf("        <td bgcolor=\"#87CEFA\" width=\"%d\">MBR<br/></td>\n", anchoMBR))

	// Estructura para ordenar particiones por posición
	type PartInfo struct {
		Index int
		Start int64
		Size  int64
	}
	var particiones []PartInfo
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Status == models.PartStatusUsed && mbr.Partitions[i].Size > 0 {
			particiones = append(particiones, PartInfo{
				Index: i,
				Start: mbr.Partitions[i].Start,
				Size:  mbr.Partitions[i].Size,
			})
		}
	}

	// Ordenar por posición de inicio
	for i := 0; i < len(particiones); i++ {
		for j := i + 1; j < len(particiones); j++ {
			if particiones[j].Start < particiones[i].Start {
				particiones[i], particiones[j] = particiones[j], particiones[i]
			}
		}
	}

	// Posición actual en el disco
	posActual := sizeMBR

	// Procesar cada partición en orden
	for _, partInfo := range particiones {
		part := mbr.Partitions[partInfo.Index]

		// Si hay espacio libre antes de esta partición
		if part.Start > posActual {
			espacioLibre := part.Start - posActual
			porcentajeLibre := float64(espacioLibre) / float64(mbr.Size) * 100
			anchoLibre := int(porcentajeLibre * 10)
			if anchoLibre < 1 {
				anchoLibre = 1
			}
			dotContent.WriteString(fmt.Sprintf("        <td bgcolor=\"#D3D3D3\" width=\"%d\">Libre<br/>%.2f%% del disco</td>\n", anchoLibre, porcentajeLibre))
		}

		// Limpiar nombre
		name := string(part.Name[:])
		for j, c := range name {
			if c == 0 {
				name = name[:j]
				break
			}
		}

		porcentajePart := float64(part.Size) / float64(mbr.Size) * 100
		anchoParticion := int(porcentajePart * 10)
		if anchoParticion < 1 {
			anchoParticion = 1
		}

		if part.Type == models.PartTypeExtend {
			// Partición extendida con tabla interna para EBRs
			dotContent.WriteString(fmt.Sprintf("        <td width=\"%d\">\n", anchoParticion))
			dotContent.WriteString("          <table border=\"0\" cellborder=\"1\" cellspacing=\"0\" width=\"100%\">\n")
			dotContent.WriteString("            <tr>\n")
			dotContent.WriteString(fmt.Sprintf("              <td bgcolor=\"#FFFACD\" colspan=\"10\">Extendida<br/>%s<br/>%.2f%% del disco</td>\n", name, porcentajePart))
			dotContent.WriteString("            </tr>\n")
			dotContent.WriteString("            <tr>\n")

			// Procesar particiones lógicas dentro de la extendida
			g.processLogicalPartitions(f, part, mbr.Size, &dotContent)

			dotContent.WriteString("            </tr>\n")
			dotContent.WriteString("          </table>\n")
			dotContent.WriteString("        </td>\n")
		} else {
			// Partición primaria
			dotContent.WriteString(fmt.Sprintf("        <td bgcolor=\"#98FB98\" width=\"%d\">Primaria<br/>%s<br/>%.2f%% del disco</td>\n", anchoParticion, name, porcentajePart))
		}

		posActual = part.Start + part.Size
	}

	// Espacio libre al final
	if posActual < mbr.Size {
		espacioLibre := mbr.Size - posActual
		porcentajeLibre := float64(espacioLibre) / float64(mbr.Size) * 100
		anchoLibre := int(porcentajeLibre * 10)
		if anchoLibre < 1 {
			anchoLibre = 1
		}
		dotContent.WriteString(fmt.Sprintf("        <td bgcolor=\"#D3D3D3\" width=\"%d\">Libre<br/>%.2f%% del disco</td>\n", anchoLibre, porcentajeLibre))
	}

	dotContent.WriteString("      </tr>\n")
	dotContent.WriteString("    </table>\n")
	dotContent.WriteString("  >];\n")
	dotContent.WriteString("}\n")

	// Generar imagen usando Graphviz
	return g.gv.Generate(dotContent.String(), out)
}

// processLogicalPartitions procesa las particiones lógicas dentro de una extendida
func (g *ReportGenerator) processLogicalPartitions(f *os.File, extPart models.Partition, diskSize int64, dotContent *strings.Builder) {
	currentPos := extPart.Start
	posActual := extPart.Start

	for currentPos != -1 {
		var ebr models.EBR
		if _, err := f.Seek(currentPos, 0); err != nil {
			break
		}
		if err := binary.Read(f, binary.LittleEndian, &ebr); err != nil {
			break
		}

		// Verificar si es un EBR válido
		if ebr.Status == 0 && ebr.Size == 0 {
			break
		}

		if ebr.Status == models.PartStatusUsed && ebr.Size > 0 {
			// Espacio libre antes del EBR (si existe)
			if currentPos > posActual && posActual != extPart.Start {
				espacioLibre := currentPos - posActual
				porcentajeLibre := float64(espacioLibre) / float64(diskSize) * 100
				if porcentajeLibre > 0.01 {
					anchoLibre := int(porcentajeLibre * 10)
					if anchoLibre < 1 {
						anchoLibre = 1
					}
					dotContent.WriteString(fmt.Sprintf("              <td bgcolor=\"#E6E6FA\" width=\"%d\">Libre<br/>%.2f%%</td>\n", anchoLibre, porcentajeLibre))
				}
			}

			// EBR header
			porcentajeEBR := 0.5 // Pequeño espacio fijo para el EBR
			anchoEBR := int(porcentajeEBR * 10)
			if anchoEBR < 5 {
				anchoEBR = 5
			}
			dotContent.WriteString(fmt.Sprintf("              <td bgcolor=\"#B0C4DE\" width=\"%d\">EBR</td>\n", anchoEBR))

			// Partición lógica
			ebrName := string(ebr.Name[:])
			for j, c := range ebrName {
				if c == 0 {
					ebrName = ebrName[:j]
					break
				}
			}
			porcentajeLogica := float64(ebr.Size) / float64(diskSize) * 100
			anchoLogica := int(porcentajeLogica * 10)
			if anchoLogica < 1 {
				anchoLogica = 1
			}
			dotContent.WriteString(fmt.Sprintf("              <td bgcolor=\"#ADD8E6\" width=\"%d\">Lógica<br/>%s<br/>%.2f%%</td>\n", anchoLogica, ebrName, porcentajeLogica))

			posActual = ebr.Start + ebr.Size
		}

		currentPos = ebr.Next
	}

	// Espacio libre al final de la partición extendida
	finExtendida := extPart.Start + extPart.Size
	if posActual < finExtendida {
		espacioLibre := finExtendida - posActual
		porcentajeLibre := float64(espacioLibre) / float64(diskSize) * 100
		if porcentajeLibre > 0.01 {
			anchoLibre := int(porcentajeLibre * 10)
			if anchoLibre < 1 {
				anchoLibre = 1
			}
			dotContent.WriteString(fmt.Sprintf("              <td bgcolor=\"#E6E6FA\" width=\"%d\">Libre<br/>%.2f%%</td>\n", anchoLibre, porcentajeLibre))
		}
	}
}

func (g *ReportGenerator) Inode(id, out string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Leer SuperBlock
	sb, _, err := g.fsRepo.ReadSuper(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %w", err)
	}

	// Buscar disco path para acceso directo
	var diskPath string
	for _, mount := range g.mountStore.List() {
		if mount.ID == id {
			diskPath = mount.Path
			break
		}
	}
	if diskPath == "" {
		return "", fmt.Errorf("id no montado: %s", id)
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Leer bitmap de inodos
	bmInode, err := g.fsRepo.ReadBitmapInode(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap inodos: %w", err)
	}

	// Crear contenido DOT para Graphviz
	var dotContent strings.Builder
	dotContent.WriteString("digraph {\n")
	dotContent.WriteString("  node [shape=plain]\n")
	dotContent.WriteString("  rankdir=TB;\n")

	// Título del reporte
	dotContent.WriteString("  titulo [label=<<table border=\"1\">\n")
	dotContent.WriteString("    <tr>\n")
	dotContent.WriteString("      <td bgcolor=\"SlateBlue\" COLSPAN=\"2\"> Reporte de Inodos </td>\n")
	dotContent.WriteString("    </tr>\n")
	dotContent.WriteString("  </table>>]\n\n")

	// Mapa para rastrear bloques mostrados
	bloquesMostrados := make(map[int32]bool)

	// Estructura para rastrear relaciones entre inodos
	type InodeRelation struct {
		FromInode int32
		ToInode   int32
		Name      string
	}
	var relaciones []InodeRelation

	// Primera pasada: crear nodos de inodos y recopilar relaciones
	for i := int32(0); i < sb.SInodesCount; i++ {
		// Verificar si el inodo está en uso
		if i >= int32(len(bmInode)) || bmInode[i] == 0 {
			continue
		}

		inode, err := g.fsRepo.ReadInode(id, int(i))
		if err != nil || inode.IUid == -1 {
			continue
		}

		// Determinar tipo y color
		tipoInodo := "Carpeta"
		colorInodo := "#90EE90" // Verde claro para carpetas
		if inode.IType == models.FileTypeRegular {
			tipoInodo = "Archivo"
			colorInodo = "#87CEEB" // Azul claro para archivos
		}

		// Formatear fechas
		atimeStr := time.Unix(inode.IAtime, 0).Format("02/01/2006 15:04")
		ctimeStr := time.Unix(inode.ICtime, 0).Format("02/01/2006 15:04")
		mtimeStr := time.Unix(inode.IMtime, 0).Format("02/01/2006 15:04")

		// Limpiar permisos
		permStr := strings.TrimRight(string(inode.IPerm[:]), "\x00")
		if permStr == "" {
			if tipoInodo == "Carpeta" {
				permStr = "755"
			} else {
				permStr = "664"
			}
		}

		// Crear nodo del inodo
		dotContent.WriteString(fmt.Sprintf("  inode%d [label=<<table border=\"1\">\n", i))
		dotContent.WriteString(fmt.Sprintf("    <tr><td bgcolor=\"%s\" COLSPAN=\"2\">Inodo %d (%s)</td></tr>\n", colorInodo, i, tipoInodo))
		dotContent.WriteString(fmt.Sprintf("    <tr><td bgcolor=\"#E6E6FA\">i_uid</td><td bgcolor=\"#E6E6FA\">%d</td></tr>\n", inode.IUid))
		dotContent.WriteString(fmt.Sprintf("    <tr><td bgcolor=\"#F0F8FF\">i_gid</td><td bgcolor=\"#F0F8FF\">%d</td></tr>\n", inode.IGid))
		dotContent.WriteString(fmt.Sprintf("    <tr><td bgcolor=\"#E6E6FA\">i_size</td><td bgcolor=\"#E6E6FA\">%d</td></tr>\n", inode.ISize))
		dotContent.WriteString(fmt.Sprintf("    <tr><td bgcolor=\"#F0F8FF\">i_atime</td><td bgcolor=\"#F0F8FF\">%s</td></tr>\n", atimeStr))
		dotContent.WriteString(fmt.Sprintf("    <tr><td bgcolor=\"#E6E6FA\">i_ctime</td><td bgcolor=\"#E6E6FA\">%s</td></tr>\n", ctimeStr))
		dotContent.WriteString(fmt.Sprintf("    <tr><td bgcolor=\"#F0F8FF\">i_mtime</td><td bgcolor=\"#F0F8FF\">%s</td></tr>\n", mtimeStr))
		dotContent.WriteString(fmt.Sprintf("    <tr><td bgcolor=\"#E6E6FA\">i_type</td><td bgcolor=\"#E6E6FA\">%s</td></tr>\n", tipoInodo))
		dotContent.WriteString(fmt.Sprintf("    <tr><td bgcolor=\"#F0F8FF\">i_perm</td><td bgcolor=\"#F0F8FF\">%s</td></tr>\n", permStr))
		dotContent.WriteString(fmt.Sprintf("  </table>>]\n\n"))

		// Si es un directorio, leer sus bloques para encontrar relaciones con otros inodos
		if inode.IType == models.FileTypeFolder {
			for j := 0; j < 12; j++ {
				blkIdx := inode.IBlock[j]
				if blkIdx == -1 {
					break
				}

				// Leer bloque de directorio
				var dirBlk models.FolderBlock
				blkOffset := sb.SBlockStart + int64(blkIdx)*64
				if _, err := f.Seek(blkOffset, 0); err != nil {
					continue
				}
				if err := binary.Read(f, binary.LittleEndian, &dirBlk); err != nil {
					continue
				}

				// Revisar entradas del directorio
				for _, entry := range dirBlk.Content {
					if entry.BInodo == -1 {
						continue
					}

					entryName := strings.TrimRight(string(entry.BName[:]), "\x00")
					// Evitar . y .. para no crear ciclos confusos
					if entryName == "." || entryName == ".." {
						continue
					}

					// Agregar relación
					relaciones = append(relaciones, InodeRelation{
						FromInode: i,
						ToInode:   entry.BInodo,
						Name:      entryName,
					})
				}
			}
		}
	}

	// Segunda pasada: crear nodos de bloques y conexiones inodo->bloque
	for i := int32(0); i < sb.SInodesCount; i++ {
		if i >= int32(len(bmInode)) || bmInode[i] == 0 {
			continue
		}

		inode, err := g.fsRepo.ReadInode(id, int(i))
		if err != nil || inode.IUid == -1 {
			continue
		}

		// Procesar bloques directos
		for j := 0; j < 12; j++ {
			if inode.IBlock[j] != -1 {
				bloqueId := inode.IBlock[j]

				// Crear nodo del bloque si no existe
				if !bloquesMostrados[bloqueId] {
					dotContent.WriteString(fmt.Sprintf("  bloque%d [label=\"Bloque %d\", shape=box, style=filled, fillcolor=\"#FFE4B5\"]\n", bloqueId, bloqueId))
					bloquesMostrados[bloqueId] = true
				}

				// Crear flecha del inodo al bloque
				dotContent.WriteString(fmt.Sprintf("  inode%d -> bloque%d [label=\"bloque[%d]\"]\n", i, bloqueId, j))
			}
		}

		// Procesar bloques indirectos
		if inode.IBlock[12] != -1 {
			bloqueId := inode.IBlock[12]
			if !bloquesMostrados[bloqueId] {
				dotContent.WriteString(fmt.Sprintf("  bloque%d [label=\"Bloque %d\\n(Indirecto Simple)\", shape=box, style=filled, fillcolor=\"#DDA0DD\"]\n", bloqueId, bloqueId))
				bloquesMostrados[bloqueId] = true
			}
			dotContent.WriteString(fmt.Sprintf("  inode%d -> bloque%d [label=\"indirecto\", style=dashed]\n", i, bloqueId))
		}

		if inode.IBlock[13] != -1 {
			bloqueId := inode.IBlock[13]
			if !bloquesMostrados[bloqueId] {
				dotContent.WriteString(fmt.Sprintf("  bloque%d [label=\"Bloque %d\\n(Indirecto Doble)\", shape=box, style=filled, fillcolor=\"#DDA0DD\"]\n", bloqueId, bloqueId))
				bloquesMostrados[bloqueId] = true
			}
			dotContent.WriteString(fmt.Sprintf("  inode%d -> bloque%d [label=\"doble_indirecto\", style=dashed]\n", i, bloqueId))
		}

		if inode.IBlock[14] != -1 {
			bloqueId := inode.IBlock[14]
			if !bloquesMostrados[bloqueId] {
				dotContent.WriteString(fmt.Sprintf("  bloque%d [label=\"Bloque %d\\n(Indirecto Triple)\", shape=box, style=filled, fillcolor=\"#DDA0DD\"]\n", bloqueId, bloqueId))
				bloquesMostrados[bloqueId] = true
			}
			dotContent.WriteString(fmt.Sprintf("  inode%d -> bloque%d [label=\"triple_indirecto\", style=dashed]\n", i, bloqueId))
		}
	}

	// Tercera pasada: agregar flechas entre inodos (relaciones de directorio)
	dotContent.WriteString("\n  // Relaciones entre inodos (directorios -> contenido)\n")
	for _, rel := range relaciones {
		dotContent.WriteString(fmt.Sprintf("  inode%d -> inode%d [label=\"%s\", color=\"blue\", style=\"bold\"]\n",
			rel.FromInode, rel.ToInode, rel.Name))
	}

	dotContent.WriteString("}\n")

	// Generar imagen usando Graphviz
	return g.gv.Generate(dotContent.String(), out)
}

func (g *ReportGenerator) Block(id, out string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Leer SuperBlock
	sb, _, err := g.fsRepo.ReadSuper(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %w", err)
	}

	// Buscar disco path para acceso directo
	var diskPath string
	for _, mount := range g.mountStore.List() {
		if mount.ID == id {
			diskPath = mount.Path
			break
		}
	}
	if diskPath == "" {
		return "", fmt.Errorf("id no montado: %s", id)
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Leer bitmap de bloques
	bmBlock, err := g.fsRepo.ReadBitmapBlock(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap bloques: %w", err)
	}

	// Leer bitmap de inodos
	bmInode, err := g.fsRepo.ReadBitmapInode(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap inodos: %w", err)
	}

	// Crear contenido DOT para Graphviz
	var dotContent strings.Builder
	dotContent.WriteString("digraph G {\n")
	dotContent.WriteString("  node [shape=plaintext];\n")
	dotContent.WriteString("  rankdir=TB;\n")
	dotContent.WriteString("  label=\"Reporte de Bloques\";\n\n")

	// Mapas para rastrear tipos de bloques y relaciones
	bloquesTipo := make(map[int32]string)
	conexionesCarpetas := make(map[int32][]int32)    // bloque_carpeta -> [inodos_referenciados]
	conexionesApuntadores := make(map[int32][]int32) // bloque_apuntador -> [bloques_apuntados]

	// Primera pasada: determinar tipos de bloques analizando inodos
	for i := int32(0); i < sb.SInodesCount; i++ {
		if i >= int32(len(bmInode)) || bmInode[i] == 0 {
			continue
		}

		inode, err := g.fsRepo.ReadInode(id, int(i))
		if err != nil || inode.IUid == -1 {
			continue
		}

		// Determinar tipo de bloque basado en tipo de inodo
		tipoBloque := "archivo"
		if inode.IType == models.FileTypeFolder {
			tipoBloque = "carpeta"
		}

		// Marcar bloques directos
		for j := 0; j < 12; j++ {
			if inode.IBlock[j] != -1 {
				bloquesTipo[inode.IBlock[j]] = tipoBloque
			}
		}

		// Marcar bloques indirectos como apuntadores
		if inode.IBlock[12] != -1 {
			bloquesTipo[inode.IBlock[12]] = "apuntador"
		}
		if inode.IBlock[13] != -1 {
			bloquesTipo[inode.IBlock[13]] = "apuntador"
		}
		if inode.IBlock[14] != -1 {
			bloquesTipo[inode.IBlock[14]] = "apuntador"
		}
	}

	// Segunda pasada: crear nodos de bloques y recopilar conexiones
	for i := int32(0); i < sb.SBlocksCount; i++ {
		// Verificar si el bloque está en uso
		if i >= int32(len(bmBlock)) || bmBlock[i] == 0 {
			continue
		}

		tipoBloque, existe := bloquesTipo[i]
		if !existe {
			tipoBloque = "desconocido"
		}

		// Generar nodo según el tipo de bloque
		if tipoBloque == "carpeta" {
			// Leer bloque de carpeta
			var dirBlk models.FolderBlock
			blkOffset := sb.SBlockStart + int64(i)*64
			if _, err := f.Seek(blkOffset, 0); err != nil {
				continue
			}
			if err := binary.Read(f, binary.LittleEndian, &dirBlk); err != nil {
				continue
			}

			// Crear tabla para bloque de carpeta
			dotContent.WriteString(fmt.Sprintf("  bloque%d [label=<\n", i))
			dotContent.WriteString("    <table border=\"1\" cellborder=\"1\" cellspacing=\"0\">\n")
			dotContent.WriteString(fmt.Sprintf("      <tr><td colspan=\"2\" bgcolor=\"#FFFFCC\">Bloque Carpeta %d</td></tr>\n", i))
			dotContent.WriteString("      <tr><td bgcolor=\"#F0E68C\">b_name</td><td bgcolor=\"#F0E68C\">b_inodo</td></tr>\n")

			// Recopilar referencias a inodos
			var inodosRef []int32

			// Añadir entradas
			for _, entry := range dirBlk.Content {
				if entry.BInodo == -1 {
					continue
				}

				entryName := strings.TrimRight(string(entry.BName[:]), "\x00")
				dotContent.WriteString(fmt.Sprintf("      <tr><td>%s</td><td>%d</td></tr>\n", entryName, entry.BInodo))

				// Agregar a referencias si no es . o ..
				if entryName != "." && entryName != ".." {
					inodosRef = append(inodosRef, entry.BInodo)
				}
			}

			conexionesCarpetas[i] = inodosRef
			dotContent.WriteString("    </table>\n")
			dotContent.WriteString("  >, style=filled, fillcolor=\"#FFFFCC\"];\n\n")

		} else if tipoBloque == "archivo" {
			// Leer bloque de archivo
			var fileBlk models.FileBlock
			blkOffset := sb.SBlockStart + int64(i)*64
			if _, err := f.Seek(blkOffset, 0); err != nil {
				continue
			}
			if err := binary.Read(f, binary.LittleEndian, &fileBlk); err != nil {
				continue
			}

			// Obtener contenido y limitar para visualización
			contenido := strings.TrimRight(string(fileBlk.Content[:]), "\x00")
			if len(contenido) > 30 {
				contenido = contenido[:30] + "..."
			}

			// Escapar caracteres especiales para HTML
			contenido = strings.ReplaceAll(contenido, "&", "&amp;")
			contenido = strings.ReplaceAll(contenido, "<", "&lt;")
			contenido = strings.ReplaceAll(contenido, ">", "&gt;")
			contenido = strings.ReplaceAll(contenido, "\"", "&quot;")
			contenido = strings.ReplaceAll(contenido, "\n", "&lt;br/&gt;")
			contenido = strings.ReplaceAll(contenido, "\r", "")

			// Crear tabla para bloque de archivo
			dotContent.WriteString(fmt.Sprintf("  bloque%d [label=<\n", i))
			dotContent.WriteString("    <table border=\"1\" cellborder=\"1\" cellspacing=\"0\">\n")
			dotContent.WriteString(fmt.Sprintf("      <tr><td bgcolor=\"#E6F5FF\">Bloque Archivo %d</td></tr>\n", i))
			dotContent.WriteString(fmt.Sprintf("      <tr><td>%s</td></tr>\n", contenido))
			dotContent.WriteString("    </table>\n")
			dotContent.WriteString("  >, style=filled, fillcolor=\"#E6F5FF\"];\n\n")

		} else if tipoBloque == "apuntador" {
			// Leer bloque de apuntadores
			var ptrBlk models.PointerBlock
			blkOffset := sb.SBlockStart + int64(i)*64
			if _, err := f.Seek(blkOffset, 0); err != nil {
				continue
			}
			if err := binary.Read(f, binary.LittleEndian, &ptrBlk); err != nil {
				continue
			}

			// Recopilar bloques apuntados
			var bloquesApuntados []int32

			// Crear tabla para bloque de apuntadores
			dotContent.WriteString(fmt.Sprintf("  bloque%d [label=<\n", i))
			dotContent.WriteString("    <table border=\"1\" cellborder=\"1\" cellspacing=\"0\">\n")
			dotContent.WriteString(fmt.Sprintf("      <tr><td bgcolor=\"#D8BFD8\">Bloque Apuntadores %d</td></tr>\n", i))
			dotContent.WriteString("      <tr><td>")

			// Formatear apuntadores
			var apuntadoresTexto []string
			for j := 0; j < len(ptrBlk.Pointers); j++ {
				apuntadoresTexto = append(apuntadoresTexto, fmt.Sprintf("%d", ptrBlk.Pointers[j]))
				if ptrBlk.Pointers[j] != -1 {
					bloquesApuntados = append(bloquesApuntados, ptrBlk.Pointers[j])
				}
			}

			conexionesApuntadores[i] = bloquesApuntados
			dotContent.WriteString(strings.Join(apuntadoresTexto, ", "))
			dotContent.WriteString("</td></tr>\n")
			dotContent.WriteString("    </table>\n")
			dotContent.WriteString("  >, style=filled, fillcolor=\"#D8BFD8\"];\n\n")

		} else {
			// Bloque desconocido
			dotContent.WriteString(fmt.Sprintf("  bloque%d [label=\"Bloque Desconocido %d\", style=filled, fillcolor=\"#DDDDDD\"];\n", i, i))
		}
	}

	// Tercera pasada: crear nodos de inodos referenciados y agregar conexiones
	inodosCreados := make(map[int32]bool)

	dotContent.WriteString("\n  // Conexiones de bloques de carpeta a inodos\n")
	for bloqueId, inodos := range conexionesCarpetas {
		for _, inodoId := range inodos {
			// Crear nodo del inodo si no existe
			if !inodosCreados[inodoId] {
				dotContent.WriteString(fmt.Sprintf("  inode%d [label=\"Inodo %d\", shape=ellipse, style=filled, fillcolor=\"#98FB98\"]\n", inodoId, inodoId))
				inodosCreados[inodoId] = true
			}
			// Crear flecha del bloque carpeta al inodo
			dotContent.WriteString(fmt.Sprintf("  bloque%d -> inode%d [label=\"ref\", color=\"blue\"]\n", bloqueId, inodoId))
		}
	}

	dotContent.WriteString("\n  // Conexiones de bloques apuntadores a otros bloques\n")
	for bloqueApuntador, bloquesApuntados := range conexionesApuntadores {
		for _, bloqueApuntado := range bloquesApuntados {
			// Solo crear flecha si el bloque apuntado está en uso
			if bloqueApuntado >= 0 && bloqueApuntado < int32(len(bmBlock)) && bmBlock[bloqueApuntado] != 0 {
				dotContent.WriteString(fmt.Sprintf("  bloque%d -> bloque%d [label=\"ptr\", color=\"red\", style=dashed]\n", bloqueApuntador, bloqueApuntado))
			}
		}
	}

	dotContent.WriteString("}\n")

	// Generar imagen usando Graphviz
	return g.gv.Generate(dotContent.String(), out)
}

func (g *ReportGenerator) BmInode(id, out string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Leer bitmap de inodos desde el filesystem
	bitmap, err := g.fsRepo.ReadBitmapInode(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de inodos: %w", err)
	}

	// Leer SuperBlock para obtener el total de inodos
	sb, _, err := g.fsRepo.ReadSuper(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %w", err)
	}

	// Crear contenido del reporte
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Bitmap de Inodos (ID: %s)\n", id))
	content.WriteString(fmt.Sprintf("Total de inodos: %d\n", sb.SInodesCount))
	content.WriteString(fmt.Sprintf("Inodos libres: %d\n", sb.SFreeInodesCount))
	content.WriteString(fmt.Sprintf("Inodos usados: %d\n\n", sb.SInodesCount-sb.SFreeInodesCount))

	// Convertir bytes a bits y mostrar 20 bits por línea
	bitCounter := 0
	totalInodos := int(sb.SInodesCount)

	for i := 0; i < totalInodos; i++ {
		// Calcular byte y bit dentro del byte
		byteIndex := i / 8
		bitIndex := uint(i % 8)

		// Si estamos al inicio de una línea, agregar número de inodo inicial
		if bitCounter%20 == 0 {
			if bitCounter > 0 {
				content.WriteString("\n")
			}
			content.WriteString(fmt.Sprintf("%4d: ", i))
		}

		// Leer el bit del bitmap
		var bitValue byte = 0
		if byteIndex < len(bitmap) {
			bitValue = (bitmap[byteIndex] >> bitIndex) & 1
		}

		// Escribir el bit
		content.WriteString(fmt.Sprintf("%d", bitValue))

		// Agregar espacio cada 5 bits para mejor legibilidad
		if (bitCounter+1)%5 == 0 && (bitCounter+1)%20 != 0 {
			content.WriteString(" ")
		}

		bitCounter++
	}

	content.WriteString("\n")

	// Escribir archivo
	if err := os.WriteFile(out, []byte(content.String()), 0o644); err != nil {
		return "", err
	}

	return out, nil
}

func (g *ReportGenerator) BmBlock(id, out string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Leer bitmap de bloques desde el filesystem
	bitmap, err := g.fsRepo.ReadBitmapBlock(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap de bloques: %w", err)
	}

	// Leer SuperBlock para obtener el total de bloques
	sb, _, err := g.fsRepo.ReadSuper(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %w", err)
	}

	// Crear contenido del reporte
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Bitmap de Bloques (ID: %s)\n", id))
	content.WriteString(fmt.Sprintf("Total de bloques: %d\n", sb.SBlocksCount))
	content.WriteString(fmt.Sprintf("Bloques libres: %d\n", sb.SFreeBlocksCount))
	content.WriteString(fmt.Sprintf("Bloques usados: %d\n\n", sb.SBlocksCount-sb.SFreeBlocksCount))

	// Convertir bytes a bits y mostrar 20 bits por línea
	bitCounter := 0
	totalBloques := int(sb.SBlocksCount)

	for i := 0; i < totalBloques; i++ {
		// Calcular byte y bit dentro del byte
		byteIndex := i / 8
		bitIndex := uint(i % 8)

		// Si estamos al inicio de una línea, agregar número de bloque inicial
		if bitCounter%20 == 0 {
			if bitCounter > 0 {
				content.WriteString("\n")
			}
			content.WriteString(fmt.Sprintf("%4d: ", i))
		}

		// Leer el bit del bitmap
		var bitValue byte = 0
		if byteIndex < len(bitmap) {
			bitValue = (bitmap[byteIndex] >> bitIndex) & 1
		}

		// Escribir el bit
		content.WriteString(fmt.Sprintf("%d", bitValue))

		// Agregar espacio cada 5 bits para mejor legibilidad
		if (bitCounter+1)%5 == 0 && (bitCounter+1)%20 != 0 {
			content.WriteString(" ")
		}

		bitCounter++
	}

	content.WriteString("\n")

	// Escribir archivo
	if err := os.WriteFile(out, []byte(content.String()), 0o644); err != nil {
		return "", err
	}

	return out, nil
}

func (g *ReportGenerator) Tree(id, out string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Leer SuperBlock
	sb, _, err := g.fsRepo.ReadSuper(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %w", err)
	}

	// Buscar disco path y nombre
	var diskPath, diskName string
	for _, mount := range g.mountStore.List() {
		if mount.ID == id {
			diskPath = mount.Path
			// Extraer nombre del disco
			parts := strings.Split(diskPath, "/")
			diskName = strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
			break
		}
	}
	if diskPath == "" {
		return "", fmt.Errorf("id no montado: %s", id)
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Leer bitmaps
	bmInode, err := g.fsRepo.ReadBitmapInode(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap inodos: %w", err)
	}

	// Estructuras para almacenar información
	inodosInfo := make(map[int32]models.Inode)
	bloquesCarpeta := make(map[int32]models.FolderBlock)
	bloquesArchivo := make(map[int32]models.FileBlock)
	tipoBloque := make(map[int32]string)

	// Recopilar todos los inodos en uso
	var inodosEnUso []int32
	for i := int32(0); i < sb.SInodesCount; i++ {
		if i >= int32(len(bmInode)) || bmInode[i] == 0 {
			continue
		}

		inode, err := g.fsRepo.ReadInode(id, int(i))
		if err != nil {
			continue
		}
		if inode.IUid == -1 {
			continue
		}

		inodosEnUso = append(inodosEnUso, i)
		inodosInfo[i] = inode
	}

	// Recopilar información de bloques
	for _, idInodo := range inodosEnUso {
		inodo := inodosInfo[idInodo]
		tipoInodo := "archivo"
		if inodo.IType == models.FileTypeFolder {
			tipoInodo = "carpeta"
		}

		// Procesar bloques directos
		for j := 0; j < 12; j++ {
			idBloque := inodo.IBlock[j]
			if idBloque == -1 {
				break
			}

			tipoBloque[idBloque] = tipoInodo
			blkOffset := sb.SBlockStart + int64(idBloque)*64

			if tipoInodo == "carpeta" {
				var dirBlk models.FolderBlock
				if _, err := f.Seek(blkOffset, 0); err == nil {
					if err := binary.Read(f, binary.LittleEndian, &dirBlk); err == nil {
						bloquesCarpeta[idBloque] = dirBlk
					}
				}
			} else {
				var fileBlk models.FileBlock
				if _, err := f.Seek(blkOffset, 0); err == nil {
					if err := binary.Read(f, binary.LittleEndian, &fileBlk); err == nil {
						bloquesArchivo[idBloque] = fileBlk
					}
				}
			}
		}
	}

	// Generar DOT
	var dotContent strings.Builder
	dotContent.WriteString("digraph G {\n")
	dotContent.WriteString("  node [shape=none fontname=\"Arial\"];\n")
	dotContent.WriteString("  edge [fontname=\"Arial\", fontsize=10];\n")
	dotContent.WriteString("  rankdir=TB;\n")
	dotContent.WriteString("  ranksep=0.6;\n")
	dotContent.WriteString("  nodesep=0.4;\n")
	dotContent.WriteString(fmt.Sprintf("  label=\"Árbol del Sistema de Archivos: %s\";\n", diskName))

	// Generar nodos de inodos
	for _, idInodo := range inodosEnUso {
		inodo := inodosInfo[idInodo]

		tipoInodo := "Carpeta"
		fillColor := "#FFFACD" // Amarillo para carpetas
		if inodo.IType == models.FileTypeRegular {
			tipoInodo = "Archivo"
			fillColor = "#E6F5FF" // Azul para archivos
		}

		dotContent.WriteString(fmt.Sprintf("  inodo%d [label=<\n", idInodo))
		dotContent.WriteString("    <table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n")
		dotContent.WriteString(fmt.Sprintf("      <tr><td colspan=\"2\" bgcolor=\"#4682B4\"><font color=\"white\">Inodo %d</font></td></tr>\n", idInodo))
		dotContent.WriteString(fmt.Sprintf("      <tr><td>ID</td><td>%d</td></tr>\n", idInodo))
		dotContent.WriteString(fmt.Sprintf("      <tr><td>UID</td><td>%d</td></tr>\n", inodo.IUid))
		dotContent.WriteString(fmt.Sprintf("      <tr><td>GID</td><td>%d</td></tr>\n", inodo.IGid))
		dotContent.WriteString(fmt.Sprintf("      <tr><td>Tipo</td><td>%s</td></tr>\n", tipoInodo))
		dotContent.WriteString(fmt.Sprintf("      <tr><td>Size</td><td>%d bytes</td></tr>\n", inodo.ISize))

		// Bloques directos (AD)
		dotContent.WriteString("      <tr><td>AD</td><td>")
		hasBlocks := false
		for j := 0; j < 12; j++ {
			if inodo.IBlock[j] != -1 {
				if hasBlocks {
					dotContent.WriteString(", ")
				}
				dotContent.WriteString(fmt.Sprintf("%d", inodo.IBlock[j]))
				hasBlocks = true
			}
		}
		dotContent.WriteString("</td></tr>\n")

		dotContent.WriteString("    </table>\n")
		dotContent.WriteString(fmt.Sprintf("  >, style=filled, fillcolor=\"%s\"];\n", fillColor))
	}

	// Generar nodos de bloques
	bloquesGenerados := make(map[int32]bool)

	for idBloque, tipo := range tipoBloque {
		if bloquesGenerados[idBloque] {
			continue
		}
		bloquesGenerados[idBloque] = true

		if tipo == "carpeta" {
			dirBlk := bloquesCarpeta[idBloque]
			dotContent.WriteString(fmt.Sprintf("  bloque%d [label=<\n", idBloque))
			dotContent.WriteString("    <table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n")
			dotContent.WriteString(fmt.Sprintf("      <tr><td colspan=\"2\" bgcolor=\"#F0E68C\">Bloque Carpeta %d</td></tr>\n", idBloque))

			for _, entry := range dirBlk.Content {
				if entry.BInodo == -1 {
					continue
				}
				entryName := strings.TrimRight(string(entry.BName[:]), "\x00")
				if entryName != "" {
					dotContent.WriteString(fmt.Sprintf("      <tr><td>%s</td><td>%d</td></tr>\n", entryName, entry.BInodo))
				}
			}

			dotContent.WriteString("    </table>\n")
			dotContent.WriteString("  >, style=filled, fillcolor=\"#FFFFCC\"];\n")

		} else if tipo == "archivo" {
			fileBlk := bloquesArchivo[idBloque]
			contenido := strings.TrimRight(string(fileBlk.Content[:]), "\x00")
			if len(contenido) > 20 {
				contenido = contenido[:20] + "..."
			}

			// Escapar caracteres especiales
			contenido = strings.ReplaceAll(contenido, "&", "&amp;")
			contenido = strings.ReplaceAll(contenido, "<", "&lt;")
			contenido = strings.ReplaceAll(contenido, ">", "&gt;")
			contenido = strings.ReplaceAll(contenido, "\n", "&lt;br/&gt;")

			dotContent.WriteString(fmt.Sprintf("  bloque%d [label=<\n", idBloque))
			dotContent.WriteString("    <table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n")
			dotContent.WriteString(fmt.Sprintf("      <tr><td bgcolor=\"#87CEFA\">Bloque Archivo %d</td></tr>\n", idBloque))
			dotContent.WriteString(fmt.Sprintf("      <tr><td>%s</td></tr>\n", contenido))
			dotContent.WriteString("    </table>\n")
			dotContent.WriteString("  >, style=filled, fillcolor=\"#E6F5FF\"];\n")
		}
	}

	// Generar conexiones inodo -> bloque (AD[X])
	for _, idInodo := range inodosEnUso {
		inodo := inodosInfo[idInodo]
		for j := 0; j < 12; j++ {
			idBloque := inodo.IBlock[j]
			if idBloque != -1 && tipoBloque[idBloque] != "" {
				dotContent.WriteString(fmt.Sprintf("  inodo%d -> bloque%d [label=\"AD[%d]\"];\n", idInodo, idBloque, j))
			}
		}
	}

	// Generar conexiones bloque -> inodo (nombres)
	for idBloque, dirBlk := range bloquesCarpeta {
		for _, entry := range dirBlk.Content {
			if entry.BInodo == -1 {
				continue
			}
			entryName := strings.TrimRight(string(entry.BName[:]), "\x00")
			if entryName != "." && entryName != ".." && entryName != "" {
				_, existeInodo := inodosInfo[entry.BInodo]
				if existeInodo {
					dotContent.WriteString(fmt.Sprintf("  bloque%d -> inodo%d [label=\"%s\"];\n", idBloque, entry.BInodo, entryName))
				}
			}
		}
	}

	// Estructura de rangos
	dotContent.WriteString("  { rank=min; inodo0; }\n")
	for nivel := 1; nivel <= 6; nivel++ {
		dotContent.WriteString(fmt.Sprintf("  subgraph nivel_%d {\n", nivel))
		dotContent.WriteString("    rank=same;\n")
		dotContent.WriteString("  }\n")
	}

	dotContent.WriteString("}\n")

	// Generar imagen
	return g.gv.Generate(dotContent.String(), out)
}

func (g *ReportGenerator) SuperBlock(id, out string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Leer SuperBlock desde el filesystem
	sb, _, err := g.fsRepo.ReadSuper(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %w", err)
	}

	// Convertir timestamps a fechas legibles
	mtimeStr := "N/A"
	umtimeStr := "N/A"
	if sb.SMtime > 0 {
		mtimeStr = time.Unix(sb.SMtime, 0).Format("02/01/2006 15:04:05")
	}
	if sb.SUmtime > 0 {
		umtimeStr = time.Unix(sb.SUmtime, 0).Format("02/01/2006 15:04:05")
	}

	// Generar DOT para Graphviz con formato HTML table (similar al reporte MBR)
	var dotContent strings.Builder
	dotContent.WriteString("digraph {\n")
	dotContent.WriteString("  node [shape=plaintext]\n\n")
	dotContent.WriteString("  TablaSuperBlock [\n")
	dotContent.WriteString("    label=<\n")
	dotContent.WriteString("      <table border=\"1\" cellborder=\"1\" cellspacing=\"0\">\n")
	dotContent.WriteString("        <tr>\n")
	dotContent.WriteString("          <td bgcolor=\"SlateBlue\" colspan=\"2\"><b>Reporte de SuperBloque</b></td>\n")
	dotContent.WriteString("        </tr>\n")

	// Información del SuperBlock con colores alternados
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"Azure\">s_filesystem_type</td>\n          <td bgcolor=\"Azure\">%d</td>\n        </tr>\n", sb.SFilesystemType))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#AFA1D1\">s_inodes_count</td>\n          <td bgcolor=\"#AFA1D1\">%d</td>\n        </tr>\n", sb.SInodesCount))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"Azure\">s_blocks_count</td>\n          <td bgcolor=\"Azure\">%d</td>\n        </tr>\n", sb.SBlocksCount))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#AFA1D1\">s_free_inodes_count</td>\n          <td bgcolor=\"#AFA1D1\">%d</td>\n        </tr>\n", sb.SFreeInodesCount))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"Azure\">s_free_blocks_count</td>\n          <td bgcolor=\"Azure\">%d</td>\n        </tr>\n", sb.SFreeBlocksCount))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#AFA1D1\">s_mtime</td>\n          <td bgcolor=\"#AFA1D1\">%s</td>\n        </tr>\n", mtimeStr))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"Azure\">s_umtime</td>\n          <td bgcolor=\"Azure\">%s</td>\n        </tr>\n", umtimeStr))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#AFA1D1\">s_magic</td>\n          <td bgcolor=\"#AFA1D1\">0x%X</td>\n        </tr>\n", sb.SMagic))

	// Separador visual
	dotContent.WriteString("        <tr>\n")
	dotContent.WriteString("          <td bgcolor=\"LightGray\" colspan=\"2\"><b>Offsets de Estructuras</b></td>\n")
	dotContent.WriteString("        </tr>\n")

	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightCyan\">s_bm_inode_start</td>\n          <td bgcolor=\"LightCyan\">%d (0x%X)</td>\n        </tr>\n", sb.SBmInodeStart, sb.SBmInodeStart))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#B0E0E6\">s_bm_block_start</td>\n          <td bgcolor=\"#B0E0E6\">%d (0x%X)</td>\n        </tr>\n", sb.SBmBlockStart, sb.SBmBlockStart))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightCyan\">s_inode_start</td>\n          <td bgcolor=\"LightCyan\">%d (0x%X)</td>\n        </tr>\n", sb.SInodeStart, sb.SInodeStart))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#B0E0E6\">s_block_start</td>\n          <td bgcolor=\"#B0E0E6\">%d (0x%X)</td>\n        </tr>\n", sb.SBlockStart, sb.SBlockStart))

	// Información adicional calculada
	dotContent.WriteString("        <tr>\n")
	dotContent.WriteString("          <td bgcolor=\"LightGray\" colspan=\"2\"><b>Información Calculada</b></td>\n")
	dotContent.WriteString("        </tr>\n")

	inodosUsados := sb.SInodesCount - sb.SFreeInodesCount
	bloquesUsados := sb.SBlocksCount - sb.SFreeBlocksCount
	porcentajeInodos := float64(inodosUsados) / float64(sb.SInodesCount) * 100.0
	porcentajeBloques := float64(bloquesUsados) / float64(sb.SBlocksCount) * 100.0

	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"LightYellow\">Inodos en uso</td>\n          <td bgcolor=\"LightYellow\">%d (%.2f%%)</td>\n        </tr>\n", inodosUsados, porcentajeInodos))
	dotContent.WriteString(fmt.Sprintf("        <tr>\n          <td bgcolor=\"#FFFFE0\">Bloques en uso</td>\n          <td bgcolor=\"#FFFFE0\">%d (%.2f%%)</td>\n        </tr>\n", bloquesUsados, porcentajeBloques))

	dotContent.WriteString("      </table>\n")
	dotContent.WriteString("    >\n")
	dotContent.WriteString("  ]\n")
	dotContent.WriteString("}")

	// Generar imagen usando Graphviz
	return g.gv.Generate(dotContent.String(), out)
}

func (g *ReportGenerator) File(id, out, filePath string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Buscar información del disco y partición montada
	var diskPath, diskName string
	for _, mount := range g.mountStore.List() {
		if mount.ID == id {
			diskPath = mount.Path
			// Extraer nombre del disco del path
			parts := strings.Split(diskPath, "/")
			diskName = strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
			break
		}
	}
	if diskPath == "" {
		return "", fmt.Errorf("id no montado: %s", id)
	}

	// Parsear el path
	pathParts := strings.Split(strings.Trim(filePath, "/"), "/")
	if pathParts[0] == "" {
		pathParts = pathParts[1:]
	}

	// Usar Cat para leer el contenido del archivo como root (uid=1, gid=1)
	// ya que es un reporte del sistema
	fileContent, err := g.fsRepo.Cat(id, [][]string{pathParts}, 1, 1)
	if err != nil {
		return "", fmt.Errorf("error leyendo archivo: %w", err)
	}

	// Crear el reporte con formato profesional
	var report strings.Builder
	report.WriteString("REPORTE DE ARCHIVO\n")
	report.WriteString("=================\n\n")
	report.WriteString(fmt.Sprintf("Disco: %s\n", diskName))
	report.WriteString(fmt.Sprintf("Partición: %s\n", id))
	report.WriteString(fmt.Sprintf("Archivo: %s\n\n", filePath))
	report.WriteString("CONTENIDO:\n")
	report.WriteString("----------\n\n")
	report.WriteString(fileContent)

	// Asegurar que termina con salto de línea
	if !strings.HasSuffix(fileContent, "\n") {
		report.WriteString("\n")
	}

	// Escribir reporte al archivo de salida
	if err := os.WriteFile(out, []byte(report.String()), 0o644); err != nil {
		return "", err
	}

	return out, nil
}

func (g *ReportGenerator) LS(id, out, pathForLs string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Leer SuperBlock
	sb, _, err := g.fsRepo.ReadSuper(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo superbloque: %w", err)
	}

	// Buscar disco path
	var diskPath string
	for _, mount := range g.mountStore.List() {
		if mount.ID == id {
			diskPath = mount.Path
			break
		}
	}
	if diskPath == "" {
		return "", fmt.Errorf("id no montado: %s", id)
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Parsear el path
	pathParts := strings.Split(strings.Trim(pathForLs, "/"), "/")
	if pathForLs == "/" || pathForLs == "" {
		pathParts = []string{}
	} else if pathParts[0] == "" {
		pathParts = pathParts[1:]
	}

	// Navegar hasta el directorio
	currentInodeIdx := int32(0) // raíz
	constructedPath := ""

	for _, name := range pathParts {
		if name == "" {
			continue
		}

		// Leer inodo actual
		var currentInode models.Inode
		inodeSize := int64(binary.Size(currentInode))
		offset := sb.SInodeStart + int64(currentInodeIdx)*inodeSize
		if _, err := f.Seek(offset, 0); err != nil {
			return "", fmt.Errorf("error navegando (inodo %d): %w", currentInodeIdx, err)
		}
		if err := binary.Read(f, binary.LittleEndian, &currentInode); err != nil {
			return "", fmt.Errorf("error leyendo inodo %d: %w", currentInodeIdx, err)
		}

		// Verificar que es un directorio (el inodo actual debe ser directorio para buscar dentro)
		// IType puede ser byte binario (0/1) O carácter ASCII ('0'/'1')
		if currentInode.IType != models.FileTypeFolder && currentInode.IType != '0' {
			return "", fmt.Errorf("'%s' (inodo %d, tipo=%d/'%c') no es un directorio, no se puede buscar '%s' dentro",
				constructedPath, currentInodeIdx, currentInode.IType, currentInode.IType, name)
		}

		// Buscar entrada en el directorio
		found := false
		for i := 0; i < 12; i++ {
			blkIdx := currentInode.IBlock[i]
			if blkIdx == -1 {
				continue // Seguir buscando en otros bloques
			}

			var dirBlk models.FolderBlock
			blkOffset := sb.SBlockStart + int64(blkIdx)*64
			if _, err := f.Seek(blkOffset, 0); err != nil {
				continue
			}
			if err := binary.Read(f, binary.LittleEndian, &dirBlk); err != nil {
				continue
			}

			for _, entry := range dirBlk.Content {
				if entry.BInodo == -1 {
					continue
				}
				entryName := strings.TrimRight(string(entry.BName[:]), "\x00")
				if entryName == name {
					currentInodeIdx = entry.BInodo
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			return "", fmt.Errorf("no se encontró '%s' en '%s' (inodo %d)", name, constructedPath, currentInodeIdx)
		}

		// Actualizar la ruta construida
		if constructedPath == "" {
			constructedPath = "/" + name
		} else {
			constructedPath = constructedPath + "/" + name
		}
	}

	// Leer el directorio final
	var dirInode models.Inode
	inodeSize := int64(binary.Size(dirInode))
	offset := sb.SInodeStart + int64(currentInodeIdx)*inodeSize
	if _, err := f.Seek(offset, 0); err != nil {
		return "", err
	}
	if err := binary.Read(f, binary.LittleEndian, &dirInode); err != nil {
		return "", err
	}

	// Verificar que el inodo final es un directorio
	if dirInode.IType != models.FileTypeFolder && dirInode.IType != '0' {
		return "", fmt.Errorf("'%s' no es un directorio (tipo=%d)", pathForLs, dirInode.IType)
	}

	// Crear contenido del reporte con formato mejorado
	content := fmt.Sprintf("Directory Listing Report\n")
	content += fmt.Sprintf("ID: %s\n", id)
	content += fmt.Sprintf("Path: %s\n", pathForLs)
	content += "================================================================================\n\n"
	content += fmt.Sprintf("%-4s %-15s %-6s %-8s %-5s %-6s %-19s %-19s %-19s\n",
		"Type", "Name", "Perms", "Owner", "Group", "Size", "Created", "Modified", "Accessed")
	content += strings.Repeat("-", 120) + "\n"

	// Listar entradas
	for i := 0; i < 12; i++ {
		blkIdx := dirInode.IBlock[i]
		if blkIdx == -1 {
			break
		}

		var dirBlk models.FolderBlock
		blkOffset := sb.SBlockStart + int64(blkIdx)*64
		if _, err := f.Seek(blkOffset, 0); err != nil {
			continue
		}
		if err := binary.Read(f, binary.LittleEndian, &dirBlk); err != nil {
			continue
		}

		for _, entry := range dirBlk.Content {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.TrimRight(string(entry.BName[:]), "\x00")

			// Leer inodo de la entrada para obtener toda la información
			var entryInode models.Inode
			entryInodeSize := int64(binary.Size(entryInode))
			entryOffset := sb.SInodeStart + int64(entry.BInodo)*entryInodeSize
			if _, err := f.Seek(entryOffset, 0); err != nil {
				continue
			}
			if err := binary.Read(f, binary.LittleEndian, &entryInode); err != nil {
				continue
			}

			// Tipo (puede ser byte binario 0/1 o carácter ASCII '0'/'1')
			typeStr := "file"
			if entryInode.IType == models.FileTypeFolder || entryInode.IType == '0' {
				typeStr = "dir"
			}

			// Permisos
			permsStr := strings.TrimRight(string(entryInode.IPerm[:]), "\x00")
			if permsStr == "" {
				if entryInode.IType == models.FileTypeFolder || entryInode.IType == '0' {
					permsStr = "755"
				} else {
					permsStr = "664"
				}
			}

			// Propietario y Grupo
			ownerStr := fmt.Sprintf("uid:%d", entryInode.IUid)
			groupStr := fmt.Sprintf("gid:%d", entryInode.IGid)

			// Fechas formateadas
			createdTime := time.Unix(entryInode.ICtime, 0)
			modifiedTime := time.Unix(entryInode.IMtime, 0)
			accessedTime := time.Unix(entryInode.IAtime, 0)

			createdStr := createdTime.Format("2006-01-02 15:04:05")
			modifiedStr := modifiedTime.Format("2006-01-02 15:04:05")
			accessedStr := accessedTime.Format("2006-01-02 15:04:05")

			// Formatear línea con toda la información
			content += fmt.Sprintf("%-4s %-15s %-6s %-8s %-5s %6d %-19s %-19s %-19s\n",
				typeStr, entryName, permsStr, ownerStr, groupStr, entryInode.ISize,
				createdStr, modifiedStr, accessedStr)
		}
	}

	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
		return "", err
	}

	return out, nil
}
