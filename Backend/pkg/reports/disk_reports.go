package reports

import (
	"fmt"
)

// ReportMBR genera un DOT con una tabla record para MBR + primarias/extendida.
// Si una partición es extendida, agrega EBRs enlazados.
func ReportMBR(mbr MBRInfo, opt Options) string {
	d := newDot(opt.Title)
	d.line(`rankdir=` + or(opt.Rankdir, "LR") + `;`)
	d.line(`node [shape=record];`)
	// Nodo MBR
	mbrRows := []string{
		rowKV("SizeBytes", fmt.Sprintf("%d", mbr.SizeBytes)),
		rowKV("CreatedAt", mbr.CreatedAt.Format("2006-01-02 15:04:05")),
		rowKV("DiskSig", fmt.Sprintf("%d", mbr.DiskSig)),
		rowKV("Fit", mbr.Fit),
	}
	d.line(`MBR [label="` + tableRecord(mbrRows) + `"];`)

	// Particiones
	for i, p := range mbr.Parts {
		name := fmt.Sprintf("P%d", i+1)
		rows := []string{
			rowKV("Name", p.Name),
			rowKV("Type", p.Type),
			rowKV("Status", p.Status),
			rowKV("Fit", p.Fit),
			rowKV("Start", fmt.Sprintf("%d", p.Start)),
			rowKV("Size", fmt.Sprintf("%d", p.Size)),
		}
		d.line(fmt.Sprintf(`%s [label="%s"];`, name, tableRecord(rows)))
		d.line(fmt.Sprintf(`MBR -> %s;`, name))

		// EBRs si es extendida
		if p.Type == "E" && len(p.EBRs) > 0 {
			prev := ""
			for j, e := range p.EBRs {
				en := fmt.Sprintf("EBR_%d_%d", i+1, j+1)
				erows := []string{
					rowKV("Name", e.Name),
					rowKV("Status", e.Status),
					rowKV("Fit", e.Fit),
					rowKV("Start", fmt.Sprintf("%d", e.Start)),
					rowKV("Size", fmt.Sprintf("%d", e.Size)),
					rowKV("Next", fmt.Sprintf("%d", e.Next)),
				}
				d.line(fmt.Sprintf(`%s [label="%s"];`, en, tableRecord(erows)))
				if j == 0 {
					d.line(fmt.Sprintf(`%s -> %s [label="EBR"];`, name, en))
				} else if prev != "" {
					d.line(fmt.Sprintf(`%s -> %s [label="next"];`, prev, en))
				}
				prev = en
			}
		}
	}
	return d.close()
}

// ReportDiskLayout produce una barra horizontal con proporciones de particiones.
func ReportDiskLayout(mbr MBRInfo, opt Options) string {
	d := newDot(opt.Title)
	d.line(`rankdir=LR; node [shape=plaintext];`)
	// Construimos una tabla con celdas proporcionadas
	total := float64(mbr.SizeBytes)
	row := `<<TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0"><TR>`
	for _, p := range mbr.Parts {
		if p.Size <= 0 || p.Status != "used" {
			continue
		}
		w := (float64(p.Size) / total) * 100.0
		row += fmt.Sprintf(`<TD WIDTH="%.2f" FIXEDSIZE="false">%s<br/>%s<br/>%dB</TD>`, w, escape(p.Name), p.Type, p.Size)
	}
	row += `</TR></TABLE>>`
	d.line(`disk [label=` + row + `];`)
	return d.close()
}

func or(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
