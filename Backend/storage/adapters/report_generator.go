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

	// Crear contenido del reporte
	content := fmt.Sprintf("Inode Table Report (ID: %s)\n", id)
	content += "================================\n\n"
	content += fmt.Sprintf("Total Inodes: %d\n", sb.SInodesCount)
	content += fmt.Sprintf("Free Inodes:  %d\n", sb.SFreeInodesCount)
	content += fmt.Sprintf("Used Inodes:  %d\n\n", sb.SInodesCount-sb.SFreeInodesCount)

	// Leer todos los inodos
	for i := 0; i < int(sb.SInodesCount); i++ {
		inode, err := g.fsRepo.ReadInode(id, i)
		if err != nil {
			continue
		}

		// Solo mostrar inodos en uso (IUid != -1)
		if inode.IUid == -1 {
			continue
		}

		content += fmt.Sprintf("Inode %d:\n", i)
		content += fmt.Sprintf("  UID:    %d\n", inode.IUid)
		content += fmt.Sprintf("  GID:    %d\n", inode.IGid)
		content += fmt.Sprintf("  Size:   %d bytes\n", inode.ISize)
		content += fmt.Sprintf("  Type:   %d ", inode.IType)
		if inode.IType == 0 {
			content += "(Directory)\n"
		} else {
			content += "(File)\n"
		}
		content += fmt.Sprintf("  Perm:   %s\n", string(inode.IPerm[:]))
		content += fmt.Sprintf("  Atime:  %d\n", inode.IAtime)
		content += fmt.Sprintf("  Ctime:  %d\n", inode.ICtime)
		content += fmt.Sprintf("  Mtime:  %d\n", inode.IMtime)
		content += "  Blocks: "
		for j, blk := range inode.IBlock {
			if blk != -1 {
				content += fmt.Sprintf("%d ", blk)
			} else if j < 12 {
				// Solo mostrar los primeros bloques directos
				break
			}
		}
		content += "\n\n"
	}

	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
		return "", err
	}

	return out, nil
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

	// Leer bitmap de inodos para saber qué bloques revisar
	bmInode, err := g.fsRepo.ReadBitmapInode(id)
	if err != nil {
		return "", fmt.Errorf("error leyendo bitmap inodos: %w", err)
	}

	// Crear contenido del reporte
	content := fmt.Sprintf("Block Report (ID: %s)\n", id)
	content += "================================\n\n"
	content += fmt.Sprintf("Total Blocks: %d\n", sb.SBlocksCount)
	content += fmt.Sprintf("Free Blocks:  %d\n", sb.SFreeBlocksCount)
	content += fmt.Sprintf("Used Blocks:  %d\n\n", sb.SBlocksCount-sb.SFreeBlocksCount)

	// Recorrer inodos y mostrar sus bloques
	for i := 0; i < int(sb.SInodesCount); i++ {
		if bmInode[i] == 0 {
			continue
		}

		inode, err := g.fsRepo.ReadInode(id, i)
		if err != nil || inode.IUid == -1 {
			continue
		}

		content += fmt.Sprintf("Inode %d blocks:\n", i)

		// Mostrar bloques directos
		for j := 0; j < 12; j++ {
			blkIdx := inode.IBlock[j]
			if blkIdx == -1 {
				break
			}

			// Leer contenido del bloque
			blockData, err := g.fsRepo.ReadBlock(id, int(blkIdx))
			if err != nil {
				content += fmt.Sprintf("  Block %d: [Error reading]\n", blkIdx)
				continue
			}

			content += fmt.Sprintf("  Block %d: ", blkIdx)

			// Si es directorio (tipo 0), mostrar entradas
			if inode.IType == 0 {
				content += "[Directory]\n"
			} else {
				// Si es archivo, mostrar primeros bytes
				content += "[File data] "
				displayLen := 32
				if len(blockData) < displayLen {
					displayLen = len(blockData)
				}
				for k := 0; k < displayLen; k++ {
					if blockData[k] >= 32 && blockData[k] < 127 {
						content += string(blockData[k])
					} else {
						content += "."
					}
				}
				content += "...\n"
			}
		}
		content += "\n"
	}

	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
		return "", err
	}

	return out, nil
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

	// Crear archivo de texto con el bitmap
	content := fmt.Sprintf("Bitmap de Inodos (ID: %s)\n", id)
	content += fmt.Sprintf("Total bytes: %d\n\n", len(bitmap))

	// Mostrar el bitmap en formato hexadecimal y binario
	for i, b := range bitmap {
		if i%16 == 0 {
			content += fmt.Sprintf("\n%04x: ", i)
		}
		content += fmt.Sprintf("%02x ", b)
	}
	content += "\n"

	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
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

	// Crear archivo de texto con el bitmap
	content := fmt.Sprintf("Bitmap de Bloques (ID: %s)\n", id)
	content += fmt.Sprintf("Total bytes: %d\n\n", len(bitmap))

	// Mostrar el bitmap en formato hexadecimal y binario
	for i, b := range bitmap {
		if i%16 == 0 {
			content += fmt.Sprintf("\n%04x: ", i)
		}
		content += fmt.Sprintf("%02x ", b)
	}
	content += "\n"

	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
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

	// Generar DOT para Graphviz
	var dotContent strings.Builder
	dotContent.WriteString("digraph G {\n")
	dotContent.WriteString("  node [shape=box];\n")

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

	// Función recursiva para recorrer el árbol
	var walkTree func(inodeIdx int32, nodeName string, parentID string)
	walkTree = func(inodeIdx int32, nodeName string, parentID string) {
		// Leer inodo
		var ino models.Inode
		offset := sb.SInodeStart + int64(inodeIdx)*64
		if _, err := f.Seek(offset, 0); err != nil {
			return
		}
		if err := binary.Read(f, binary.LittleEndian, &ino); err != nil {
			return
		}

		// Crear nodo
		nodeID := fmt.Sprintf("inode%d", inodeIdx)
		label := fmt.Sprintf("%s\\n(Inode %d)", nodeName, inodeIdx)
		dotContent.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", nodeID, label))

		// Conectar con padre
		if parentID != "" {
			dotContent.WriteString(fmt.Sprintf("  %s -> %s;\n", parentID, nodeID))
		}

		// Si es directorio, recorrer sus entradas
		if ino.IType == models.FileTypeFolder {
			for i := 0; i < 12; i++ {
				blkIdx := ino.IBlock[i]
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

				// Recorrer entradas
				for _, entry := range dirBlk.Content {
					if entry.BInodo == -1 {
						continue
					}

					entryName := strings.TrimRight(string(entry.BName[:]), "\x00")
					// Evitar recursión infinita con . y ..
					if entryName == "." || entryName == ".." {
						continue
					}

					walkTree(entry.BInodo, entryName, nodeID)
				}
			}
		}
	}

	// Comenzar desde la raíz (inodo 0)
	walkTree(0, "/", "")

	dotContent.WriteString("}\n")
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

	// Crear archivo de texto con la información del superbloque
	content := fmt.Sprintf("SuperBlock (ID: %s)\n", id)
	content += "================================\n\n"
	content += fmt.Sprintf("Filesystem Type:      %d\n", sb.SFilesystemType)
	content += fmt.Sprintf("Magic Number:         0x%X\n", sb.SMagic)
	content += fmt.Sprintf("Total Inodes:         %d\n", sb.SInodesCount)
	content += fmt.Sprintf("Total Blocks:         %d\n", sb.SBlocksCount)
	content += fmt.Sprintf("Free Inodes:          %d\n", sb.SFreeInodesCount)
	content += fmt.Sprintf("Free Blocks:          %d\n", sb.SFreeBlocksCount)
	content += fmt.Sprintf("Mount Time:           %d\n", sb.SMtime)
	content += fmt.Sprintf("Unmount Time:         %d\n", sb.SUmtime)
	content += fmt.Sprintf("\nBitmap Inode Start:   0x%X\n", sb.SBmInodeStart)
	content += fmt.Sprintf("Bitmap Block Start:   0x%X\n", sb.SBmBlockStart)
	content += fmt.Sprintf("Inode Table Start:    0x%X\n", sb.SInodeStart)
	content += fmt.Sprintf("Block Area Start:     0x%X\n", sb.SBlockStart)

	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
		return "", err
	}

	return out, nil
}

