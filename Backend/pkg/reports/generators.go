package reports

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"MIA_2S2025_P2_201905884/internal/disk"
	"MIA_2S2025_P2_201905884/internal/fs/ext2"
)

// GenerateMBRReport genera reporte MBR desde disco
func GenerateMBRReport(diskPath, outputPath string) error {
	// Leer MBR
	f, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	var mbr disk.MBR
	if err := disk.ReadStruct(f, 0, &mbr); err != nil {
		return fmt.Errorf("error al leer MBR: %v", err)
	}

	// Convertir a modelo de reporte
	mbrInfo := MBRInfo{
		SizeBytes: mbr.SizeBytes,
		CreatedAt: time.Unix(mbr.CreatedAt, 0),
		DiskSig:   mbr.DiskSig,
		Fit:       fitToString(mbr.Fit),
		Parts:     []PartInfo{},
	}

	// Agregar particiones
	for i := 0; i < disk.MaxPrimaries; i++ {
		p := mbr.Parts[i]
		if p.Status == disk.PartStatusFree {
			continue
		}

		partInfo := PartInfo{
			Status: statusToString(p.Status),
			Type:   string(p.Type),
			Fit:    fitToString(p.Fit),
			Start:  p.Start,
			Size:   p.Size,
			Name:   trimPartName(p.Name),
			EBRs:   []EBRInfo{},
		}

		// Si es extendida, leer EBRs
		if p.Type == disk.PartTypeExtended {
			ebrs, err := disk.ListEBRs(f, p.Start, p.Start+p.Size)
			if err == nil {
				for _, ebr := range ebrs {
					if ebr.Status == disk.PartStatusUsed {
						partInfo.EBRs = append(partInfo.EBRs, EBRInfo{
							Status: statusToString(ebr.Status),
							Fit:    fitToString(ebr.Fit),
							Start:  ebr.Start,
							Size:   ebr.Size,
							Next:   ebr.Next,
							Name:   trimPartName(ebr.Name),
						})
					}
				}
			}
		}

		mbrInfo.Parts = append(mbrInfo.Parts, partInfo)
	}

	// Generar DOT
	dotContent := ReportMBR(mbrInfo, Options{
		Title:   "Reporte MBR",
		Rankdir: "TB",
	})

	// Guardar y renderizar
	return renderDOT(dotContent, outputPath)
}

// GenerateDISKReport genera reporte de layout del disco
func GenerateDISKReport(diskPath, outputPath string) error {
	// Leer MBR
	f, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	var mbr disk.MBR
	if err := disk.ReadStruct(f, 0, &mbr); err != nil {
		return fmt.Errorf("error al leer MBR: %v", err)
	}

	// Convertir a modelo
	mbrInfo := MBRInfo{
		SizeBytes: mbr.SizeBytes,
		CreatedAt: time.Unix(mbr.CreatedAt, 0),
		DiskSig:   mbr.DiskSig,
		Fit:       fitToString(mbr.Fit),
		Parts:     []PartInfo{},
	}

	// MBR ocupa espacio
	mbrInfo.Parts = append(mbrInfo.Parts, PartInfo{
		Status: "used",
		Type:   "MBR",
		Name:   "MBR",
		Start:  0,
		Size:   int64(binary_sizeof_mbr()),
	})

	// Agregar particiones
	for i := 0; i < disk.MaxPrimaries; i++ {
		p := mbr.Parts[i]
		if p.Status == disk.PartStatusUsed {
			partInfo := PartInfo{
				Status: "used",
				Type:   string(p.Type),
				Fit:    fitToString(p.Fit),
				Start:  p.Start,
				Size:   p.Size,
				Name:   trimPartName(p.Name),
			}

			// Si es extendida, incluir lógicas
			if p.Type == disk.PartTypeExtended {
				ebrs, err := disk.ListEBRs(f, p.Start, p.Start+p.Size)
				if err == nil {
					for _, ebr := range ebrs {
						if ebr.Status == disk.PartStatusUsed {
							partInfo.EBRs = append(partInfo.EBRs, EBRInfo{
								Status: "used",
								Fit:    fitToString(ebr.Fit),
								Start:  ebr.Start,
								Size:   ebr.Size,
								Next:   ebr.Next,
								Name:   trimPartName(ebr.Name),
							})
						}
					}
				}
			}

			mbrInfo.Parts = append(mbrInfo.Parts, partInfo)
		}
	}

	// Generar DOT
	dotContent := ReportDiskLayout(mbrInfo, Options{
		Title: "Reporte DISK",
	})

	return renderDOT(dotContent, outputPath)
}

