package tests

import (
	"strings"
	"testing"

	"backend/utils"
)

// FakeCommandRunner simula la ejecución de comandos individuales.
type FakeCommandRunner struct{}

func (f FakeCommandRunner) Run(line string) (string, error) {
	// Simula resultados diferentes por comando.
	switch {
	case strings.HasPrefix(line, "mkdisk"):
		return "[ok] mkdisk", nil
	case strings.HasPrefix(line, "fdisk"):
		return "[ok] fdisk", nil
	case strings.HasPrefix(line, "mount"):
		return "[ok] mount 841A", nil
	case strings.HasPrefix(line, "mkfs"):
		return "[ok] mkfs", nil
	case strings.HasPrefix(line, "mkdir"):
		return "[ok] mkdir", nil
	case strings.HasPrefix(line, "mkfile"):
		return "[ok] mkfile", nil
	case strings.HasPrefix(line, "cat"):
		return "contenido users.txt...", nil
	default:
		return "[ok] noop", nil
	}
}

func TestSMIASimpleE2E(t *testing.T) {
	t.Parallel()

	script := `
# --- Discos ---
mkdisk -size=10 -unit=M -path="/tmp/Disco1.mia"

# --- Particiones ---
fdisk -size=1024 -unit=K -type=P -path="/tmp/Disco1.mia" -name=Part1

# --- Montaje ---
mount -path="/tmp/Disco1.mia" -name=Part1

# --- Formateo ---
mkfs -id=841A

# --- Archivos ---
mkdir -p -path=/home/docs
mkfile -size=15 -path=/home/docs/a.txt
cat -file1=/users.txt -file2=/home/docs/a.txt
`
	lines := utils.ParseSMIA(script)
	if len(lines) == 0 {
		t.Fatal("ParseSMIA devolvió 0 líneas")
	}

	runner := FakeCommandRunner{}
	var outputs []string
	for _, ln := range lines {
		out, err := runner.Run(ln)
		if err != nil {
			t.Fatalf("error ejecutando %q: %v", ln, err)
		}
		outputs = append(outputs, out)
	}
	joined := strings.Join(outputs, "\n")
	// Validaciones básicas del flujo:
	mustContain(t, joined, "[ok] mkdisk")
	mustContain(t, joined, "[ok] fdisk")
	mustContain(t, joined, "mount 841A")
	mustContain(t, joined, "[ok] mkfs")
	mustContain(t, joined, "[ok] mkdir")
	mustContain(t, joined, "contenido users.txt")
}

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("salida no contiene %q:\n%s", sub, s)
	}
}
