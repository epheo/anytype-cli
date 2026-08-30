package output

import (
	"strings"
	"testing"
)

func TestTruncateIsRuneAware(t *testing.T) {
	s := "héllo wörld"
	got := Truncate(s, 8)
	if got != "héllo..." {
		t.Fatalf("got %q", got)
	}
	if Truncate(s, 100) != s {
		t.Fatal("short strings must pass through")
	}
	if Truncate("日本語テキスト", 4) != "日..." {
		t.Fatalf("got %q", Truncate("日本語テキスト", 4))
	}
}

func TestTableAlignsMultiByteCells(t *testing.T) {
	var b strings.Builder
	NewTable("ID", "NAME").MaxWidth(1, 6).
		Row("1", "日本語テキスト").
		Row("22", "a\nb").
		Write(&b)

	lines := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("lines = %d: %q", len(lines), b.String())
	}
	if lines[2] != "1   日本語..." {
		t.Errorf("row 1 = %q", lines[2])
	}
	if lines[3] != "22  a b" {
		t.Errorf("row 2 = %q", lines[3])
	}
}

func TestDetailsAlignment(t *testing.T) {
	var b strings.Builder
	NewDetails().Add("ID", "x").Add("Long Key", "y").AddIf("Skip", "").Write(&b)
	want := "ID:       x\nLong Key: y\n"
	if b.String() != want {
		t.Fatalf("got %q want %q", b.String(), want)
	}
}
