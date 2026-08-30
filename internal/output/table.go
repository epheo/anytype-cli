package output

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	padding    = 2
	ellipsis   = "..."
	noMaxWidth = 0
)

// Truncate counts runes so multi-byte names are never cut mid-character.
func Truncate(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	if max <= len(ellipsis) {
		return string([]rune(s)[:max])
	}
	return string([]rune(s)[:max-len(ellipsis)]) + ellipsis
}

type Table struct {
	headers []string
	max     []int
	rows    [][]string
}

func NewTable(headers ...string) *Table {
	return &Table{headers: headers, max: make([]int, len(headers))}
}

// MaxWidth caps one column; ID columns should stay uncapped since users copy them.
func (t *Table) MaxWidth(col, width int) *Table {
	if col >= 0 && col < len(t.max) {
		t.max[col] = width
	}
	return t
}

func (t *Table) Row(cells ...string) *Table {
	row := make([]string, len(t.headers))
	for i := range row {
		if i < len(cells) {
			row[i] = strings.ReplaceAll(cells[i], "\n", " ")
		}
	}
	t.rows = append(t.rows, row)
	return t
}

func (t *Table) Len() int { return len(t.rows) }

func (t *Table) Write(w io.Writer) {
	if len(t.headers) == 0 {
		return
	}
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if n := utf8.RuneCountInString(cell); n > widths[i] {
				widths[i] = n
			}
		}
	}
	for i := range widths {
		if t.max[i] != noMaxWidth && widths[i] > t.max[i] {
			widths[i] = t.max[i]
		}
	}

	writeRow := func(cells []string) {
		for i, cell := range cells {
			cell = Truncate(cell, widths[i])
			if i > 0 {
				fmt.Fprint(w, strings.Repeat(" ", padding))
			}
			fmt.Fprint(w, cell)
			if i < len(cells)-1 {
				fmt.Fprint(w, strings.Repeat(" ", widths[i]-utf8.RuneCountInString(cell)))
			}
		}
		fmt.Fprintln(w)
	}

	writeRow(t.headers)
	sep := make([]string, len(widths))
	for i, wd := range widths {
		sep[i] = strings.Repeat("-", wd)
	}
	writeRow(sep)
	for _, row := range t.rows {
		writeRow(row)
	}
}

// Details prints aligned "Key: value" lines for single-resource views.
type Details struct {
	pairs [][2]string
}

func NewDetails() *Details { return &Details{} }

func (d *Details) Add(key, value string) *Details {
	d.pairs = append(d.pairs, [2]string{key, value})
	return d
}

// AddIf skips empty values so optional fields do not clutter the view.
func (d *Details) AddIf(key, value string) *Details {
	if value != "" {
		d.Add(key, value)
	}
	return d
}

func (d *Details) Write(w io.Writer) {
	width := 0
	for _, p := range d.pairs {
		if n := utf8.RuneCountInString(p[0]); n > width {
			width = n
		}
	}
	for _, p := range d.pairs {
		fmt.Fprintf(w, "%s:%s %s\n", p[0], strings.Repeat(" ", width-utf8.RuneCountInString(p[0])), p[1])
	}
}