func (g *ReportGenerator) File(id, out, filePath string) (string, error) {
	if err := g.ensureDir(out); err != nil {
		return "", err
	}

	// Parsear el path
	pathParts := strings.Split(strings.Trim(filePath, "/"), "/")
	if pathParts[0] == "" {
		pathParts = pathParts[1:]
	}

	// Usar Cat para leer el contenido del archivo como root (uid=1, gid=1)
	// ya que es un reporte del sistema
	content, err := g.fsRepo.Cat(id, [][]string{pathParts}, 1, 1)
	if err != nil {
		return "", fmt.Errorf("error leyendo archivo: %w", err)
	}

	// Escribir contenido al archivo de salida
	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
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

	for _, name := range pathParts {
		if name == "" {
			continue
		}

		// Leer inodo actual
		var currentInode models.Inode
		offset := sb.SInodeStart + int64(currentInodeIdx)*64
		if _, err := f.Seek(offset, 0); err != nil {
			return "", fmt.Errorf("error navegando: %w", err)
		}
		if err := binary.Read(f, binary.LittleEndian, &currentInode); err != nil {
			return "", fmt.Errorf("error leyendo inodo: %w", err)
		}

		// Buscar entrada en el directorio
		found := false
		for i := 0; i < 12; i++ {
			blkIdx := currentInode.IBlock[i]
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
			return "", fmt.Errorf("directorio no encontrado: %s", name)
		}
	}

	// Leer el directorio final
	var dirInode models.Inode
	offset := sb.SInodeStart + int64(currentInodeIdx)*64
	if _, err := f.Seek(offset, 0); err != nil {
		return "", err
	}
	if err := binary.Read(f, binary.LittleEndian, &dirInode); err != nil {
		return "", err
	}

	if dirInode.IType != models.FileTypeFolder {
		return "", fmt.Errorf("no es un directorio")
	}

	// Crear contenido del reporte
	content := fmt.Sprintf("Directory Listing (ID: %s, Path: %s)\n", id, pathForLs)
	content += "================================\n\n"

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

			// Leer inodo de la entrada para obtener tipo y tamaño
			var entryInode models.Inode
			entryOffset := sb.SInodeStart + int64(entry.BInodo)*64
			if _, err := f.Seek(entryOffset, 0); err != nil {
				continue
			}
			if err := binary.Read(f, binary.LittleEndian, &entryInode); err != nil {
				continue
			}

			typeStr := "file"
			if entryInode.IType == models.FileTypeFolder {
				typeStr = "dir "
			}

			content += fmt.Sprintf("[%s] %-12s  Size: %6d  Inode: %d  Perms: %s\n",
				typeStr, entryName, entryInode.ISize, entry.BInodo, string(entryInode.IPerm[:]))
		}
	}

	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
		return "", err
	}

	return out, nil
}
