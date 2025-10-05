package reports

import (
	"fmt"
	"strings"
)

// ReportFSTree dibuja el árbol de directorios/archivos desde el nodo dado.
func ReportFSTree(root TreeNode, opt Options) string {
	d := newDot(or(opt.Title, "FS Tree"))
	d.line(`rankdir=` + or(opt.Rankdir, "TB") + `;`)
	d.line(`node [shape=box];`)
	var idx int
	var walk func(n TreeNode, parent string)
	walk = func(n TreeNode, parent string) {
		idx++
		id := fmt.Sprintf("N%d", idx)
		label := fmt.Sprintf("%s\n%s %s:%s",
			n.Path,
			permString(n.Mode),
			escape(n.Owner),
			escape(n.Group),
		)
		shape := "box"
		if n.IsDir {
			shape = "folder"
		} // igual usamos box; shape "folder" no es estándar, se ignora
		d.line(fmt.Sprintf(`%s [label="%s", shape=%s];`, id, escape(label), shape))
		if parent != "" {
			d.line(fmt.Sprintf(`%s -> %s;`, parent, id))
		}
		for _, c := range n.Children {
			walk(c, id)
		}
	}
	walk(root, "")
	return d.close()
}

// ReportBitmap genera un DOT con una fila de bits (█=1, ░=0) y una tabla.
func ReportBitmap(title string, bm Bitmap, opt Options) string {
	d := newDot(or(opt.Title, title))
	d.line(`rankdir=TB; node [shape=plaintext];`)
	// Línea compacta
	var bar strings.Builder
	for i, b := range bm.Bits {
		if b {
			bar.WriteRune('█')
		} else {
			bar.WriteRune('░')
		}
		if (i+1)%64 == 0 {
			bar.WriteRune('\n')
		}
	}
	d.line(fmt.Sprintf(`barchar [label="%s"];`, escape(bar.String())))
	// Tabla 0/1
	row := `<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0"><TR>`
	for i, b := range bm.Bits {
		val := "0"
		if b {
			val = "1"
		}
		row += fmt.Sprintf(`<TD>%s</TD>`, val)
		if (i+1)%32 == 0 {
			row += `</TR><TR>`
		}
	}
	row += `</TR></TABLE>>`
	d.line(`bittable [label=` + row + `];`)
	d.line(`barchar -> bittable [style=dashed];`)
	return d.close()
}

// ReportSuperblock es un alias para ReportSuperBlock
func ReportSuperblock(sb SuperBlock, opt Options) string {
	return ReportSuperBlock(sb, opt)
}

// ReportSuperBlock lista campos clave del superblock.
func ReportSuperBlock(sb SuperBlock, opt Options) string {
	d := newDot(or(opt.Title, "SuperBlock"))
	d.line(`rankdir=TB; node [shape=record];`)
	rows := []string{
		rowKV("BlockSize", fmt.Sprintf("%d", sb.BlockSize)),
		rowKV("InodeSize", fmt.Sprintf("%d", sb.InodeSize)),
		rowKV("CountInodes", fmt.Sprintf("%d", sb.CountInodes)),
		rowKV("CountBlocks", fmt.Sprintf("%d", sb.CountBlocks)),
		rowKV("FreeInodes", fmt.Sprintf("%d", sb.FreeInodes)),
		rowKV("FreeBlocks", fmt.Sprintf("%d", sb.FreeBlocks)),
		rowKV("JournalN", fmt.Sprintf("%d", sb.JournalN)),
		rowKV("FirstDataAt", fmt.Sprintf("%d", sb.FirstDataAt)),
	}
	d.line(`SB [label="` + tableRecord(rows) + `"];`)
	return d.close()
}

func permString(mode uint16) string {
	// simple: "0755"
	return fmt.Sprintf("%04o", mode)
}
