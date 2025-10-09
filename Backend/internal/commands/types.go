package commands

import (
	"context"
	"fmt"
	"strings"
)

// CommandName representa el nombre de un comando
type CommandName string

const (
	// Comandos de disco/particiones
	CmdMkdisk  CommandName = "mkdisk"
	CmdFdisk   CommandName = "fdisk"
	CmdMount   CommandName = "mount"
	CmdUnmount CommandName = "unmount"
	CmdMounted CommandName = "mounted"

	// Comandos de formateo
	CmdMkfs CommandName = "mkfs"

	// Comandos de árbol/archivos
	CmdMkdir  CommandName = "mkdir"
	CmdMkfile CommandName = "mkfile"
	CmdRemove CommandName = "remove"
	CmdEdit   CommandName = "edit"
	CmdRename CommandName = "rename"
	CmdCopy   CommandName = "copy"
	CmdMove   CommandName = "move"
	CmdFind   CommandName = "find"
	CmdChown  CommandName = "chown"
	CmdChmod  CommandName = "chmod"

	// Comandos EXT3 específicos
	CmdJournaling CommandName = "journaling"
	CmdRecovery   CommandName = "recovery"
	CmdLoss       CommandName = "loss"

	// Comandos de reportes
	CmdRep CommandName = "rep"
)

// CommandHandler es la interfaz que implementan todos los handlers de comandos
type CommandHandler interface {
	Execute(ctx context.Context, adapter *Adapter) (string, error)
	Validate() error
	Name() CommandName
}

// BaseCommand contiene campos comunes para todos los comandos
type BaseCommand struct {
	CmdName CommandName
}

func (c *BaseCommand) Name() CommandName {
	return c.CmdName
}

// ==================== Comandos de Disco ====================

// MkdiskCommand representa el comando mkdisk
type MkdiskCommand struct {
	BaseCommand
	Path string
	Size int64
	Unit string // b|k|m
	Fit  string // bf|ff|wf
}

func (c *MkdiskCommand) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("mkdisk: falta parámetro 'path'")
	}
	if c.Size <= 0 {
		return fmt.Errorf("mkdisk: 'size' debe ser > 0")
	}
	unit := strings.ToLower(c.Unit)
	if unit != "" && unit != "b" && unit != "k" && unit != "m" {
		return fmt.Errorf("mkdisk: 'unit' debe ser b|k|m")
	}
	fit := strings.ToLower(c.Fit)
	if fit != "" && fit != "bf" && fit != "ff" && fit != "wf" {
		return fmt.Errorf("mkdisk: 'fit' debe ser bf|ff|wf")
	}
	return nil
}

// FdiskCommand representa el comando fdisk
type FdiskCommand struct {
	BaseCommand
	Path     string
	Mode     string // add|delete
	PartName string
	Size     int64
	Unit     string // b|k|m
	Type     string // p|e|l
	Fit      string // bf|ff|wf
	Delete   string // full|fast
}

func (c *FdiskCommand) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("fdisk: falta parámetro 'path'")
	}
	mode := strings.ToLower(c.Mode)
	if mode != "add" && mode != "delete" {
		return fmt.Errorf("fdisk: 'mode' debe ser add|delete")
	}
	if mode == "add" {
		if c.PartName == "" {
			return fmt.Errorf("fdisk add: falta parámetro 'name'")
		}
		if c.Size <= 0 {
			return fmt.Errorf("fdisk add: 'size' debe ser > 0")
		}
		ptype := strings.ToLower(c.Type)
		if ptype != "p" && ptype != "e" && ptype != "l" {
			return fmt.Errorf("fdisk add: 'type' debe ser p|e|l")
		}
	}
	if mode == "delete" {
		if c.PartName == "" {
			return fmt.Errorf("fdisk delete: falta parámetro 'name'")
		}
		delMode := strings.ToLower(c.Delete)
		if delMode != "" && delMode != "full" && delMode != "fast" {
			return fmt.Errorf("fdisk delete: 'delete' debe ser full|fast")
		}
	}
	return nil
}

// MountCommand representa el comando mount
type MountCommand struct {
	BaseCommand
	Path     string
	PartName string
}

func (c *MountCommand) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("mount: falta parámetro 'path'")
	}
	if c.PartName == "" {
		return fmt.Errorf("mount: falta parámetro 'name'")
	}
	return nil
}

// UnmountCommand representa el comando unmount
type UnmountCommand struct {
	BaseCommand
	ID string
}

func (c *UnmountCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("unmount: falta parámetro 'id'")
	}
	return nil
}

// MountedCommand representa el comando mounted
type MountedCommand struct {
	BaseCommand
}

func (c *MountedCommand) Validate() error {
	// No requiere parámetros
	return nil
}

// ==================== Comandos de Formateo ====================

