package diskio

import (
	"Backend/core/models"
	"fmt"
	"os"
	"strings"
	"time"
)

// ================================================================
// COMANDO EDIT - Edita el contenido de un archivo existente
// ================================================================

// Edit reemplaza el contenido de un archivo existente
// Reglas:
// - El archivo debe existir
// - Requiere permisos de LECTURA y ESCRITURA
// - Libera bloques viejos y asigna nuevos según el tamaño
// - Actualiza i_size del inodo
// - Registra en Journal (EXT3)
func (r *FileFsRepository) Edit(id string, path []string, content []byte, uid int, gid int) error {
	// 1. Resolver montaje
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	// 2. El contenido ya viene como []byte
	hostContent := content

	// 3. Abrir disco
	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// 4. Leer superblock
	sb, isExt3, err := r.readAnySuperblock(f, region)
	if err != nil {
		return err
	}

	// 5. Write-ahead journal para EXT3
	if isExt3 {
		pathStr := "/" + strings.Join(path, "/")
		j := models.Journal{
			Count: 0,
			Content: models.Information{
				Date: float64(time.Now().Unix()),
			},
		}
		copy(j.Content.Op[:], "EDIT")
		copy(j.Content.Path[:], pathStr)

		// Guardar un preview del contenido
		preview := string(hostContent)
		if len(preview) > 60 {
			preview = preview[:60] + "..."
		}
		copy(j.Content.Content[:], preview)

		_ = r.AppendJournal(f, sb.Ext3, j)
	}

	// 6. Navegar al archivo
	fileIno, _, _, err := r.walkToNode(f, sb, path, region)
	if err != nil {
		return fmt.Errorf("archivo no encontrado: %w", err)
	}

	// 7. Verificar que es un archivo (no directorio)
	if fileIno.IsDir() {
		return fmt.Errorf("la ruta apunta a un directorio, no a un archivo")
	}

	// TODO: Verificar permisos de lectura y escritura

	// 8. Leer bitmaps
	bmBlock, err := r.readBitmapFromSB(f, sb, false, region)
	if err != nil {
		return err
	}

	// 9. Liberar bloques antiguos
	freedBlocks := 0
	for i := 0; i < 12; i++ {
		blkIdx := fileIno.Block(i)
		if blkIdx == -1 {
			break
		}
		bmBlock[blkIdx] = 0
		fileIno.SetBlock(i, -1)
		freedBlocks++
	}

	// 10. Calcular bloques necesarios para el nuevo contenido
	newSize := len(hostContent)
	blocksNeeded := (newSize + models.BlockSizeFile - 1) / models.BlockSizeFile
	if blocksNeeded > 12 {
		blocksNeeded = 12 // Limitar a bloques directos
	}

	// 11. Asignar nuevos bloques
	allocatedBlocks := 0
	for i := 0; i < blocksNeeded; i++ {
		// Buscar bloque libre
		blkIdx := findFreeBlock(bmBlock)
		if blkIdx == -1 {
			return fmt.Errorf("no hay bloques libres disponibles")
		}

		// Marcar como ocupado
		bmBlock[blkIdx] = 1
		fileIno.SetBlock(i, int32(blkIdx))
		allocatedBlocks++

		// Escribir contenido al bloque
		start := i * models.BlockSizeFile
		end := start + models.BlockSizeFile
		if end > newSize {
			end = newSize
		}

		var fileBlock models.FileBlock
		copy(fileBlock.Content[:], hostContent[start:end])

		blockU := BlockUnified{
			FileBlock: fileBlock,
			IsFolder:  false,
		}
		if err := r.writeBlockToSB(f, sb, int32(blkIdx), blockU, region); err != nil {
			return err
		}
	}

	// 12. Actualizar tamaño del inodo
	fileIno.Ext2.ISize = int32(newSize)
	fileIno.SetMtime(time.Now().Unix())

	// 13. Escribir inodo actualizado
	if err := r.writeInodeToSB(f, sb, fileIno, region); err != nil {
		return err
	}

	// 14. Escribir bitmap de bloques
	if err := r.writeBitmapToSB(f, sb, bmBlock, false, region); err != nil {
		return err
	}

	// 15. Actualizar contador de bloques libres en superblock
	netChange := allocatedBlocks - freedBlocks
	if sb.IsExt3 {
		sb.Ext3.FreeBlocks -= int32(netChange)
	} else {
		sb.Ext2.SFreeBlocksCount -= int32(netChange)
	}

	if err := r.writeSuperblock(f, sb, region); err != nil {
		return err
	}

	// Registrar en journal (ignorar errores para no romper la operación)
	fullPath := "/" + strings.Join(path, "/")
	contentInfo := fmt.Sprintf("size=%d", len(hostContent))
	_ = r.JournalAppendPublic(id, "EDIT", fullPath, contentInfo, time.Now().Unix())

	return nil
}

// findFreeBlock busca el primer bloque libre en el bitmap
func findFreeBlock(bitmap []byte) int {
	for i, b := range bitmap {
		if b == 0 {
			return i
		}
	}
	return -1
}
