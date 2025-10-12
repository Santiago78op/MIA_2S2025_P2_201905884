package services

type FsService interface {
	Mkfs(id string) (string, error)
	Mkdir(id, path string, parents bool) (string, error)
	Mkfile(id, path string, size int, cont string, recursive bool) (string, error)
	Cat(id string, files []string) (string, error)
}
