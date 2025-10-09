package commands

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseCommand parsea una línea de comando y retorna el handler apropiado
func ParseCommand(line string) (CommandHandler, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("línea vacía")
	}

	// Parsear nombre y argumentos
	cmd, args, err := parseLineToArgs(line)
	if err != nil {
		return nil, err
	}

	cmdName := CommandName(strings.ToLower(cmd))

	// Crear el comando apropiado basado en el nombre
	switch cmdName {
	case CmdMkdisk:
		return parseMkdisk(args)
	case CmdFdisk:
		return parseFdisk(args)
	case CmdMount:
		return parseMount(args)
	case CmdUnmount:
		return parseUnmount(args)
	case CmdMounted:
		return parseMounted(args)
	case CmdMkfs:
		return parseMkfs(args)
	case CmdMkdir:
		return parseMkdir(args)
	case CmdMkfile:
		return parseMkfile(args)
	case CmdRemove:
		return parseRemove(args)
	case CmdEdit:
		return parseEdit(args)
	case CmdRename:
		return parseRename(args)
	case CmdCopy:
		return parseCopy(args)
	case CmdMove:
		return parseMove(args)
	case CmdFind:
		return parseFind(args)
	case CmdChown:
		return parseChown(args)
	case CmdChmod:
		return parseChmod(args)
	case CmdJournaling:
		return parseJournaling(args)
	case CmdRecovery:
		return parseRecovery(args)
	case CmdLoss:
		return parseLoss(args)
	case CmdRep:
		return parseRep(args)
	default:
		return nil, fmt.Errorf("comando desconocido: %s", cmd)
	}
}

// parseLineToArgs parsea una línea en nombre de comando y mapa de argumentos
func parseLineToArgs(line string) (string, map[string]string, error) {
	parts := tokenize(line)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("no se pudo parsear el comando")
	}

	cmdName := parts[0]
	args := make(map[string]string)

	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if !strings.HasPrefix(part, "-") {
			continue
		}

		// Remover el prefijo '-'
		key := strings.TrimPrefix(part, "-")

		// Verificar si es un flag booleano o tiene valor
		if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
			args[key] = parts[i+1]
			i++
		} else {
			// Flag booleano
			args[key] = "true"
		}
	}

	return cmdName, args, nil
}

