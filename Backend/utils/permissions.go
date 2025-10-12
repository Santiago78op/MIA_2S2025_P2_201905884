package utils

// Permisos UGO y bypass de root. Los permisos del inodo vienen típicamente
// como {'6','6','4'} (caracteres octales). Este paquete interpreta y verifica.

type PermMode struct {
	U, G, O uint8 // cada uno en [0..7] (rwx en octal)
}

// ParsePerm convierte {'6','6','4'} o "664" a modos octales U,G,O.
func ParsePerm(p3 [3]byte) PermMode {
	toOct := func(b byte) uint8 {
		if b >= '0' && b <= '7' {
			return uint8(b - '0')
		}
		return 0
	}
	return PermMode{U: toOct(p3[0]), G: toOct(p3[1]), O: toOct(p3[2])}
}

// Bits: r=4, w=2, x=1
const (
	bitR = 4
	bitW = 2
	bitX = 1
)

func has(mode uint8, bit uint8) bool { return (mode & bit) == bit }

// CanRead/Write/Exec aplican UGO; si isRoot => true siempre (bypass del proyecto).
func CanRead(currUID, currGID int, isRoot bool, fileUID, fileGID int, p3 [3]byte) bool {
	if isRoot {
		return true
	}
	perm := ParsePerm(p3)
	switch {
	case currUID == fileUID:
		return has(perm.U, bitR)
	case currGID == fileGID:
		return has(perm.G, bitR)
	default:
		return has(perm.O, bitR)
	}
}

func CanWrite(currUID, currGID int, isRoot bool, fileUID, fileGID int, p3 [3]byte) bool {
	if isRoot {
		return true
	}
	perm := ParsePerm(p3)
	switch {
	case currUID == fileUID:
		return has(perm.U, bitW)
	case currGID == fileGID:
		return has(perm.G, bitW)
	default:
		return has(perm.O, bitW)
	}
}

func CanExec(currUID, currGID int, isRoot bool, fileUID, fileGID int, p3 [3]byte) bool {
	if isRoot {
		return true
	}
	perm := ParsePerm(p3)
	switch {
	case currUID == fileUID:
		return has(perm.U, bitX)
	case currGID == fileGID:
		return has(perm.G, bitX)
	default:
		return has(perm.O, bitX)
	}
}
