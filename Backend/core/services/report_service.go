package services

type ReportService interface {
	Generate(name, id, out, extra string) (string, error)
}
