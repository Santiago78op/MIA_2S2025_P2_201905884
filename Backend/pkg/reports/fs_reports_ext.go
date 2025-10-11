package reports

import (
	"fmt"
)

// GenerateINODEReport genera un reporte DOT de un inodo específico
func GenerateINODEReport(diskPath, partName, outputPath string) error {
	// Stub simple: genera un DOT básico
	dot := `digraph INODE {
	rankdir=TB;
	node [shape=record];

	inode [label="<f0> Inode Report|<f1> Disk: ` + diskPath + `|<f2> Partition: ` + partName + `|<f3> Status: Not fully implemented"];
}
`
	return WriteDOT(outputPath, dot)
}

// GenerateBLOCKReport genera un reporte DOT de un bloque específico
func GenerateBLOCKReport(diskPath, partName, outputPath string) error {
	// Stub simple: genera un DOT básico
	dot := `digraph BLOCK {
	rankdir=TB;
	node [shape=record];

	block [label="<f0> Block Report|<f1> Disk: ` + diskPath + `|<f2> Partition: ` + partName + `|<f3> Status: Not fully implemented"];
}
`
	return WriteDOT(outputPath, dot)
}

// GenerateBMINODEReport genera un reporte DOT del bitmap de inodos
func GenerateBMINODEReport(diskPath, partName, outputPath string) error {
	// Stub: genera un bitmap básico de ejemplo
	d := newDot("Bitmap Inodes - " + partName)
	d.line(`rankdir=TB; node [shape=plaintext];`)

	// Generar bitmap de ejemplo (32 bits)
	row := `<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0"><TR>`
	for i := 0; i < 32; i++ {
		val := "0"
		if i < 8 {
			val = "1" // Primeros 8 ocupados
		}
		row += fmt.Sprintf(`<TD>%s</TD>`, val)
	}
	row += `</TR></TABLE>>`
	d.line(`bm_inode [label=` + row + `];`)

	dot := d.close()
	return WriteDOT(outputPath, dot)
}

// GenerateBMBLOCKReport genera un reporte DOT del bitmap de bloques
func GenerateBMBLOCKReport(diskPath, partName, outputPath string) error {
	// Stub: genera un bitmap básico de ejemplo
	d := newDot("Bitmap Blocks - " + partName)
	d.line(`rankdir=TB; node [shape=plaintext];`)

	// Generar bitmap de ejemplo (32 bits)
	row := `<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0"><TR>`
	for i := 0; i < 32; i++ {
		val := "0"
		if i < 12 {
			val = "1" // Primeros 12 ocupados
		}
		row += fmt.Sprintf(`<TD>%s</TD>`, val)
	}
	row += `</TR></TABLE>>`
	d.line(`bm_block [label=` + row + `];`)

	dot := d.close()
	return WriteDOT(outputPath, dot)
}

// GenerateTREEReport genera un reporte DOT del árbol de directorios
func GenerateTREEReport(diskPath, partName, outputPath string) error {
	// Stub: genera árbol básico de ejemplo
	d := newDot("Tree - " + partName)
	d.line(`rankdir=TB; node [shape=box];`)

	// Árbol básico
	d.line(`N1 [label="/\n0755 root:root", shape=folder];`)
	d.line(`N2 [label="/users.txt\n0644 root:root"];`)
	d.line(`N1 -> N2;`)

	dot := d.close()
	return WriteDOT(outputPath, dot)
}

// GenerateFILEReport genera un reporte DOT del contenido de un archivo
func GenerateFILEReport(diskPath, partName, filePath, outputPath string) error {
	// Stub: genera reporte básico
	d := newDot("File Content - " + filePath)
	d.line(`rankdir=TB; node [shape=plaintext];`)

	content := `<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
		<TR><TD COLSPAN="2">File: ` + filePath + `</TD></TR>
		<TR><TD>Disk</TD><TD>` + diskPath + `</TD></TR>
		<TR><TD>Partition</TD><TD>` + partName + `</TD></TR>
		<TR><TD>Status</TD><TD>Report not fully implemented</TD></TR>
	</TABLE>>`
	d.line(`file [label=` + content + `];`)

	dot := d.close()
	return WriteDOT(outputPath, dot)
}

// GenerateLSReport genera un reporte tipo ls de un directorio
func GenerateLSReport(diskPath, partName, dirPath, outputPath string) error {
	// Stub: similar a TREE pero para un directorio específico
	d := newDot("LS - " + dirPath)
	d.line(`rankdir=TB; node [shape=plaintext];`)

	content := `<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
		<TR><TD COLSPAN="4">Directory Listing: ` + dirPath + `</TD></TR>
		<TR><TD>Type</TD><TD>Name</TD><TD>Perms</TD><TD>Owner</TD></TR>
		<TR><TD>d</TD><TD>/</TD><TD>0755</TD><TD>root:root</TD></TR>
		<TR><TD>f</TD><TD>users.txt</TD><TD>0644</TD><TD>root:root</TD></TR>
	</TABLE>>`
	d.line(`ls [label=` + content + `];`)

	dot := d.close()
	return WriteDOT(outputPath, dot)
}
