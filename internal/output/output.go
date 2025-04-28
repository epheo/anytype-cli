package output

import (
	"encoding/json"
	"fmt"
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
