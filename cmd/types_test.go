package cmd

import (
	"testing"

	"github.com/epheo/anytype-go"
)

func TestParsePropertyDefs(t *testing.T) {
	defs, err := parsePropertyDefs([]string{"due:Due date:date", "notes::text"})
	if err != nil {
		t.Fatal(err)
	}
	if defs[0].Name != "Due date" || defs[0].Format != anytype.PropertyFormatDate {
		t.Errorf("got %+v", defs[0])
	}
	if defs[1].Name != "notes" {
		t.Errorf("name should default to key, got %+v", defs[1])
	}
	if _, err := parsePropertyDefs([]string{"key:name"}); err == nil {
		t.Error("expected error for missing format")
	}
}
