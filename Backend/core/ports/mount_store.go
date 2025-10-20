package ports

// Estado de montajes (solo en RAM), con persistencia opcional de letra/contador en /tmp
type MountedEntry struct {
	ID   string // "841A"
	Path string // ruta al .mia
	Name string // nombre de la partición
}

type MountStore interface {
	NextID(carnet2, diskSignature string) (string, error) // genera id estable por firma
	SetMounted(id, path, name string) error
	List() []MountedEntry
	Unmount(id string) error                           // desmonta y opcionalmente resetea correlativo
	SetPartitionSeq(diskSignature string, seq int) error // establece correlativo manualmente
	GetMount(id string) (*MountedEntry, error)         // obtiene un montaje por ID
}
