package reports

import "time"

// Opciones generales de render
type Options struct {
	Title     string
	Rankdir   string // TB, LR
	NodeShape string // record, box, plaintext, etc.
	FontName  string // "Inter", "Helvetica", ...
	BgColor   string // "white"
}

// ====== Modelos de entrada minimalistas ======
// Adapta estos modelos a tus structs reales (MBR/EBR/SB, etc).

type MBRInfo struct {
	SizeBytes int64
	CreatedAt time.Time
	DiskSig   int32
	Fit       string
	// Particiones primarias/extendida
	Parts []PartInfo
}

type PartInfo struct {
	Status string // "used"|"free"
	Type   string // "P"|"E"
	Fit    string // "FF"|"BF"|"WF"
	Start  int64
	Size   int64
	Name   string
	// Para extendida:
	EBRs []EBRInfo
}

type EBRInfo struct {
	Status string
	Fit    string
	Start  int64 // offset del propio EBR
	Size   int64 // tamaño del bloque de datos lógico
	Next   int64 // -1 si no hay
	Name   string
}

// FS / Árbol
type TreeNode struct {
	Path     string
	IsDir    bool
	Mode     uint16
	Owner    string
	Group    string
	Children []TreeNode
}

// Bitmaps / Tablas
type Bitmap struct {
	Bits []bool // true=ocupado, false=libre
}

type SuperBlock struct {
	BlockSize   int
	InodeSize   int
	CountInodes int
	CountBlocks int
	FreeInodes  int
	FreeBlocks  int
	JournalN    int
	FirstDataAt int64 // offset o index
}

// Journal
type JournalEntry struct {
	Op        string
	Path      string
	Content   string // textual (corta)
	Timestamp time.Time
}
