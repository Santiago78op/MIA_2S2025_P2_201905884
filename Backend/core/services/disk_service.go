package services

import "Backend/core/ports"

type DiskService interface {
	MkDisk(args any) (string, error) // usa tu DTO real (p.ej. command/disk.MkDiskArgs)
	RmDisk(args any) (string, error)
	FDisk(args any) (string, error)
	Mount(args any) (string, error)
	Mounted() []ports.MountedEntry
}