// GenerateSuperblockReport genera reporte del superbloque
func GenerateSuperblockReport(diskPath, partName, outputPath string) error {
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

	// Convertir a modelo
	sbInfo := SuperBlock{
		BlockSize:   int(sb.S_block_size),
		InodeSize:   int(sb.S_inode_size),
		CountInodes: int(sb.S_inodes_count),
		CountBlocks: int(sb.S_blocks_count),
		FreeInodes:  int(sb.S_free_inodes_count),
		FreeBlocks:  int(sb.S_free_blocks_count),
		FirstDataAt: int64(sb.S_block_start),
	}

	// Generar DOT
	dotContent := ReportSuperblock(sbInfo, Options{
		Title: fmt.Sprintf("Superblock - %s", trimPartName(part.Name)),
	})

	return renderDOT(dotContent, outputPath)
}

// renderDOT guarda el contenido DOT y lo renderiza a imagen
func renderDOT(dotContent, outputPath string) error {
	// Determinar formato de salida
	ext := strings.ToLower(filepath.Ext(outputPath))
	format := "png"
	switch ext {
	case ".png":
		format = "png"
	case ".jpg", ".jpeg":
		format = "jpg"
	case ".svg":
		format = "svg"
	case ".pdf":
		format = "pdf"
	}

	// Guardar archivo DOT temporal
	dotPath := strings.TrimSuffix(outputPath, ext) + ".dot"
	if err := os.WriteFile(dotPath, []byte(dotContent), 0644); err != nil {
		return fmt.Errorf("error al guardar DOT: %v", err)
	}

	// Renderizar con Graphviz
	cmd := exec.Command("dot", "-T"+format, dotPath, "-o", outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error al renderizar con Graphviz: %v\nOutput: %s", err, output)
	}

	// Eliminar archivo DOT temporal (opcional)
	// os.Remove(dotPath)

	return nil
}

// Funciones helper
func fitToString(fit byte) string {
	switch fit {
	case disk.FitFF:
		return "FF"
	case disk.FitBF:
		return "BF"
	case disk.FitWF:
		return "WF"
	default:
		return "?"
	}
}

func statusToString(status byte) string {
	switch status {
	case disk.PartStatusUsed:
		return "used"
	case disk.PartStatusFree:
		return "free"
	default:
		return "unknown"
	}
}

func trimPartName(n [16]byte) string {
	for i, b := range n {
		if b == 0 {
			return string(n[:i])
		}
	}
	return string(n[:])
}

func getMountedPartitionInfo(diskPath, partName string) (*disk.Partition, int64, error) {
	f, err := os.Open(diskPath)
	if err != nil {
		return nil, 0, fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	var mbr disk.MBR
	if err := disk.ReadStruct(f, 0, &mbr); err != nil {
		return nil, 0, fmt.Errorf("error al leer MBR: %v", err)
	}

	// Buscar en primarias
	for i := 0; i < disk.MaxPrimaries; i++ {
		p := &mbr.Parts[i]
		if p.Status == disk.PartStatusUsed && trimPartName(p.Name) == partName {
			return p, p.Start, nil
		}
	}

	return nil, 0, fmt.Errorf("partición %s no encontrada", partName)
}

func binary_sizeof_mbr() int {
	return 8 + 8 + 4 + 1 + 7 + 4*(1+1+1+8+8+16+8)
}
