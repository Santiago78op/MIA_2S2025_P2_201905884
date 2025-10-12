package tests

import (
	"testing"
	"time"

	fscmd "Backend/command/fs"
)

// Fake FsRepository para probar FsService sin EXT2 real
type fakeFsRepo struct {
	mkfsCount  int
	mkdirCalls []struct {
		id   string
		path []string
		p    bool
	}
	mkfileCalls []struct {
		id   string
		path []string
		size int
		cont string
		r    bool
	}
	catCalls [][]string
}

func (f *fakeFsRepo) Mkfs(id string) error { f.mkfsCount++; return nil }

func (f *fakeFsRepo) Mkdir(id string, absPath []string, parents bool, now time.Time) error {
	f.mkdirCalls = append(f.mkdirCalls, struct {
		id   string
		path []string
		p    bool
	}{id, absPath, parents})
	return nil
}

func (f *fakeFsRepo) Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool, now time.Time) error {
	f.mkfileCalls = append(f.mkfileCalls, struct {
		id   string
		path []string
		size int
		cont string
		r    bool
	}{id, absPath, size, contentHostPath, recursive})
	return nil
}

func (f *fakeFsRepo) Cat(id string, files [][]string) (string, error) {
	f.catCalls = append(f.catCalls, []string{}) // solo marca llamada
	return "contenido-a\ncontenido-b", nil
}

// Fake sesión: user normal no-root
type fakeSess struct{}

func (fakeSess) Current() (bool, string, int, int, bool) { return true, "user1", 10, 1, false }

func TestFsService_MkfsMkdirMkfileCat(t *testing.T) {
	t.Parallel()

	repo := &fakeFsRepo{}
	sess := &fakeSess{}
	svc := fscmd.NewFsService(repo, sess)

	// mkfs
	out, err := svc.Mkfs("841A")
	if err != nil || out == "" {
		t.Fatalf("mkfs err=%v out=%q", err, out)
	}
	if repo.mkfsCount != 1 {
		t.Fatalf("mkfsCount=%d", repo.mkfsCount)
	}

	// mkdir
	out, err = svc.Mkdir("841A", "/home/docs/usac", true)
	if err != nil || out == "" {
		t.Fatalf("mkdir err=%v out=%q", err, out)
	}
	if len(repo.mkdirCalls) != 1 || len(repo.mkdirCalls[0].path) != 3 {
		t.Fatalf("mkdirCalls=%#v", repo.mkdirCalls)
	}

	// mkfile
	out, err = svc.Mkfile("841A", "/home/docs/usac/a.txt", 15, "", true)
	if err != nil || out == "" {
		t.Fatalf("mkfile err=%v out=%q", err, out)
	}
	if len(repo.mkfileCalls) != 1 || repo.mkfileCalls[0].size != 15 {
		t.Fatalf("mkfileCalls=%#v", repo.mkfileCalls)
	}

	// cat
	txt, err := svc.Cat("841A", []string{"/home/docs/usac/a.txt", "/users.txt"})
	if err != nil || txt == "" {
		t.Fatalf("cat err=%v out=%q", err, txt)
	}
}

func TestFsService_ErroresBasicos(t *testing.T) {
	t.Parallel()

	repo := &fakeFsRepo{}
	sess := &fakeSess{}
	svc := fscmd.NewFsService(repo, sess)

	if _, err := svc.Mkdir("841A", "   ", true); err == nil {
		t.Fatal("esperaba error por path inválido")
	}
	if _, err := svc.Mkfile("841A", "/a/b", -1, "", false); err == nil {
		t.Fatal("esperaba error por size inválido")
	}
	if _, err := svc.Cat("841A", nil); err == nil {
		t.Fatal("esperaba error por sin archivos")
	}
}
