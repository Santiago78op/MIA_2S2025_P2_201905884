package fs

import "errors"

var (
	ErrNotFound     = errors.New("fs: not found")
	ErrExists       = errors.New("fs: already exists")
	ErrInvalidPath  = errors.New("fs: invalid path")
	ErrInvalidPerm  = errors.New("fs: invalid permission")
	ErrNotAFile     = errors.New("fs: not a file")
	ErrNotADir      = errors.New("fs: not a directory")
	ErrNoSpace      = errors.New("fs: no space left")
	ErrUnsupported  = errors.New("fs: unsupported")
	ErrUnauthorized = errors.New("fs: unauthorized")
)