// MkfsCommand representa el comando mkfs
type MkfsCommand struct {
	BaseCommand
	ID     string
	FSKind string // 2fs|3fs
}

func (c *MkfsCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("mkfs: falta parámetro 'id'")
	}
	kind := strings.ToLower(c.FSKind)
	if kind != "2fs" && kind != "3fs" {
		return fmt.Errorf("mkfs: 'fs' debe ser 2fs|3fs")
	}
	return nil
}

// ==================== Comandos de Árbol/Archivos ====================

// MkdirCommand representa el comando mkdir
type MkdirCommand struct {
	BaseCommand
	ID   string
	Path string
	Deep bool // flag -p
}

func (c *MkdirCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("mkdir: falta parámetro 'id'")
	}
	if c.Path == "" {
		return fmt.Errorf("mkdir: falta parámetro 'path'")
	}
	return nil
}

// MkfileCommand representa el comando mkfile
type MkfileCommand struct {
	BaseCommand
	ID      string
	Path    string
	Content string
	Size    int64
}

func (c *MkfileCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("mkfile: falta parámetro 'id'")
	}
	if c.Path == "" {
		return fmt.Errorf("mkfile: falta parámetro 'path'")
	}
	return nil
}

// RemoveCommand representa el comando remove
type RemoveCommand struct {
	BaseCommand
	ID   string
	Path string
}

func (c *RemoveCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("remove: falta parámetro 'id'")
	}
	if c.Path == "" {
		return fmt.Errorf("remove: falta parámetro 'path'")
	}
	return nil
}

// EditCommand representa el comando edit
type EditCommand struct {
	BaseCommand
	ID      string
	Path    string
	Content string
	Append  bool
}

func (c *EditCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("edit: falta parámetro 'id'")
	}
	if c.Path == "" {
		return fmt.Errorf("edit: falta parámetro 'path'")
	}
	return nil
}

// RenameCommand representa el comando rename
type RenameCommand struct {
	BaseCommand
	ID   string
	From string
	To   string
}

func (c *RenameCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("rename: falta parámetro 'id'")
	}
	if c.From == "" {
		return fmt.Errorf("rename: falta parámetro 'from'")
	}
	if c.To == "" {
		return fmt.Errorf("rename: falta parámetro 'to'")
	}
	return nil
}

// CopyCommand representa el comando copy
type CopyCommand struct {
	BaseCommand
	ID   string
	From string
	To   string
}

func (c *CopyCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("copy: falta parámetro 'id'")
	}
	if c.From == "" {
		return fmt.Errorf("copy: falta parámetro 'from'")
	}
	if c.To == "" {
		return fmt.Errorf("copy: falta parámetro 'to'")
	}
	return nil
}

// MoveCommand representa el comando move
type MoveCommand struct {
	BaseCommand
	ID   string
	From string
	To   string
}

func (c *MoveCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("move: falta parámetro 'id'")
	}
	if c.From == "" {
		return fmt.Errorf("move: falta parámetro 'from'")
	}
	if c.To == "" {
		return fmt.Errorf("move: falta parámetro 'to'")
	}
	return nil
}

// FindCommand representa el comando find
type FindCommand struct {
	BaseCommand
	ID      string
	Base    string
	Pattern string
	Limit   int
}

func (c *FindCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("find: falta parámetro 'id'")
	}
	return nil
}

// ChownCommand representa el comando chown
type ChownCommand struct {
	BaseCommand
	ID    string
	Path  string
	User  string
	Group string
}

func (c *ChownCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("chown: falta parámetro 'id'")
	}
	if c.Path == "" {
		return fmt.Errorf("chown: falta parámetro 'path'")
	}
	return nil
}

// ChmodCommand representa el comando chmod
type ChmodCommand struct {
	BaseCommand
	ID   string
	Path string
	Perm string
}

func (c *ChmodCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("chmod: falta parámetro 'id'")
	}
	if c.Path == "" {
		return fmt.Errorf("chmod: falta parámetro 'path'")
	}
	if c.Perm == "" {
		return fmt.Errorf("chmod: falta parámetro 'perm'")
	}
	return nil
}

// ==================== Comandos EXT3 ====================

// JournalingCommand representa el comando journaling
type JournalingCommand struct {
	BaseCommand
	ID string
}

func (c *JournalingCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("journaling: falta parámetro 'id'")
	}
	return nil
}

// RecoveryCommand representa el comando recovery
type RecoveryCommand struct {
	BaseCommand
	ID string
}

func (c *RecoveryCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("recovery: falta parámetro 'id'")
	}
	return nil
}

// LossCommand representa el comando loss
type LossCommand struct {
	BaseCommand
	ID string
}

func (c *LossCommand) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("loss: falta parámetro 'id'")
	}
	return nil
}

// JournalingCommand representa el comando journaling