// tokenize divide una línea respetando comillas
func tokenize(line string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		char := line[i]

		switch char {
		case '"':
			inQuotes = !inQuotes
		case ' ', '\t':
			if inQuotes {
				current.WriteByte(char)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// Funciones helper para obtener argumentos
func getStringArg(args map[string]string, key string, defaultVal string) string {
	if val, ok := args[key]; ok {
		return val
	}
	return defaultVal
}

func getInt64Arg(args map[string]string, key string, defaultVal int64) int64 {
	if val, ok := args[key]; ok {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return defaultVal
}

func getBoolArg(args map[string]string, key string) bool {
	if val, ok := args[key]; ok {
		return val == "true" || val == "1"
	}
	return false
}

// ==================== Parsers específicos ====================

func parseMkdisk(args map[string]string) (*MkdiskCommand, error) {
	return &MkdiskCommand{
		BaseCommand: BaseCommand{CmdName: CmdMkdisk},
		Path:        getStringArg(args, "path", ""),
		Size:        getInt64Arg(args, "size", 0),
		Unit:        getStringArg(args, "unit", "m"),
		Fit:         getStringArg(args, "fit", "ff"),
	}, nil
}

func parseFdisk(args map[string]string) (*FdiskCommand, error) {
	return &FdiskCommand{
		BaseCommand: BaseCommand{CmdName: CmdFdisk},
		Path:        getStringArg(args, "path", ""),
		Mode:        getStringArg(args, "mode", ""),
		PartName:    getStringArg(args, "name", ""),
		Size:        getInt64Arg(args, "size", 0),
		Unit:        getStringArg(args, "unit", "m"),
		Type:        getStringArg(args, "type", "p"),
		Fit:         getStringArg(args, "fit", "ff"),
		Delete:      getStringArg(args, "delete", "fast"),
	}, nil
}

func parseMount(args map[string]string) (*MountCommand, error) {
	return &MountCommand{
		BaseCommand: BaseCommand{CmdName: CmdMount},
		Path:        getStringArg(args, "path", ""),
		PartName:    getStringArg(args, "name", ""),
	}, nil
}

func parseUnmount(args map[string]string) (*UnmountCommand, error) {
	return &UnmountCommand{
		BaseCommand: BaseCommand{CmdName: CmdUnmount},
		ID:          getStringArg(args, "id", ""),
	}, nil
}

func parseMounted(args map[string]string) (*MountedCommand, error) {
	return &MountedCommand{
		BaseCommand: BaseCommand{CmdName: CmdMounted},
	}, nil
}

func parseMkfs(args map[string]string) (*MkfsCommand, error) {
	return &MkfsCommand{
		BaseCommand: BaseCommand{CmdName: CmdMkfs},
		ID:          getStringArg(args, "id", ""),
		FSKind:      getStringArg(args, "fs", "2fs"),
	}, nil
}

func parseMkdir(args map[string]string) (*MkdirCommand, error) {
	return &MkdirCommand{
		BaseCommand: BaseCommand{CmdName: CmdMkdir},
		ID:          getStringArg(args, "id", ""),
		Path:        getStringArg(args, "path", ""),
		Deep:        getBoolArg(args, "p"),
	}, nil
}

func parseMkfile(args map[string]string) (*MkfileCommand, error) {
	return &MkfileCommand{
		BaseCommand: BaseCommand{CmdName: CmdMkfile},
		ID:          getStringArg(args, "id", ""),
		Path:        getStringArg(args, "path", ""),
		Content:     getStringArg(args, "cont", ""),
		Size:        getInt64Arg(args, "size", 0),
	}, nil
}

func parseRemove(args map[string]string) (*RemoveCommand, error) {
	return &RemoveCommand{
		BaseCommand: BaseCommand{CmdName: CmdRemove},
		ID:          getStringArg(args, "id", ""),
		Path:        getStringArg(args, "path", ""),
	}, nil
}

func parseEdit(args map[string]string) (*EditCommand, error) {
	return &EditCommand{
		BaseCommand: BaseCommand{CmdName: CmdEdit},
		ID:          getStringArg(args, "id", ""),
		Path:        getStringArg(args, "path", ""),
		Content:     getStringArg(args, "cont", ""),
		Append:      getBoolArg(args, "append"),
	}, nil
}

func parseRename(args map[string]string) (*RenameCommand, error) {
	return &RenameCommand{
		BaseCommand: BaseCommand{CmdName: CmdRename},
		ID:          getStringArg(args, "id", ""),
		From:        getStringArg(args, "from", ""),
		To:          getStringArg(args, "to", ""),
	}, nil
}

func parseCopy(args map[string]string) (*CopyCommand, error) {
	return &CopyCommand{
		BaseCommand: BaseCommand{CmdName: CmdCopy},
		ID:          getStringArg(args, "id", ""),
		From:        getStringArg(args, "from", ""),
		To:          getStringArg(args, "to", ""),
	}, nil
}

func parseMove(args map[string]string) (*MoveCommand, error) {
	return &MoveCommand{
		BaseCommand: BaseCommand{CmdName: CmdMove},
		ID:          getStringArg(args, "id", ""),
		From:        getStringArg(args, "from", ""),
		To:          getStringArg(args, "to", ""),
	}, nil
}

func parseFind(args map[string]string) (*FindCommand, error) {
	return &FindCommand{
		BaseCommand: BaseCommand{CmdName: CmdFind},
		ID:          getStringArg(args, "id", ""),
		Base:        getStringArg(args, "base", "/"),
		Pattern:     getStringArg(args, "name", ""),
		Limit:       int(getInt64Arg(args, "limit", 100)),
	}, nil
}

func parseChown(args map[string]string) (*ChownCommand, error) {
	return &ChownCommand{
		BaseCommand: BaseCommand{CmdName: CmdChown},
		ID:          getStringArg(args, "id", ""),
		Path:        getStringArg(args, "path", ""),
		User:        getStringArg(args, "user", ""),
		Group:       getStringArg(args, "group", ""),
	}, nil
}

func parseChmod(args map[string]string) (*ChmodCommand, error) {
	return &ChmodCommand{
		BaseCommand: BaseCommand{CmdName: CmdChmod},
		ID:          getStringArg(args, "id", ""),
		Path:        getStringArg(args, "path", ""),
		Perm:        getStringArg(args, "perm", ""),
	}, nil
}

func parseJournaling(args map[string]string) (*JournalingCommand, error) {
	return &JournalingCommand{
		BaseCommand: BaseCommand{CmdName: CmdJournaling},
		ID:          getStringArg(args, "id", ""),
	}, nil
}

func parseRecovery(args map[string]string) (*RecoveryCommand, error) {
	return &RecoveryCommand{
		BaseCommand: BaseCommand{CmdName: CmdRecovery},
		ID:          getStringArg(args, "id", ""),
	}, nil
}

func parseLoss(args map[string]string) (*LossCommand, error) {
	return &LossCommand{
		BaseCommand: BaseCommand{CmdName: CmdLoss},
		ID:          getStringArg(args, "id", ""),
	}, nil
}

// Usage retorna el mensaje de uso para un comando
func Usage(cmdName CommandName) string {
	usageMap := map[CommandName]string{
		CmdMkdisk:     "mkdisk -path <ruta> -size <tamaño> [-unit b|k|m] [-fit bf|ff|wf]",
		CmdFdisk:      "fdisk -path <ruta> -mode add|delete [-name <nombre>] [-size <tamaño>] [-unit b|k|m] [-type p|e|l] [-fit bf|ff|wf] [-delete full|fast]",
		CmdMount:      "mount -path <ruta> -name <nombre>",
		CmdUnmount:    "unmount -id <id>",
		CmdMounted:    "mounted",
		CmdMkfs:       "mkfs -id <id> -fs 2fs|3fs",
		CmdMkdir:      "mkdir -id <id> -path <ruta> [-p]",
		CmdMkfile:     "mkfile -id <id> -path <ruta> [-cont <contenido>] [-size <tamaño>]",
		CmdRemove:     "remove -id <id> -path <ruta>",
		CmdEdit:       "edit -id <id> -path <ruta> -cont <contenido> [-append]",
		CmdRename:     "rename -id <id> -from <origen> -to <destino>",
		CmdCopy:       "copy -id <id> -from <origen> -to <destino>",
		CmdMove:       "move -id <id> -from <origen> -to <destino>",
		CmdFind:       "find -id <id> [-base <ruta>] [-name <patrón>] [-limit <n>]",
		CmdChown:      "chown -id <id> -path <ruta> -user <usuario> -group <grupo>",
		CmdChmod:      "chmod -id <id> -path <ruta> -perm <permisos>",
		CmdJournaling: "journaling -id <id>",
		CmdRecovery:   "recovery -id <id>",
		CmdLoss:       "loss -id <id>",
	}

	if usage, ok := usageMap[cmdName]; ok {
		return fmt.Sprintf("Uso: %s", usage)
	}
	return "Comando desconocido"
}
