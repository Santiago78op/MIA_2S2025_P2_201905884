package disk

import (
	perrors "MIA_2S2025_P2_201905884/internal/errors"
)

// Aliases to P1 standard errors for backward compatibility
var (
	ErrInvalidParam   = perrors.ErrParams
	ErrNotFound       = perrors.ErrPartitionNotFound
	ErrExists         = perrors.ErrAlreadyExists
	ErrNoSpace        = perrors.ErrNoSpace
	ErrBadLayout      = perrors.ErrParams
	ErrUnsupported    = perrors.ErrParams
	ErrNotMounted     = perrors.ErrIDNotFound
	ErrAlreadyMounted = perrors.ErrAlreadyMounted
	ErrNoExtended     = perrors.ErrPartitionNotFound
)
