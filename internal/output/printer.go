// Package output renders CLI results as table, json, or yaml.
package output

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case FormatTable, FormatJSON, FormatYAML:
		return Format(s), nil
	}
	return "", fmt.Errorf("unknown output format %q (table, json, yaml)", s)
}

type Printer struct {
	Format Format
	Out    io.Writer
}

// Print marshals data for json/yaml and delegates to human for table.
func (p Printer) Print(data any, human func(w io.Writer)) error {
	switch p.Format {
	case FormatJSON:
		enc := json.NewEncoder(p.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case FormatYAML:
		return yaml.NewEncoder(p.Out).Encode(data)
	default:
		human(p.Out)
		return nil
	}
}
