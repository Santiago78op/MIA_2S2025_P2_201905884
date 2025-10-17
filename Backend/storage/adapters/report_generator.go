package adapters

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Leer EBRs si hay partición extendida
	var extIndex int = -1
	for i, part := range mbr.Partitions {
		if part.Status == models.PartStatusUsed && part.Type == models.PartTypeExtend {
			extIndex = i
			break
		}
	}

	// Generar DOT para Graphviz
	var dotContent strings.Builder
	dotContent.WriteString("digraph MBR {\n")
	dotContent.WriteString("  rankdir=LR;\n")
	dotContent.WriteString("  node [shape=record];\n\n")

	// Nodo del MBR principal
	mbrLabel := fmt.Sprintf("MBR|Size: %d bytes|Fit: %c|Signature: %d", mbr.Size, mbr.Fit, mbr.Signature)
	dotContent.WriteString(fmt.Sprintf("  mbr [label=\"%s\"];\n\n", mbrLabel))

	// Procesar particiones
	for i, part := range mbr.Partitions {
		if part.Status == 0 {
			continue
		}

		name := string(part.Name[:])
		// Trim null bytes
		for j, c := range name {
			if c == 0 {
				name = name[:j]
				break
			}
		}

		nodeID := fmt.Sprintf("part%d", i)
		partLabel := fmt.Sprintf("Partition %d|Name: %s|Type: %c|Start: %d|Size: %d bytes|Fit: %c",
			i+1, name, part.Type, part.Start, part.Size, part.Fit)
		dotContent.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", nodeID, partLabel))
		dotContent.WriteString(fmt.Sprintf("  mbr -> %s;\n", nodeID))

		// Si es partición extendida, leer sus EBRs
		if part.Type == models.PartTypeExtend && extIndex != -1 && i == extIndex {
			// Abrir archivo del disco para leer EBRs
			f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
			if err == nil {
				defer f.Close()

				// Leer cadena de EBRs
				currentPos := part.Start
				ebrCount := 0
				prevEbrID := nodeID

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

					ebrName := string(ebr.Name[:])
					for j, c := range ebrName {
						if c == 0 {
							ebrName = ebrName[:j]
							break
						}
					}

					ebrID := fmt.Sprintf("ebr%d", ebrCount)
					ebrLabel := fmt.Sprintf("EBR %d|Name: %s|Start: %d|Size: %d|Fit: %c|Next: %d",
						ebrCount+1, ebrName, ebr.Start, ebr.Size, ebr.Fit, ebr.Next)
					dotContent.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", ebrID, ebrLabel))
					dotContent.WriteString(fmt.Sprintf("  %s -> %s;\n", prevEbrID, ebrID))

					prevEbrID = ebrID
					ebrCount++
					currentPos = ebr.Next
				}
			}
		}
	}

	dotContent.WriteString("}\n")

	// Generar imagen usando Graphviz
	return g.gv.Generate(dotContent.String(), out)
}

func (g *ReportGenerator) Disk(id, out string) (string, error) {
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

	// Generar DOT para Graphviz
	var dotContent strings.Builder
	dotContent.WriteString("digraph G {\n")
	dotContent.WriteString("  node [shape=box];\n")

	// Calcular espacio usado y libre
	type segment struct {
		start int64
		end   int64
		label string
	}

	segments := []segment{
		{0, 512, "MBR"}, // Típicamente el MBR ocupa 512 bytes
	}

	// Agregar particiones
	for i, part := range mbr.Partitions {
		if part.Status == 0 {
			continue
		}

		name := string(part.Name[:])
		for j, c := range name {
			if c == 0 {
				name = name[:j]
				break
			}
		}

		nodeID := fmt.Sprintf("part%d", i)
		label := fmt.Sprintf("%s\\n%c\\n%.2f MB", name, part.Type, float64(part.Size)/(1024*1024))
		dotContent.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", nodeID, label))
		dotContent.WriteString(fmt.Sprintf("  mbr -> %s;\n", nodeID))

		segments = append(segments, segment{
			start: part.Start,
			end:   part.Start + part.Size,
			label: name,
		})
	}

	// Calcular espacio libre
	// Ordenar segmentos por start
	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			if segments[j].start < segments[i].start {
				segments[i], segments[j] = segments[j], segments[i]
			}
		}
	}

	// Buscar huecos
	freeCount := 0
	for i := 0; i < len(segments)-1; i++ {
		gap := segments[i+1].start - segments[i].end
		if gap > 0 {
			freeCount++
			nodeID := fmt.Sprintf("free%d", freeCount)
			label := fmt.Sprintf("Free\\n%.2f MB", float64(gap)/(1024*1024))
			dotContent.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", nodeID, label))
			dotContent.WriteString(fmt.Sprintf("  mbr -> %s;\n", nodeID))
		}
	}

	// Espacio libre al final
	lastEnd := segments[len(segments)-1].end
	if mbr.Size > lastEnd {
		gap := mbr.Size - lastEnd
		freeCount++
		nodeID := fmt.Sprintf("free%d", freeCount)
		label := fmt.Sprintf("Free\\n%.2f MB", float64(gap)/(1024*1024))
		dotContent.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", nodeID, label))
		dotContent.WriteString(fmt.Sprintf("  mbr -> %s;\n", nodeID))
	}

	dotContent.WriteString("}\n")
	return g.gv.Generate(dotContent.String(), out)
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

	// Usar Cat para leer el contenido del archivo
	content, err := g.fsRepo.Cat(id, [][]string{pathParts})
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
