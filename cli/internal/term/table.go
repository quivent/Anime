package term

import "strings"

// Table is a simple aligned table; widths tolerate ANSI-colored cells.
type Table struct {
	headers []string
	rows    [][]string
	right   map[int]bool
}

// NewTable creates a table with the given header cells.
func NewTable(headers ...string) *Table {
	return &Table{headers: headers, right: map[int]bool{}}
}

// Row appends a row and returns the table (chainable).
func (t *Table) Row(cells ...string) *Table {
	t.rows = append(t.rows, cells)
	return t
}

// Right marks 0-based column indexes as right-aligned (numbers).
func (t *Table) Right(cols ...int) *Table {
	for _, c := range cols {
		t.right[c] = true
	}
	return t
}

// Render returns the table: cyan header, dim hairline, then aligned rows.
func (t *Table) Render() string {
	cols := len(t.headers)
	for _, r := range t.rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	if cols == 0 {
		return ""
	}

	// compute visible column widths across header + rows
	w := make([]int, cols)
	at := func(cells []string, i int) string {
		if i < len(cells) {
			return cells[i]
		}
		return ""
	}
	for i := 0; i < cols; i++ {
		if l := visibleLen(at(t.headers, i)); l > w[i] {
			w[i] = l
		}
	}
	for _, r := range t.rows {
		for i := 0; i < cols; i++ {
			if l := visibleLen(at(r, i)); l > w[i] {
				w[i] = l
			}
		}
	}

	var b strings.Builder
	// header row in cyan
	b.WriteString("  ")
	for i := 0; i < cols; i++ {
		b.WriteString(t.cell(Cyan.S(at(t.headers, i)), at(t.headers, i), w[i], i))
		if i < cols-1 {
			b.WriteString("  ")
		}
	}
	b.WriteByte('\n')
	// dim hairline under header
	b.WriteString("  ")
	for i := 0; i < cols; i++ {
		b.WriteString(Dim(strings.Repeat("─", w[i])))
		if i < cols-1 {
			b.WriteString("  ")
		}
	}
	b.WriteByte('\n')
	// rows
	for _, r := range t.rows {
		b.WriteString("  ")
		for i := 0; i < cols; i++ {
			c := at(r, i)
			b.WriteString(t.cell(c, c, w[i], i))
			if i < cols-1 {
				b.WriteString("  ")
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// cell pads rendered (possibly colored) text to width w using the visible
// length of measure; honors right-alignment for column i.
func (t *Table) cell(rendered, measure string, w, i int) string {
	pad := w - visibleLen(measure)
	if pad < 0 {
		pad = 0
	}
	if t.right[i] {
		return strings.Repeat(" ", pad) + rendered
	}
	return rendered + strings.Repeat(" ", pad)
}
