package commands

import (
	"context"
	"fmt"
	"strings"

	"MIA_2S2025_P2_201905884/pkg/reports"
)

// RepCommand representa el comando REP para generar reportes
type RepCommand struct {
	ReportName string // Tipo de reporte (mbr, disk, inode, etc.)
	Path       string // Ruta donde guardar el reporte
	ID         string // ID de partición montada
	Ruta       string // Ruta de archivo (para file y ls)
}

func (c *RepCommand) Name() CommandName {
	return CmdRep
}

func (c *RepCommand) Validate() error {
	if c.ReportName == "" {
		return fmt.Errorf("rep: parámetro 'name' es obligatorio")
	}
	if c.Path == "" {
		return fmt.Errorf("rep: parámetro 'path' es obligatorio")
	}
	if c.ID == "" {
		return fmt.Errorf("rep: parámetro 'id' es obligatorio")
	}

	// Validar tipo de reporte
	validTypes := []string{"mbr", "disk", "inode", "block", "bm_inode", "bm_block", "tree", "sb", "file"}
	valid := false
	nameLower := strings.ToLower(c.ReportName)
	for _, t := range validTypes {
		if nameLower == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("rep: tipo de reporte '%s' no válido. Tipos: %v", c.ReportName, validTypes)
	}

	// Validar que file requiere ruta
	if (nameLower == "file" || nameLower == "ls") && c.Ruta == "" {
		return fmt.Errorf("rep: reporte '%s' requiere parámetro 'ruta'", c.ReportName)
	}

	return nil
}

func (c *RepCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	// Obtener información de la partición montada
	ref, ok := adapter.Index.GetByID(c.ID)
	if !ok {
		return "", fmt.Errorf("partición con ID '%s' no está montada", c.ID)
	}

	diskPath := ref.DiskPath
	partName := ref.PartitionID

	// Ejecutar según tipo de reporte
	nameLower := strings.ToLower(c.ReportName)
	var err error

	switch nameLower {
	case "mbr":
		err = reports.GenerateMBRReport(diskPath, c.Path)
	case "disk":
		err = reports.GenerateDISKReport(diskPath, c.Path)
	case "sb", "superblock":
		err = reports.GenerateSuperblockReport(diskPath, partName, c.Path)
	// case "inode":
	// 	err = reports.GenerateINODEReport(diskPath, partName, c.Path)
	// case "block":
	// 	err = reports.GenerateBLOCKReport(diskPath, partName, c.Path)
	// case "bm_inode":
	// 	err = reports.GenerateBMINODEReport(diskPath, partName, c.Path)
	// case "bm_block":
	// 	err = reports.GenerateBMBLOCKReport(diskPath, partName, c.Path)
	// case "tree":
	// 	err = reports.GenerateTREEReport(diskPath, partName, c.Path)
	// case "file":
	// 	err = reports.GenerateFILEReport(diskPath, partName, c.Ruta, c.Path)
	default:
		return "", fmt.Errorf("reporte '%s' aún no implementado", c.ReportName)
	}

	if err != nil {
		return "", fmt.Errorf("error al generar reporte %s: %v", c.ReportName, err)
	}

	return fmt.Sprintf("rep OK: Reporte %s generado en %s", c.ReportName, c.Path), nil
}

// parseRep parsea los argumentos del comando REP
func parseRep(args map[string]string) (CommandHandler, error) {
	cmd := &RepCommand{
		ReportName: args["name"],
		Path:       args["path"],
		ID:         args["id"],
		Ruta:       args["ruta"], // También puede ser path_file_ls
	}

	// Compatibilidad con path_file_ls
	if cmd.Ruta == "" {
		cmd.Ruta = args["path_file_ls"]
	}

	return cmd, nil
}
