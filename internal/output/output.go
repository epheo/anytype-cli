package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Format constants
const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatYAML  = "yaml"
)

// FormatAsJSON formats the data as JSON
func FormatAsJSON(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error formatting JSON: %w", err)
	}
	return string(jsonData), nil
}

// FormatAsYAML formats the data as YAML
func FormatAsYAML(data interface{}) (string, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error formatting YAML: %w", err)
	}
	return string(yamlData), nil
}

// Truncate limits a string to the specified length, adding ellipsis if needed
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// FormatTime converts Unix time to a human-readable format
func FormatTime(unixTime int64) string {
	if unixTime == 0 {
		return "N/A"
	}
	t := time.Unix(unixTime/1000, 0) // Anytype uses milliseconds
	return t.Format(time.RFC1123)
}

// Table represents a dynamic table for CLI output
type Table struct {
	Headers        []string
	Rows           [][]string
	MinWidth       int
	MaxWidth       int
	Padding        int
	TruncateLong   bool
	ColumnWidths   []int  // Custom max width per column
	ColumnTruncate []bool // Whether to truncate specific columns
}

// NewTable creates a new table with the given headers
func NewTable(headers []string) *Table {
	return &Table{
		Headers:        headers,
		Rows:           make([][]string, 0),
		MinWidth:       5,                          // Minimum width of 5 characters
		MaxWidth:       80,                         // Maximum width for any column
		Padding:        2,                          // Default padding of 2 characters
		TruncateLong:   false,                      // By default, don't truncate long values
		ColumnWidths:   make([]int, len(headers)),  // Default to 0 (use MaxWidth)
		ColumnTruncate: make([]bool, len(headers)), // Default to false (don't truncate)
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
}

// SetMinWidth sets the minimum width for columns
func (t *Table) SetMinWidth(width int) *Table {
	t.MinWidth = width
	return t
}

// SetMaxWidth sets the maximum width for columns
func (t *Table) SetMaxWidth(width int) *Table {
	t.MaxWidth = width
	return t
}

// SetPadding sets the padding between columns
func (t *Table) SetPadding(padding int) *Table {
	t.Padding = padding
	return t
}

// SetTruncate sets whether to truncate long values
func (t *Table) SetTruncate(truncate bool) *Table {
	t.TruncateLong = truncate
	return t
}

// SetColumnWidth sets the maximum width for a specific column
func (t *Table) SetColumnWidth(columnIndex int, width int) *Table {
	if columnIndex >= 0 && columnIndex < len(t.ColumnWidths) {
		t.ColumnWidths[columnIndex] = width
	}
	return t
}

// SetColumnTruncate sets whether to truncate a specific column
func (t *Table) SetColumnTruncate(columnIndex int, truncate bool) *Table {
	if columnIndex >= 0 && columnIndex < len(t.ColumnTruncate) {
		t.ColumnTruncate[columnIndex] = truncate
	}
	return t
}

// String returns a string representation of the table
func (t *Table) String() string {
	if len(t.Headers) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(t.Headers))
	for i, header := range t.Headers {
		widths[i] = len(header)
	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Apply minimum and maximum width
	for i := range widths {
		if widths[i] < t.MinWidth {
			widths[i] = t.MinWidth
		}

		// Apply custom column width if set
		if t.ColumnWidths[i] > 0 && widths[i] > t.ColumnWidths[i] {
			widths[i] = t.ColumnWidths[i]
		} else if t.MaxWidth > 0 && widths[i] > t.MaxWidth {
			// Otherwise apply global MaxWidth
			widths[i] = t.MaxWidth
		}
	}

	var b strings.Builder

	// Write header
	for i, header := range t.Headers {
		if i > 0 {
			b.WriteString(strings.Repeat(" ", t.Padding))
		}
		format := fmt.Sprintf("%%-%ds", widths[i])
		b.WriteString(fmt.Sprintf(format, header))
	}
	b.WriteString("\n")

	// Write header separator
	for i, w := range widths {
		if i > 0 {
			b.WriteString(strings.Repeat(" ", t.Padding))
		}
		b.WriteString(strings.Repeat("-", w))
	}
	b.WriteString("\n")

	// Write rows
	for _, row := range t.Rows {
		for i, cell := range row {
			if i >= len(widths) {
				continue
			}
			if i > 0 {
				b.WriteString(strings.Repeat(" ", t.Padding))
			}
			format := fmt.Sprintf("%%-%ds", widths[i])
			shouldTruncate := (t.TruncateLong || (i < len(t.ColumnTruncate) && t.ColumnTruncate[i])) && len(cell) > widths[i]

			if shouldTruncate {
				// Truncate with ellipsis if needed
				b.WriteString(fmt.Sprintf(format, Truncate(cell, widths[i])))
			} else {
				b.WriteString(fmt.Sprintf(format, cell))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}
