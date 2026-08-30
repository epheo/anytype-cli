package spaces

import (
	"errors"
	"testing"

	"github.com/epheo/anytype-go"
)

var fixture = []anytype.Space{
	{ID: "bafy1", Name: "Work"},
	{ID: "bafy2", Name: "Work Notes"},
	{ID: "bafy3", Name: "Personal"},
}

func TestMatch(t *testing.T) {
	cases := []struct {
		query   string
		want    string
		wantErr error
	}{
		{"bafy3", "bafy3", nil},
		{"work", "bafy1", nil},
		{"WORK NOTES", "bafy2", nil},
		{"pers", "bafy3", nil},
		{"notes", "bafy2", nil},
		{"wor", "", ErrAmbiguous},
		{"unknown-id", "unknown-id", nil},
	}
	for _, c := range cases {
		got, err := Match(fixture, c.query)
		if !errors.Is(err, c.wantErr) {
			t.Errorf("%q: err = %v, want %v", c.query, err, c.wantErr)
		}
		if got != c.want {
			t.Errorf("%q: got %q, want %q", c.query, got, c.want)
		}
	}
}
