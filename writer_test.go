package gocsvj

import (
	"reflect"
	"strings"
	"testing"
)

func TestSimpleWriter(t *testing.T) {
	var sw strings.Builder

	w := NewWriter(&sw)
	w.WriteHeader([]string{"h1", "h2", "h3"})
	err := w.Write([]int{2, 4, 5})

	if err != nil {
		t.Error(err)
	}

	w.Flush()

	if sw.String() != `"h1","h2","h3"`+"\r\n2,4,5\r\n" {
		t.Error("unexpected CSVJ")
	}

	err = w.Error()
	if err != nil {
		t.Error(err)
	}
}

func TestInterface(t *testing.T) {
	var sw strings.Builder

	w := NewWriter(&sw)
	w.WriteHeader([]string{"h1", "h2", "h3"})
	err := w.Write([]interface{}{"test", nil, 42})

	if err != nil {
		t.Error(err)
	}

	w.Flush()

	if sw.String() != `"h1","h2","h3"`+"\r\n"+`"test",null,42`+"\r\n" {
		t.Error("unexpected CSVJ: ", sw.String())
	}

	err = w.Error()
	if err != nil {
		t.Error(err)
	}
}

func TestWriterNonSlice(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)
	w.WriteHeader([]string{"h1"})
	err := w.Write(42)

	if err == nil {
		t.Error("Expected error, but none returned")
	}
}

func TestWriterBadHeader(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)
	w.WriteHeader([]string{""})
	err := w.Write([]string{"item", "item2"})

	if err == nil {
		t.Error("Expected header error")
	}
}

func TestWriteNonCSVJ(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)

	w.WriteHeader([]string{"h1", "h2", "h3"})

	mp := make(map[string]string)
	mp["test"] = "test"
	err := w.Write([]interface{}{2, 3, mp})

	if err == nil {
		t.Error("Expected error, but none returned")
	}
}

// The tests below cover writer corners that are already pinned down (RFC 8259
// escape rules and existing writer behavior).

func TestWriterUTF8Values(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)

	w.WriteHeader([]string{"h1", "h2", "h3"})
	if err := w.Write([]interface{}{"héllo", "日本語", "🚀"}); err != nil {
		t.Fatal(err)
	}
	w.Flush()

	if !strings.Contains(sw.String(), `"héllo"`) {
		t.Error("UTF-8 string not preserved in output:", sw.String())
	}
	if !strings.Contains(sw.String(), `"日本語"`) {
		t.Error("UTF-8 string not preserved in output:", sw.String())
	}
}

func TestWriterRoundTripUTF8(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)

	w.WriteHeader([]string{"h1", "h2", "h3"})
	if err := w.Write([]interface{}{"héllo", "日本語", "🚀"}); err != nil {
		t.Fatal(err)
	}
	w.Flush()

	r := NewReader(strings.NewReader(sw.String()))
	hdr, err := r.Headers()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(hdr, []string{"h1", "h2", "h3"}) {
		t.Error("header round-trip mismatch:", hdr)
	}

	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	erow := []interface{}{"héllo", "日本語", "🚀"}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}
}

func TestWriterStringEscapes(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)

	w.WriteHeader([]string{"h1", "h2", "h3"})
	if err := w.Write([]interface{}{`line1` + "\n" + `line2`, `quote"end`, `back\slash`}); err != nil {
		t.Fatal(err)
	}
	w.Flush()

	r := NewReader(strings.NewReader(sw.String()))
	if _, err := r.Headers(); err != nil {
		t.Fatal(err)
	}
	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	erow := []interface{}{"line1\nline2", `quote"end`, `back\slash`}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}
}

func TestWriterNumberTypes(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)

	w.WriteHeader([]string{"h1", "h2", "h3", "h4"})
	if err := w.Write([]interface{}{-1, 0, 1.5, -2.5}); err != nil {
		t.Fatal(err)
	}
	w.Flush()

	r := NewReader(strings.NewReader(sw.String()))
	if _, err := r.Headers(); err != nil {
		t.Fatal(err)
	}
	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	erow := []interface{}{-1.0, 0.0, 1.5, -2.5}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}
}

func TestWriterBooleans(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)

	w.WriteHeader([]string{"h1", "h2"})
	if err := w.Write([]interface{}{true, false}); err != nil {
		t.Fatal(err)
	}
	w.Flush()

	r := NewReader(strings.NewReader(sw.String()))
	if _, err := r.Headers(); err != nil {
		t.Fatal(err)
	}
	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	erow := []interface{}{true, false}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}
}

func TestWriterMultipleRows(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)

	w.WriteHeader([]string{"h1", "h2"})
	for _, rec := range [][]int{{1, 2}, {3, 4}, {5, 6}} {
		if err := w.Write(rec); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()

	r := NewReader(strings.NewReader(sw.String()))
	if _, err := r.Headers(); err != nil {
		t.Fatal(err)
	}
	expected := [][]interface{}{
		{1.0, 2.0},
		{3.0, 4.0},
		{5.0, 6.0},
	}
	for i, erow := range expected {
		row, err := r.Read()
		if err != nil {
			t.Fatalf("row %d: %v", i, err)
		}
		if !reflect.DeepEqual(row, erow) {
			t.Errorf("row %d: got %v, expected %v", i, row, erow)
		}
	}
}

func TestWriterDuplicateHeader(t *testing.T) {
	// Spec §1 rule 4: writer should reject duplicate header names so it
	// cannot produce a file the strict reader would refuse.
	var sw strings.Builder
	w := NewWriter(&sw)
	if err := w.WriteHeader([]string{"a", "b", "a"}); err == nil {
		t.Error("expected rejection of duplicate header name")
	}
}

func TestWriterDuplicateEmptyHeader(t *testing.T) {
	var sw strings.Builder
	w := NewWriter(&sw)
	if err := w.WriteHeader([]string{"", ""}); err == nil {
		t.Error("expected rejection of duplicate empty-string headers")
	}
}

func TestWriterNonStrictHeaderLength(t *testing.T) {
	// With StrictHeaders disabled the writer should accept rows whose length
	// does not match the header. (Spec-wise this produces a ragged file —
	// the conformance work in PLAN.md §5 may invert that later, but the
	// writer's own StrictHeaders=false escape hatch should keep working.)
	var sw strings.Builder
	w := NewWriter(&sw)
	w.StrictHeaders = false

	w.WriteHeader([]string{"h1", "h2", "h3"})
	if err := w.Write([]int{1, 2}); err != nil {
		t.Error("Unexpected error with StrictHeaders=false:", err)
	}
}
