package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"MIA_2S2025_P2_201905884/internal/fs"
)

// ==================== Handlers de Disco ====================

func (c *MkdiskCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	sizeBytes, err := toBytes(c.Size, c.Unit)
	if err != nil {
		return "", err
	}

	if err := adapter.DM.Mkdisk(ctx, c.Path, sizeBytes, c.Fit); err != nil {
		return "", err
	}

	return fmt.Sprintf("mkdisk OK path=%s size=%d%s fit=%s",
		c.Path, c.Size, c.Unit, c.Fit), nil
}

func (c *FdiskCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	mode := strings.ToLower(c.Mode)

	switch mode {
	case "add":
		sizeBytes, err := toBytes(c.Size, c.Unit)
		if err != nil {
			return "", err
		}

		if err := adapter.DM.FdiskAdd(ctx, c.Path, c.PartName, sizeBytes, c.Type, c.Fit); err != nil {
			return "", err
		}

		return fmt.Sprintf("fdisk add OK path=%s name=%s size=%d%s type=%s fit=%s",
			c.Path, c.PartName, c.Size, c.Unit, c.Type, c.Fit), nil

	case "delete":
		delMode := strings.ToLower(c.Delete)
		if delMode == "" {
			delMode = "fast"
		}

		if err := adapter.DM.FdiskDelete(ctx, c.Path, c.PartName, delMode); err != nil {
			return "", err
		}

		return fmt.Sprintf("fdisk delete OK path=%s name=%s mode=%s",
			c.Path, c.PartName, delMode), nil

	default:
		return "", fmt.Errorf("fdisk: mode debe ser add|delete")
	}
}

func (c *MountCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	ref, err := adapter.DM.Mount(ctx, c.Path, c.PartName)
	if err != nil {
		return "", err
	}

	// Generar ID estable
	id := MakeID(ref.DiskPath, ref.PartitionID)
	h := fs.MountHandle{
		DiskID:      ref.DiskPath,
		PartitionID: ref.PartitionID,
	}
	adapter.Index.Put(id, ref, h)

	return fmt.Sprintf("mount OK id=%s path=%s name=%s", id, c.Path, c.PartName), nil
}

func (c *UnmountCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	ref, ok := adapter.Index.GetRef(c.ID)
	if !ok {
		return "", fmt.Errorf("unmount: id no encontrado: %s", c.ID)
	}

	h, okHandle := adapter.Index.GetHandle(c.ID)
	if okHandle {
		// Ejecutar unmount en el FS correspondiente para limpieza
		_ = adapter.pickFS(h).Unmount(ctx, h)
	}

	if err := adapter.DM.Unmount(ctx, ref); err != nil {
		return "", err
	}

	// Eliminar completamente del índice
	adapter.Index.Del(c.ID)

	return fmt.Sprintf("unmount OK id=%s", c.ID), nil
}

// ==================== Handlers de Formateo ====================

func (c *MkfsCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	_, ok := adapter.Index.GetRef(c.ID)
	if !ok {
		return "", fmt.Errorf("mkfs: id no encontrado: %s", c.ID)
	}

	kind := strings.ToLower(c.FSKind)
	req := fs.MkfsRequest{MountID: c.ID, FSKind: kind}

	switch kind {
	case "2fs":
		if err := adapter.FS2.Mkfs(ctx, req); err != nil {
			return "", err
		}
	case "3fs":
		if err := adapter.FS3.Mkfs(ctx, req); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("mkfs: fs desconocido: %s (usa 2fs|3fs)", kind)
	}

	return fmt.Sprintf("mkfs OK id=%s fs=%s", c.ID, kind), nil
}

// ==================== Handlers de Árbol/Archivos ====================

func (c *MkdirCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("mkdir: id no encontrado: %s", c.ID)
	}

	req := fs.MkdirRequest{
		Path: c.Path,
		Deep: c.Deep,
	}

	if err := adapter.pickFS(h).Mkdir(ctx, h, req); err != nil {
		return "", err
	}

	return fmt.Sprintf("mkdir OK id=%s path=%s", c.ID, c.Path), nil
}

func (c *MkfileCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("mkfile: id no encontrado: %s", c.ID)
	}

	content := []byte(c.Content)

	// Si se especifica size, generar contenido del tamaño indicado
	if c.Size > 0 && len(content) == 0 {
		content = make([]byte, c.Size)
		for i := range content {
			content[i] = byte('0' + (i % 10))
		}
	}

	req := fs.WriteFileRequest{
		Path:    c.Path,
		Content: content,
		Append:  false,
	}

	if err := adapter.pickFS(h).WriteFile(ctx, h, req); err != nil {
		return "", err
	}

	return fmt.Sprintf("mkfile OK id=%s path=%s", c.ID, c.Path), nil
}

