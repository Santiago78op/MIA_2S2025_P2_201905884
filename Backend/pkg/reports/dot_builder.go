package reports

import (
	"fmt"
	"strings"
)

type dot struct {
	sb strings.Builder
}

func newDot(title string) *dot {
	d := &dot{}
	d.sb.WriteString("digraph G {\n")
	// estilos globales razonables
	d.sb.WriteString(`graph [fontname="Helvetica"]; node [fontname="Helvetica"]; edge [fontname="Helvetica"];` + "\n")
	if title != "" {
		d.sb.WriteString(fmt.Sprintf(`labelloc="t"; label="%s";`+"\n", escape(title)))
	}
	return d
}

func (d *dot) close() string {
	d.sb.WriteString("}\n")
	return d.sb.String()
}

func (d *dot) line(s string) { d.sb.WriteString(s + "\n") }

func (d *dot) subgraph(name string, body func()) {
	d.line("subgraph " + name + " {")
	body()
	d.line("}")
}

func escape(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "|", `\|`)
	return s
}

func tableRecord(rows []string) string {
	// rows ya vienen escapadas; usamos shape=record
	return "{" + strings.Join(rows, "|") + "}"
}

func rowKV(k, v string) string {
	return fmt.Sprintf(`<%s> %s: %s`, escape(k), escape(k), escape(v))
}
