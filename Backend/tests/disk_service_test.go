package tests

import (
	"path/filepath"
	"testing"

	dcmd "Backend/command/disk"
)

// ----- Fakes para aislar la capa de aplicación -----

type fakeDiskRepo struct {
	created map[string]int64
	mbrSig  string
}

func (f *fakeDiskRepo) CreateDisk(path string, sizeBytes int64, fit rune) error {
	if f.created == nil {
		f.created = map[string]int64{}
	}
	f.created[path] = sizeBytes
	return nil
}
func (f *fakeDiskRepo) RemoveDisk(path string) error                      { delete(f.created, path); return nil }
func (f *fakeDiskRepo) FDiskPrimary(path string, a dcmd.FDiskArgs) error  { return nil }
func (f *fakeDiskRepo) FDiskExtended(path string, a dcmd.FDiskArgs) error { return nil }
func (f *fakeDiskRepo) FDiskLogical(path string, a dcmd.FDiskArgs) error  { return nil }
func (f *fakeDiskRepo) DiskSignature(path string) (string, error) {
	if f.mbrSig == "" {
		f.mbrSig = "sig-abc"
	}
	return f.mbrSig, nil
}
func (f *fakeDiskRepo) ValidatePrimaryForMount(path, name string) error { return nil }

type fakeMountStore struct {
	nextSeq int
	ents    []dcmd.MountedEntry
}

func (m *fakeMountStore) NextID(carnet2, diskSig string) (string, error) {
	m.nextSeq++
	// Formato igual que en service real: "84" + seq + "A"
	return carnet2 + itoa(m.nextSeq) + "A", nil
}
func (m *fakeMountStore) SetMounted(id, path, name string) error {
	m.ents = append(m.ents, dcmd.MountedEntry{ID: id, Path: path, Name: name})
	return nil
}
func (m *fakeMountStore) List() []dcmd.MountedEntry { return append([]dcmd.MountedEntry{}, m.ents...) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + (n % 10))
		n /= 10
	}
	return string(b[i:])
}

// ----- Tests -----

func TestDiskService_MkDiskRmDisk(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "Disco1.mia")

	repo := &fakeDiskRepo{}
	mnts := &fakeMountStore{}
	svc := dcmd.NewDiskService(repo, mnts, "84")

	out, err := svc.MkDisk(dcmd.MkDiskArgs{Size: 10, Unit: 'M', Fit: 'F', Path: path})
	if err != nil {
		t.Fatalf("MkDisk err: %v", err)
	}
	if out == "" || repo.created[path] == 0 {
		t.Fatalf("no creó disco: out=%q size=%d", out, repo.created[path])
	}

	out, err = svc.RmDisk(dcmd.RmDiskArgs{Path: path})
	if err != nil || out == "" {
		t.Fatalf("RmDisk err=%v out=%q", err, out)
	}
}

func TestDiskService_FDisk_and_Mount(t *testing.T) {
	t.Parallel()

	repo := &fakeDiskRepo{}
	mnts := &fakeMountStore{}
	svc := dcmd.NewDiskService(repo, mnts, "84")

	// Crea primaria
	_, err := svc.FDisk(dcmd.FDiskArgs{
		Size: 1024, Unit: 'K', Type: 'P', Fit: 'W', Path: "/tmp/Disco.mia", Name: "Part1",
	})
	if err != nil {
		t.Fatalf("FDisk P err: %v", err)
	}

	// Monta primaria → genera ID "841A"
	id, err := svc.Mount(dcmd.MountArgs{Path: "/tmp/Disco.mia", Name: "Part1"})
	if err != nil {
		t.Fatalf("Mount err: %v", err)
	}
	if id == "" || id[:2] != "84" {
		t.Fatalf("ID inválido: %q", id)
	}

	// Mounted lista 1
	list := svc.Mounted()
	if len(list) != 1 || list[0].ID != id {
		t.Fatalf("mounted mismatch: %#v", list)
	}
}
