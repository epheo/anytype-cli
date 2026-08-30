package cmd

import (
	"reflect"
	"testing"

	"github.com/epheo/anytype-go"
)

var testFormats = map[string]anytype.PropertyFormat{
	"title": anytype.PropertyFormatText, "count": anytype.PropertyFormatNumber,
	"done": anytype.PropertyFormatCheckbox, "tags": anytype.PropertyFormatMultiSelect,
	"due": anytype.PropertyFormatDate, "status": anytype.PropertyFormatSelect,
}

func TestParsePropertyValues(t *testing.T) {
	got, err := parsePropertyValues([]string{
		"title=a=b", "count=2.5", "done=true", "tags=t1,t2", "due=", "status=",
	}, testFormats)
	if err != nil {
		t.Fatal(err)
	}
	want := []anytype.PropertyLinkValue{
		anytype.TextPropertyLinkValue{Key: "title", Text: "a=b"},
		anytype.NumberPropertyLinkValue{Key: "count", Number: 2.5},
		anytype.CheckboxPropertyLinkValue{Key: "done", Checkbox: true},
		anytype.MultiSelectPropertyLinkValue{Key: "tags", MultiSelect: []string{"t1", "t2"}},
		anytype.DatePropertyLinkValue{Key: "due", Date: nil},
		anytype.SelectPropertyLinkValue{Key: "status", Select: nil},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v\nwant %#v", got, want)
	}
}

func TestParsePropertyValuesErrors(t *testing.T) {
	for _, bad := range []string{"novalue", "missing=1", "count=abc"} {
		if _, err := parsePropertyValues([]string{bad}, testFormats); err == nil {
			t.Errorf("%q: expected error", bad)
		}
	}
}

func TestParseFilters(t *testing.T) {
	got, err := parseFilters([]string{"title:contains:a:b", "count:gt:3", "due:empty", "tags:in:x,y", "name:eq:n"}, testFormats)
	if err != nil {
		t.Fatal(err)
	}
	want := []anytype.FilterItem{
		{Key: "title", Format: anytype.PropertyFormatText, Condition: anytype.FilterConditionContains, Value: "a:b"},
		{Key: "count", Format: anytype.PropertyFormatNumber, Condition: anytype.FilterConditionGt, Value: 3.0},
		{Key: "due", Condition: anytype.FilterConditionEmpty},
		{Key: "tags", Format: anytype.PropertyFormatMultiSelect, Condition: anytype.FilterConditionIn, Value: []string{"x", "y"}},
		{Key: "name", Format: anytype.PropertyFormatText, Condition: anytype.FilterConditionEq, Value: "n"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v\nwant %#v", got, want)
	}
	for _, bad := range []string{"title", "title:bogus:x", "count:eq", "count:eq:x"} {
		if _, err := parseFilters([]string{bad}, testFormats); err == nil {
			t.Errorf("%q: expected error", bad)
		}
	}
}
