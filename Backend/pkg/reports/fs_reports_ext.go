package reports

import (
	"fmt"
	"os"

	"MIA_2S2025_P2_201905884/internal/disk"
	"MIA_2S2025_P2_201905884/internal/fs/ext2"
)

// GenerateINODEReport genera un reporte DOT de todos los inodos
func GenerateINODEReport(diskPath, partName, outputPath string) error {
	// Obtener información de partición
	part, partStart, err := getMountedPartitionInfo(diskPath, partName)
	if err != nil {
		return fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	f, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	data, err := disk.ReadBytesAt(f, partStart, ext2.SUPERBLOCK_SIZE_ACTUAL)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	sb, err := ext2.DeserializeSuperblock(data)
	if err != nil {
		return fmt.Errorf("error al deserializar superbloque: %v", err)
	}

	// Generar DOT con todos los inodos
	d := newDot("Reporte de Inodos - " + trimPartName(part.Name))
	d.line(`rankdir=TB;`)
	d.line(`node [shape=record];`)

	// Leer bitmap de inodos
	bitmapData, err := disk.ReadBytesAt(f, partStart+int64(sb.S_bm_inode_start), int(sb.S_inodes_count))
	if err != nil {
		return fmt.Errorf("error al leer bitmap de inodos: %v", err)
	}

	// Recorrer todos los inodos
	inodeSize := int(sb.S_inode_size)
	for i := int32(0); i < sb.S_inodes_count; i++ {
		// Verificar si el inodo está en uso
		if i < int32(len(bitmapData)) && bitmapData[i] == 1 {
			// Leer inodo
			inodeOffset := partStart + int64(sb.S_inode_start) + int64(i)*int64(inodeSize)
			inodeData, err := disk.ReadBytesAt(f, inodeOffset, inodeSize)
			if err != nil {
				continue
			}

			inode, err := ext2.DeserializeInode(inodeData)
			if err != nil {
				continue
			}

			// Crear nodo DOT para este inodo
			nodeName := fmt.Sprintf("inode%d", i)
			rows := []string{
				fmt.Sprintf("<f0> Inodo %d", i),
				rowKV("UID", fmt.Sprintf("%d", inode.IUid)),
				rowKV("GID", fmt.Sprintf("%d", inode.IGid)),
				rowKV("Tamaño", fmt.Sprintf("%d bytes", inode.IS)),
				rowKV("Tipo", inodeTypeString(inode.IType)),
				rowKV("Permisos", fmt.Sprintf("%c%c%c", inode.IPerm[0], inode.IPerm[1], inode.IPerm[2])),
			}

			// Agregar bloques usados
			for j, block := range inode.IBlock {
				if block != -1 {
					blockType := "Directo"
					if j >= 12 {
						blockType = "Indirecto"
					}
					rows = append(rows, rowKV(fmt.Sprintf("Bloque %d (%s)", j, blockType), fmt.Sprintf("%d", block)))
				}
			}

			d.line(fmt.Sprintf(`%s [label="%s"];`, nodeName, tableRecord(rows)))

			// Conectar inodos consecutivos
			if i > 0 {
				d.line(fmt.Sprintf(`inode%d -> inode%d [style=dashed];`, i-1, i))
			}
		}
	}

	dotContent := d.close()
	return renderDOT(dotContent, outputPath)
}

func inodeTypeString(t byte) string {
	if t == 0 {
		return "Carpeta"
	}
	return "Archivo"
}

// GenerateBLOCKReport genera un reporte DOT de todos los bloques
func GenerateBLOCKReport(diskPath, partName, outputPath string) error {
	// Obtener información de partición
	part, partStart, err := getMountedPartitionInfo(diskPath, partName)
	if err != nil {
		return fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	f, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	data, err := disk.ReadBytesAt(f, partStart, ext2.SUPERBLOCK_SIZE_ACTUAL)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	sb, err := ext2.DeserializeSuperblock(data)
	if err != nil {
		return fmt.Errorf("error al deserializar superbloque: %v", err)
	}

	// Generar DOT con todos los bloques
	d := newDot("Reporte de Bloques - " + trimPartName(part.Name))
	d.line(`rankdir=TB;`)
	d.line(`node [shape=record];`)

	// Leer bitmap de bloques
	bitmapData, err := disk.ReadBytesAt(f, partStart+int64(sb.S_bm_block_start), int(sb.S_blocks_count))
	if err != nil {
		return fmt.Errorf("error al leer bitmap de bloques: %v", err)
	}

	// Recorrer bloques (limitar a primeros 20 para no saturar)
	blockSize := int(sb.S_block_size)
	maxBlocks := int32(20)
	if sb.S_blocks_count < maxBlocks {
		maxBlocks = sb.S_blocks_count
	}

	for i := int32(0); i < maxBlocks; i++ {
		// Verificar si el bloque está en uso
		if i < int32(len(bitmapData)) && bitmapData[i] == 1 {
			// Leer bloque
			blockOffset := partStart + int64(sb.S_block_start) + int64(i)*int64(blockSize)
			blockData, err := disk.ReadBytesAt(f, blockOffset, blockSize)
			if err != nil {
				continue
			}

			// Crear nodo DOT para este bloque
			nodeName := fmt.Sprintf("block%d", i)
			rows := []string{
				fmt.Sprintf("<f0> Bloque %d", i),
			}

			// Mostrar contenido en hexadecimal (primeros 32 bytes)
			maxBytes := 32
			if len(blockData) < maxBytes {
				maxBytes = len(blockData)
			}
			hexStr := ""
			for j := 0; j < maxBytes; j++ {
				hexStr += fmt.Sprintf("%02X ", blockData[j])
				if (j+1)%16 == 0 {
					rows = append(rows, rowKV(fmt.Sprintf("Offset %02X", j-15), hexStr))
					hexStr = ""
				}
			}
			if hexStr != "" {
				rows = append(rows, rowKV("Datos", hexStr))
			}

			d.line(fmt.Sprintf(`%s [label="%s"];`, nodeName, tableRecord(rows)))

			// Conectar bloques consecutivos
			if i > 0 {
				d.line(fmt.Sprintf(`block%d -> block%d [style=dashed];`, i-1, i))
			}
		}
	}

	dotContent := d.close()
	return renderDOT(dotContent, outputPath)
}

// GenerateBMINODEReport genera un reporte DOT del bitmap de inodos
func GenerateBMINODEReport(diskPath, partName, outputPath string) error {
	// Obtener información de partición
	part, partStart, err := getMountedPartitionInfo(diskPath, partName)
	if err != nil {
		return fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	f, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	data, err := disk.ReadBytesAt(f, partStart, ext2.SUPERBLOCK_SIZE_ACTUAL)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	sb, err := ext2.DeserializeSuperblock(data)
	if err != nil {
		return fmt.Errorf("error al deserializar superbloque: %v", err)
	}

	// Leer bitmap de inodos
	bitmapData, err := disk.ReadBytesAt(f, partStart+int64(sb.S_bm_inode_start), int(sb.S_inodes_count))
	if err != nil {
		return fmt.Errorf("error al leer bitmap de inodos: %v", err)
	}

	// Generar DOT con bitmap real
	d := newDot("Bitmap de Inodos - " + trimPartName(part.Name))
	d.line(`rankdir=TB; node [shape=plaintext];`)

	// Generar tabla con valores reales
	row := `<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4">`
	row += `<TR><TD COLSPAN="20" BGCOLOR="lightblue"><B>Bitmap de Inodos</B></TD></TR>`

	// Generar filas de 20 columnas
	for i := 0; i < len(bitmapData); i++ {
		if i%20 == 0 {
			if i > 0 {
				row += `</TR>`
			}
			row += `<TR>`
		}

		val := fmt.Sprintf("%d", bitmapData[i])
		bgcolor := "white"
		if bitmapData[i] == 1 {
			bgcolor = "lightgreen"
		}
		row += fmt.Sprintf(`<TD BGCOLOR="%s">%s</TD>`, bgcolor, val)
	}
	row += `</TR></TABLE>>`
	d.line(`bm_inode [label=` + row + `];`)

	// Agregar información de estadísticas
	used := 0
	for _, bit := range bitmapData {
		if bit == 1 {
			used++
		}
	}
	statsRow := fmt.Sprintf(`<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
		<TR><TD>Total Inodos</TD><TD>%d</TD></TR>
		<TR><TD>Inodos Usados</TD><TD>%d</TD></TR>
		<TR><TD>Inodos Libres</TD><TD>%d</TD></TR>
	</TABLE>>`, len(bitmapData), used, len(bitmapData)-used)
	d.line(`stats [label=` + statsRow + `];`)
	d.line(`bm_inode -> stats [style=dashed];`)

	dot := d.close()
	return renderDOT(dot, outputPath)
}

// GenerateBMBLOCKReport genera un reporte DOT del bitmap de bloques
func GenerateBMBLOCKReport(diskPath, partName, outputPath string) error {
	// Obtener información de partición
	part, partStart, err := getMountedPartitionInfo(diskPath, partName)
	if err != nil {
		return fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	f, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	data, err := disk.ReadBytesAt(f, partStart, ext2.SUPERBLOCK_SIZE_ACTUAL)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	sb, err := ext2.DeserializeSuperblock(data)
	if err != nil {
		return fmt.Errorf("error al deserializar superbloque: %v", err)
	}

	// Leer bitmap de bloques
	bitmapData, err := disk.ReadBytesAt(f, partStart+int64(sb.S_bm_block_start), int(sb.S_blocks_count))
	if err != nil {
		return fmt.Errorf("error al leer bitmap de bloques: %v", err)
	}

	// Generar DOT con bitmap real
	d := newDot("Bitmap de Bloques - " + trimPartName(part.Name))
	d.line(`rankdir=TB; node [shape=plaintext];`)

	// Generar tabla con valores reales
	row := `<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4">`
	row += `<TR><TD COLSPAN="20" BGCOLOR="lightcoral"><B>Bitmap de Bloques</B></TD></TR>`

	// Generar filas de 20 columnas
	for i := 0; i < len(bitmapData); i++ {
		if i%20 == 0 {
			if i > 0 {
				row += `</TR>`
			}
			row += `<TR>`
		}

		val := fmt.Sprintf("%d", bitmapData[i])
		bgcolor := "white"
		if bitmapData[i] == 1 {
			bgcolor = "lightyellow"
		}
		row += fmt.Sprintf(`<TD BGCOLOR="%s">%s</TD>`, bgcolor, val)
	}
	row += `</TR></TABLE>>`
	d.line(`bm_block [label=` + row + `];`)

	// Agregar información de estadísticas
	used := 0
	for _, bit := range bitmapData {
		if bit == 1 {
			used++
		}
	}
	statsRow := fmt.Sprintf(`<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
		<TR><TD>Total Bloques</TD><TD>%d</TD></TR>
		<TR><TD>Bloques Usados</TD><TD>%d</TD></TR>
		<TR><TD>Bloques Libres</TD><TD>%d</TD></TR>
	</TABLE>>`, len(bitmapData), used, len(bitmapData)-used)
	d.line(`stats [label=` + statsRow + `];`)
	d.line(`bm_block -> stats [style=dashed];`)

	dot := d.close()
	return renderDOT(dot, outputPath)
}

// GenerateTREEReport genera un reporte DOT del árbol de directorios
func GenerateTREEReport(diskPath, partName, outputPath string) error {
	// Obtener información de partición
	part, partStart, err := getMountedPartitionInfo(diskPath, partName)
	if err != nil {
		return fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	f, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	data, err := disk.ReadBytesAt(f, partStart, ext2.SUPERBLOCK_SIZE_ACTUAL)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	sb, err := ext2.DeserializeSuperblock(data)
	if err != nil {
		return fmt.Errorf("error al deserializar superbloque: %v", err)
	}

	// Generar DOT con árbol
	d := newDot("Árbol de Directorios - " + trimPartName(part.Name))
	d.line(`rankdir=TB; node [shape=box, style=filled];`)

	// Contador de nodos
	nodeCounter := 0

	// Función recursiva para recorrer el árbol
	var buildTree func(inodeIdx int32, path string, parentNode string) error
	buildTree = func(inodeIdx int32, path string, parentNode string) error {
		// Leer inodo
		inodeSize := int(sb.S_inode_size)
		inodeOffset := partStart + int64(sb.S_inode_start) + int64(inodeIdx)*int64(inodeSize)
		inodeData, err := disk.ReadBytesAt(f, inodeOffset, inodeSize)
		if err != nil {
			return err
		}

		inode, err := ext2.DeserializeInode(inodeData)
		if err != nil {
			return err
		}

		// Crear nodo actual
		nodeCounter++
		currentNode := fmt.Sprintf("N%d", nodeCounter)

		label := fmt.Sprintf("%s\\n", path)
		label += fmt.Sprintf("Permisos: %c%c%c\\n", inode.IPerm[0], inode.IPerm[1], inode.IPerm[2])
		label += fmt.Sprintf("UID: %d GID: %d", inode.IUid, inode.IGid)

		color := "lightblue"
		if inode.IType == 0 { // Carpeta
			color = "lightyellow"
		}

		d.line(fmt.Sprintf(`%s [label="%s", fillcolor=%s];`, currentNode, label, color))

		// Conectar con padre si existe
		if parentNode != "" {
			d.line(fmt.Sprintf(`%s -> %s;`, parentNode, currentNode))
		}

		// Si es carpeta, recorrer sus entradas
		if inode.IType == 0 {
			for _, blockIdx := range inode.IBlock {
				if blockIdx != -1 {
					// Leer bloque de carpeta
					blockSize := int(sb.S_block_size)
					blockOffset := partStart + int64(sb.S_block_start) + int64(blockIdx)*int64(blockSize)
					blockData, err := disk.ReadBytesAt(f, blockOffset, blockSize)
					if err != nil {
						continue
					}

					folderBlock, err := ext2.DeserializeFolderBlock(blockData)
					if err != nil {
						continue
					}

					// Procesar cada entrada
					for _, entry := range folderBlock.BContent {
						if entry.BInodo != -1 {
							entryName := entry.GetName()
							if entryName != "." && entryName != ".." && entryName != "" {
								newPath := path
								if path == "/" {
									newPath = "/" + entryName
								} else {
									newPath = path + "/" + entryName
								}
								buildTree(entry.BInodo, newPath, currentNode)
							}
						}
					}
				}
			}
		}

		return nil
	}

	// Comenzar desde el inodo raíz (0)
	if err := buildTree(0, "/", ""); err != nil {
		return fmt.Errorf("error al construir árbol: %v", err)
	}

	dot := d.close()
	return renderDOT(dot, outputPath)
}

// GenerateFILEReport genera un reporte DOT del contenido de un archivo
func GenerateFILEReport(diskPath, partName, filePath, outputPath string) error {
	// Obtener información de partición
	part, partStart, err := getMountedPartitionInfo(diskPath, partName)
	if err != nil {
		return fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	f, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	data, err := disk.ReadBytesAt(f, partStart, ext2.SUPERBLOCK_SIZE_ACTUAL)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	sb, err := ext2.DeserializeSuperblock(data)
	if err != nil {
		return fmt.Errorf("error al deserializar superbloque: %v", err)
	}

	// Buscar el archivo en el sistema de archivos
	inodeIdx, err := findFileInode(f, sb, partStart, filePath)
	if err != nil {
		return fmt.Errorf("archivo no encontrado: %v", err)
	}

	// Leer inodo del archivo
	inodeSize := int(sb.S_inode_size)
	inodeOffset := partStart + int64(sb.S_inode_start) + int64(inodeIdx)*int64(inodeSize)
	inodeData, err := disk.ReadBytesAt(f, inodeOffset, inodeSize)
	if err != nil {
		return fmt.Errorf("error al leer inodo: %v", err)
	}

	inode, err := ext2.DeserializeInode(inodeData)
	if err != nil {
		return fmt.Errorf("error al deserializar inodo: %v", err)
	}

	// Leer contenido del archivo
	var fileContent []byte
	for _, blockIdx := range inode.IBlock {
		if blockIdx != -1 {
			blockSize := int(sb.S_block_size)
			blockOffset := partStart + int64(sb.S_block_start) + int64(blockIdx)*int64(blockSize)
			blockData, err := disk.ReadBytesAt(f, blockOffset, blockSize)
			if err != nil {
				continue
			}
			fileContent = append(fileContent, blockData...)
		}
	}

	// Truncar al tamaño real
	if int32(len(fileContent)) > inode.IS {
		fileContent = fileContent[:inode.IS]
	}

	// Generar DOT
	d := newDot("Contenido de Archivo - " + filePath)
	d.line(`rankdir=TB; node [shape=plaintext];`)

	content := `<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4">`
	content += fmt.Sprintf(`<TR><TD COLSPAN="2" BGCOLOR="lightblue"><B>%s</B></TD></TR>`, escape(filePath))
	content += fmt.Sprintf(`<TR><TD>Disco</TD><TD>%s</TD></TR>`, escape(diskPath))
	content += fmt.Sprintf(`<TR><TD>Partición</TD><TD>%s</TD></TR>`, trimPartName(part.Name))
	content += fmt.Sprintf(`<TR><TD>Tamaño</TD><TD>%d bytes</TD></TR>`, inode.IS)
	content += fmt.Sprintf(`<TR><TD>Permisos</TD><TD>%c%c%c</TD></TR>`, inode.IPerm[0], inode.IPerm[1], inode.IPerm[2])
	content += fmt.Sprintf(`<TR><TD>UID</TD><TD>%d</TD></TR>`, inode.IUid)
	content += fmt.Sprintf(`<TR><TD>GID</TD><TD>%d</TD></TR>`, inode.IGid)
	content += `<TR><TD COLSPAN="2" BGCOLOR="lightyellow"><B>Contenido</B></TD></TR>`

	// Mostrar contenido (escapar HTML)
	contentStr := escape(string(fileContent))
	lines := splitIntoLines(contentStr, 60)
	for _, line := range lines {
		content += fmt.Sprintf(`<TR><TD COLSPAN="2" ALIGN="LEFT"><FONT FACE="monospace">%s</FONT></TD></TR>`, line)
	}

	content += `</TABLE>>`
	d.line(`file [label=` + content + `];`)

	dot := d.close()
	return renderDOT(dot, outputPath)
}

// findFileInode busca un archivo por su ruta y retorna su índice de inodo
func findFileInode(f *os.File, sb *ext2.Superblock, partStart int64, path string) (int32, error) {
	// Simplificación: buscar solo en raíz
	// Para una implementación completa, necesitarías parsear la ruta completa
	if path == "/users.txt" || path == "users.txt" {
		return 1, nil // Asumiendo que users.txt está en inodo 1
	}
	return -1, fmt.Errorf("archivo no encontrado: %s", path)
}

// splitIntoLines divide un string en líneas de longitud máxima
func splitIntoLines(s string, maxLen int) []string {
	var lines []string
	for len(s) > maxLen {
		lines = append(lines, s[:maxLen])
		s = s[maxLen:]
	}
	if len(s) > 0 {
		lines = append(lines, s)
	}
	return lines
}

// GenerateLSReport genera un reporte tipo ls de un directorio
func GenerateLSReport(diskPath, partName, dirPath, outputPath string) error {
	// Obtener información de partición
	part, partStart, err := getMountedPartitionInfo(diskPath, partName)
	if err != nil {
		return fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	f, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	data, err := disk.ReadBytesAt(f, partStart, ext2.SUPERBLOCK_SIZE_ACTUAL)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	sb, err := ext2.DeserializeSuperblock(data)
	if err != nil {
		return fmt.Errorf("error al deserializar superbloque: %v", err)
	}

	// Buscar el directorio (por ahora asumimos raíz)
	inodeIdx := int32(0) // Raíz
	if dirPath != "/" && dirPath != "" {
		// Para implementación completa, buscar el directorio por ruta
		return fmt.Errorf("solo se soporta directorio raíz actualmente")
	}

	// Leer inodo del directorio
	inodeSize := int(sb.S_inode_size)
	inodeOffset := partStart + int64(sb.S_inode_start) + int64(inodeIdx)*int64(inodeSize)
	inodeData, err := disk.ReadBytesAt(f, inodeOffset, inodeSize)
	if err != nil {
		return fmt.Errorf("error al leer inodo: %v", err)
	}

	inode, err := ext2.DeserializeInode(inodeData)
	if err != nil {
		return fmt.Errorf("error al deserializar inodo: %v", err)
	}

	// Verificar que es un directorio
	if inode.IType != 0 {
		return fmt.Errorf("la ruta especificada no es un directorio")
	}

	// Generar DOT
	d := newDot("Listado de Directorio - " + dirPath + " (Partición: " + trimPartName(part.Name) + ")")
	d.line(`rankdir=TB; node [shape=plaintext];`)

	content := `<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4">`
	content += fmt.Sprintf(`<TR><TD COLSPAN="5" BGCOLOR="lightblue"><B>Directorio: %s</B></TD></TR>`, escape(dirPath))
	content += `<TR BGCOLOR="lightgray"><TD><B>Tipo</B></TD><TD><B>Nombre</B></TD><TD><B>Permisos</B></TD><TD><B>UID</B></TD><TD><B>GID</B></TD></TR>`

	// Leer bloques del directorio
	for _, blockIdx := range inode.IBlock {
		if blockIdx != -1 {
			blockSize := int(sb.S_block_size)
			blockOffset := partStart + int64(sb.S_block_start) + int64(blockIdx)*int64(blockSize)
			blockData, err := disk.ReadBytesAt(f, blockOffset, blockSize)
			if err != nil {
				continue
			}

			folderBlock, err := ext2.DeserializeFolderBlock(blockData)
			if err != nil {
				continue
			}

			// Procesar cada entrada
			for _, entry := range folderBlock.BContent {
				if entry.BInodo != -1 {
					entryName := entry.GetName()
					if entryName == "" {
						continue
					}

					// Leer inodo de la entrada
					entryInodeOffset := partStart + int64(sb.S_inode_start) + int64(entry.BInodo)*int64(inodeSize)
					entryInodeData, err := disk.ReadBytesAt(f, entryInodeOffset, inodeSize)
					if err != nil {
						continue
					}

					entryInode, err := ext2.DeserializeInode(entryInodeData)
					if err != nil {
						continue
					}

					// Determinar tipo
					typeStr := "Archivo"
					if entryInode.IType == 0 {
						typeStr = "Carpeta"
					}

					content += fmt.Sprintf(`<TR><TD>%s</TD><TD>%s</TD><TD>%c%c%c</TD><TD>%d</TD><TD>%d</TD></TR>`,
						typeStr,
						escape(entryName),
						entryInode.IPerm[0], entryInode.IPerm[1], entryInode.IPerm[2],
						entryInode.IUid,
						entryInode.IGid)
				}
			}
		}
	}

	content += `</TABLE>>`
	d.line(`ls [label=` + content + `];`)

	dot := d.close()
	return renderDOT(dot, outputPath)
}