func (c *RemoveCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("remove: id no encontrado: %s", c.ID)
	}

	if err := adapter.pickFS(h).Remove(ctx, h, c.Path); err != nil {
		return "", err
	}

	return fmt.Sprintf("remove OK id=%s path=%s", c.ID, c.Path), nil
}

func (c *EditCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("edit: id no encontrado: %s", c.ID)
	}

	req := fs.WriteFileRequest{
		Path:    c.Path,
		Content: []byte(c.Content),
		Append:  c.Append,
	}

	if err := adapter.pickFS(h).WriteFile(ctx, h, req); err != nil {
		return "", err
	}

	return fmt.Sprintf("edit OK id=%s path=%s", c.ID, c.Path), nil
}

func (c *RenameCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("rename: id no encontrado: %s", c.ID)
	}

	if err := adapter.pickFS(h).Rename(ctx, h, c.From, c.To); err != nil {
		return "", err
	}

	return fmt.Sprintf("rename OK id=%s from=%s to=%s", c.ID, c.From, c.To), nil
}

func (c *CopyCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("copy: id no encontrado: %s", c.ID)
	}

	if err := adapter.pickFS(h).Copy(ctx, h, c.From, c.To); err != nil {
		return "", err
	}

	return fmt.Sprintf("copy OK id=%s from=%s to=%s", c.ID, c.From, c.To), nil
}

func (c *MoveCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("move: id no encontrado: %s", c.ID)
	}

	if err := adapter.pickFS(h).Move(ctx, h, c.From, c.To); err != nil {
		return "", err
	}

	return fmt.Sprintf("move OK id=%s from=%s to=%s", c.ID, c.From, c.To), nil
}

func (c *FindCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("find: id no encontrado: %s", c.ID)
	}

	req := fs.FindRequest{
		BasePath: c.Base,
		Pattern:  c.Pattern,
		Limit:    c.Limit,
	}

	list, err := adapter.pickFS(h).Find(ctx, h, req)
	if err != nil {
		return "", err
	}

	return strings.Join(list, "\n"), nil
}

func (c *ChownCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("chown: id no encontrado: %s", c.ID)
	}

	if err := adapter.pickFS(h).Chown(ctx, h, c.Path, c.User, c.Group); err != nil {
		return "", err
	}

	return fmt.Sprintf("chown OK id=%s path=%s user=%s group=%s",
		c.ID, c.Path, c.User, c.Group), nil
}

func (c *ChmodCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("chmod: id no encontrado: %s", c.ID)
	}

	// Parsear permisos octales
	perm, err := parsePermissions(c.Perm)
	if err != nil {
		return "", fmt.Errorf("chmod: permisos inválidos: %v", err)
	}

	if err := adapter.pickFS(h).Chmod(ctx, h, c.Path, perm); err != nil {
		return "", err
	}

	return fmt.Sprintf("chmod OK id=%s path=%s perm=%s", c.ID, c.Path, c.Perm), nil
}

// ==================== Handlers EXT3 ====================

func (c *JournalingCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("journaling: id no encontrado: %s", c.ID)
	}

	entries, err := adapter.FS3.Journaling(ctx, h)
	if err != nil {
		return "", err
	}

	b, _ := json.MarshalIndent(entries, "", "  ")
	return string(b), nil
}

func (c *RecoveryCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("recovery: id no encontrado: %s", c.ID)
	}

	if err := adapter.FS3.Recovery(ctx, h); err != nil {
		return "", err
	}

	return fmt.Sprintf("recovery OK id=%s", c.ID), nil
}

func (c *LossCommand) Execute(ctx context.Context, adapter *Adapter) (string, error) {
	h, ok := adapter.Index.GetHandle(c.ID)
	if !ok {
		return "", fmt.Errorf("loss: id no encontrado: %s", c.ID)
	}

	if err := adapter.FS3.Loss(ctx, h); err != nil {
		return "", err
	}

	return fmt.Sprintf("loss OK id=%s", c.ID), nil
}

// ==================== Helper Functions ====================

// toBytes convierte size + unit (b|k|m) a bytes
func toBytes(size int64, unit string) (int64, error) {
	switch strings.ToLower(unit) {
	case "", "b":
		return size, nil
	case "k":
		return size * 1024, nil
	case "m":
		return size * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unidad inválida: %s (usa b|k|m)", unit)
	}
}

// parsePermissions parsea una cadena de permisos octales (ej: "755")
func parsePermissions(perm string) (uint16, error) {
	if perm == "" {
		return 0, fmt.Errorf("permisos vacíos")
	}

	val, err := strconv.ParseUint(perm, 8, 16)
	if err != nil {
		return 0, fmt.Errorf("formato de permisos inválido: %s", perm)
	}

	return uint16(val), nil
}
