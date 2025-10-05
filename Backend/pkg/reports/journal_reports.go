package reports

import (
	"fmt"
	"time"
)

// ReportJournal lista entradas con timeline (orden cronológico).
func ReportJournal(entries []JournalEntry, opt Options) string {
	d := newDot(or(opt.Title, "Journal"))
	d.line(`rankdir=LR; node [shape=record];`)
	prev := ""
	for i, e := range entries {
		id := fmt.Sprintf("J%d", i+1)
		ts := e.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05Z")
		rows := []string{
			rowKV("Op", e.Op),
			rowKV("Path", e.Path),
			rowKV("At", ts),
		}
		if e.Content != "" {
			rows = append(rows, rowKV("Content", ellipsis(e.Content, 120)))
		}
		d.line(fmt.Sprintf(`%s [label="%s"];`, id, tableRecord(rows)))
		if prev != "" {
			d.line(fmt.Sprintf(`%s -> %s;`, prev, id))
		}
		prev = id
	}
	return d.close()
}

func ellipsis(s string, n int) string {
	if len(s) <= n { return s }
	if n <= 3 { return s[:n] }
	return s[:n-3] + "..."
}
