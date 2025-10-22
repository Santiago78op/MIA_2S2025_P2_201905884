package runner

import (
	"fmt"
	"strings"

	"Backend/command/disk"
	"Backend/command/fs"
	"Backend/command/reports"
	"Backend/command/users"
	"Backend/pkg/logger"
	"Backend/utils"
)

// Runner ejecuta comandos individuales y scripts completos
type Runner struct {
	diskSvc   *disk.DiskService
	fsSvc     *fs.FsService
	usersSvc  *users.UserService
	reportSvc *reports.ReportService
	log       logger.Logger
}

// New crea un nuevo runner con todos los servicios necesarios
func New(
	diskSvc *disk.DiskService,
	fsSvc *fs.FsService,
	usersSvc *users.UserService,
	reportSvc *reports.ReportService,
	log logger.Logger,
) *Runner {
	return &Runner{
		diskSvc:   diskSvc,
		fsSvc:     fsSvc,
		usersSvc:  usersSvc,
		reportSvc: reportSvc,
		log:       log,
	}
}

// Run ejecuta un comando individual (implementa CommandRunner)
func (r *Runner) Run(line string) (string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", fmt.Errorf("comando vacío")
	}

	// Parsear el comando: extraer el nombre y los argumentos
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", fmt.Errorf("comando vacío")
	}

	cmd := strings.ToLower(parts[0])
	r.log.Debug("Ejecutando comando: %s", cmd)

	switch cmd {
	// === Comandos de disco ===
	case "mkdisk":
		args, err := disk.ParseMkDisk(line)
		if err != nil {
			return "", err
		}
		return r.diskSvc.MkDisk(args)

	case "rmdisk":
		args, err := disk.ParseRmDisk(line)
		if err != nil {
			return "", err
		}
		return r.diskSvc.RmDisk(args)

	case "fdisk":
		// Detectar si es add, delete o creación normal
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "-add=") {
			args, err := disk.ParseFDiskAdd(line)
			if err != nil {
				return "", err
			}
			return r.diskSvc.FDiskAdd(args)
		} else if strings.Contains(lineLower, "-delete=") {
			args, err := disk.ParseFDiskDelete(line)
			if err != nil {
				return "", err
			}
			return r.diskSvc.FDiskDelete(args)
		} else {
			// Creación normal
			args, err := disk.ParseFDisk(line)
			if err != nil {
				return "", err
			}
			return r.diskSvc.FDisk(args)
		}

	case "mount":
		args, err := disk.ParseMount(line)
		if err != nil {
			return "", err
		}
		return r.diskSvc.Mount(args)

	case "mounted":
		entries := r.diskSvc.Mounted()
		if len(entries) == 0 {
			return "No hay particiones montadas", nil
		}
		var sb strings.Builder
		sb.WriteString("Particiones montadas:\n")
		for _, e := range entries {
			sb.WriteString(fmt.Sprintf("  ID: %s  Path: %s  Name: %s\n", e.ID, e.Path, e.Name))
		}
		return sb.String(), nil

	// === Comandos de filesystem ===
	case "mkfs":
		args, err := fs.ParseMkfs(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Mkfs(args.ID, args.Fs, args.Type)

	case "mkdir":
		args, err := fs.ParseMkdir(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Mkdir(args.ID, args.Path, args.Parents)

	case "mkfile":
		args, err := fs.ParseMkfile(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Mkfile(args.ID, args.Path, args.Size, args.Cont, args.Recursive)

	case "cat":
		args, err := fs.ParseCat(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Cat(args.ID, args.Files)

	// === Comandos de usuarios ===
	case "login":
		args, err := users.ParseLogin(line)
		if err != nil {
			return "", err
		}
		return r.usersSvc.Login(args.ID, args.User, args.Pass)

	case "logout":
		return r.usersSvc.Logout()

	case "mkgrp":
		args, err := users.ParseMkgrp(line)
		if err != nil {
			return "", err
		}
		return r.usersSvc.Mkgrp(args.ID, args.Name)

	case "rmgrp":
		args, err := users.ParseRmgrp(line)
		if err != nil {
			return "", err
		}
		return r.usersSvc.Rmgrp(args.ID, args.Name)

	case "mkusr":
		args, err := users.ParseMkusr(line)
		if err != nil {
			return "", err
		}
		return r.usersSvc.Mkusr(args.ID, args.User, args.Pass, args.Grp)

	case "rmusr":
		args, err := users.ParseRmusr(line)
		if err != nil {
			return "", err
		}
		return r.usersSvc.Rmusr(args.ID, args.User)

	case "chgrp":
		args, err := users.ParseChgrp(line)
		if err != nil {
			return "", err
		}
		return r.usersSvc.Chgrp(args.ID, args.User, args.Grp)

	// === Comandos de reportes ===
	case "rep":
		args, err := reports.ParseReport(line)
		if err != nil {
			return "", err
		}
		path, err := r.reportSvc.Generate(args.Name, args.ID, args.Out, args.Extra)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Reporte generado: %s", path), nil

	// === Comandos nuevos P2 - Filesystem ===
	case "remove":
		args, err := fs.ParseRemove(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Remove(args.ID, args.Path)

	case "edit":
		args, err := fs.ParseEdit(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Edit(args.ID, args.Path, args.Content)

	case "rename":
		args, err := fs.ParseRename(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Rename(args.ID, args.Path, args.Name)

	case "copy":
		args, err := fs.ParseCopy(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Copy(args.ID, args.SrcPath, args.DestPath)

	case "move":
		args, err := fs.ParseMove(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Move(args.ID, args.SrcPath, args.DestPath)

	case "find":
		args, err := fs.ParseFind(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Find(args.ID, args.Path, args.Name)

	case "chmod":
		args, err := fs.ParseChmod(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Chmod(args.ID, args.Path, args.Ugo, args.Recursive)

	case "chown":
		args, err := fs.ParseChown(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Chown(args.ID, args.Path, args.User, args.Recursive)

	case "loss":
		args, err := fs.ParseLoss(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Loss(args.ID)

	case "recovery":
		args, err := fs.ParseRecovery(line)
		if err != nil {
			return "", err
		}
		return r.fsSvc.Recovery(args.ID)

	// === Comandos nuevos P2 - Disk ===
	case "unmount":
		args, err := disk.ParseUnmount(line)
		if err != nil {
			return "", err
		}
		return r.diskSvc.Unmount(args.ID)

	case "unmountall":
		return r.diskSvc.UnmountAll()

	default:
		return "", fmt.Errorf("comando desconocido: %s", cmd)
	}
}

// RunScript ejecuta un script .smia completo (implementa ScriptRunner)
func (r *Runner) RunScript(script string, stopOnError bool) (string, error) {
	lines := utils.ParseSMIA(script)

	var results []string
	var errors []string

	for i, line := range lines {
		r.log.Debug("Ejecutando línea %d: %s", i+1, line)

		out, err := r.Run(line)
		if err != nil {
			errMsg := fmt.Sprintf("Error en línea %d (%s): %v", i+1, line, err)
			errors = append(errors, errMsg)
			r.log.Error(errMsg)

			if stopOnError {
				return strings.Join(results, "\n"), fmt.Errorf(errMsg)
			}
			results = append(results, errMsg)
		} else {
			results = append(results, fmt.Sprintf("Línea %d: %s", i+1, out))
		}
	}

	output := strings.Join(results, "\n")

	if len(errors) > 0 && !stopOnError {
		output += fmt.Sprintf("\n\n=== Errores encontrados: %d ===\n%s",
			len(errors), strings.Join(errors, "\n"))
	}

	return output, nil
}
