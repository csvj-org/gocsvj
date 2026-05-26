package gocsvj

import (
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
)

func TestSimpleReader(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `"Row1", "Row2", "Row3"` + "\n"
	csvj += " " // empty last line, just in case

	r := NewReader(strings.NewReader(csvj))

	// test initial parse and cache
	for l := 0; l < 2; l++ {
		hdr, err := r.Headers()
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(hdr, []string{"Header1", "Header2", "Header3"}) {
			t.Error("Unexpected Header")
		}
	}

	row, err := r.Read()
	if err != nil {
		t.Error(err)
	}

	erow := []interface{}{"Row1", "Row2", "Row3"}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}

	_, eofErr := r.Read()
	if eofErr != io.EOF {
		t.Error("EOF is expected on empty line")
	}
}

func TestSimpleReaderNoNewline(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `42, 42, false`

	r := NewReader(strings.NewReader(csvj))

	hdr, err := r.Headers()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(hdr, []string{"Header1", "Header2", "Header3"}) {
		t.Error("Unexpected Header")
	}

	row, err := r.Read()
	if err != nil {
		t.Error(err)
	}

	erow := []interface{}{42.0, 42.0, false}

	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow, "reason")
	}

	_, eofErr := r.Read()
	if eofErr != io.EOF {
		t.Error("EOF is expected on empty line")
	}
}

func TestReaderEmptyLineInMiddle(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += "\n"
	csvj += `null, null, true`

	r := NewReader(strings.NewReader(csvj))

	row, err := r.Read()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(row, []interface{}{}) {
		t.Error("Bad Row", row, "expected empty array")
	}

	row, err = r.Read()
	if err != nil {
		t.Error(err)
	}

	erow := []interface{}{nil, nil, true}

	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow, "reason")
	}

	_, eofErr := r.Read()
	if eofErr != io.EOF {
		t.Error("EOF is expected on empty line")
	}
}

func TestReaderParseError(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `42, $, false`

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err != nil {
		t.Error(err)
	}

	_, err = r.Read()
	if err == nil {
		t.Error("expected error, but none returned")
	}
}

func TestReaderParseJSLikeError(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `42, [], false`

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err != nil {
		t.Error(err)
	}

	_, err = r.Read()
	if err == nil {
		t.Error("expected error, but none returned")
	}
}

func TestReaderHeaderError(t *testing.T) {
	csvj := `"Header1", 1, "Header2", "Header3"` + "\n"

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err == nil {
		t.Error("expected error, but none returned")
	}
}

func TestReaderReadError(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `42, 1, false`

	r := NewReader(iotest.TimeoutReader(strings.NewReader(csvj)))

	_, err := r.Read()
	if err == nil {
		t.Error("expected error, but none returned")
	}
}

func TestReaderEmptyError(t *testing.T) {
	csvj := ""

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err == nil {
		t.Error("expected error, but none returned")
	}
}

// The tests below cover spec corners that are already pinned down (RFC 8259
// lexical rules and existing reader behavior). They exercise behavior the
// reader already supports; they intentionally do not assert on the §1
// structural rules tracked separately in PLAN.md §5 (strict §1 enforcement).

func TestReaderUTF8Values(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `"héllo", "日本語", "🚀"` + "\n"

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err != nil {
		t.Fatal(err)
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

func TestReaderCRLFLineEndings(t *testing.T) {
	csvj := `"h1", "h2", "h3"` + "\r\n"
	csvj += `1, 2, 3` + "\r\n"

	r := NewReader(strings.NewReader(csvj))

	hdr, err := r.Headers()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(hdr, []string{"h1", "h2", "h3"}) {
		t.Error("Unexpected Header", hdr)
	}

	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	erow := []interface{}{1.0, 2.0, 3.0}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}
}

func TestReaderJSONStringEscapes(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3", "Header4"` + "\n"
	csvj += `"line1\nline2", "tab\there", "quote\"end", "backslash\\"` + "\n"

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err != nil {
		t.Fatal(err)
	}

	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	erow := []interface{}{"line1\nline2", "tab\there", `quote"end`, `backslash\`}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}
}

func TestReaderUnicodeEscape(t *testing.T) {
	// RFC 8259 \uXXXX form, including a surrogate pair encoding U+1F600.
	// Raw-string literals keep the backslashes intact so the JSON layer
	// (not Go's lexer) is what resolves the escapes.
	csvj := `"Header1", "Header2"` + "\n"
	csvj += `"\u00e9", "\uD83D\uDE00"` + "\n"

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err != nil {
		t.Fatal(err)
	}

	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	erow := []interface{}{"é", "😀"}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}
}

func TestReaderNumberForms(t *testing.T) {
	csvj := `"h1", "h2", "h3", "h4", "h5"` + "\n"
	csvj += `-1, 0, 1.5, 1e10, -2.5e-3` + "\n"

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err != nil {
		t.Fatal(err)
	}

	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	erow := []interface{}{-1.0, 0.0, 1.5, 1e10, -2.5e-3}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}
}

func TestReaderBooleansAndNull(t *testing.T) {
	csvj := `"h1", "h2", "h3", "h4"` + "\n"
	csvj += `true, false, null, "string"` + "\n"

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err != nil {
		t.Fatal(err)
	}

	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	erow := []interface{}{true, false, nil, "string"}
	if !reflect.DeepEqual(row, erow) {
		t.Error("Bad Row", row, "expected", erow)
	}
}

func TestReaderMultipleRows(t *testing.T) {
	csvj := `"h1", "h2"` + "\n"
	csvj += `1, 2` + "\n"
	csvj += `3, 4` + "\n"
	csvj += `5, 6` + "\n"

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err != nil {
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
	if _, eofErr := r.Read(); eofErr != io.EOF {
		t.Error("EOF expected after final row, got", eofErr)
	}
}

func TestReaderLongValue(t *testing.T) {
	// 4 KiB string value — well under bufio.Scanner's default 64 KiB limit
	// and well above any value that would be inlined in normal CSV data.
	longValue := strings.Repeat("a", 4096)
	csvj := `"h1"` + "\n"
	csvj += `"` + longValue + `"` + "\n"

	r := NewReader(strings.NewReader(csvj))

	_, err := r.Headers()
	if err != nil {
		t.Fatal(err)
	}

	row, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(row) != 1 {
		t.Fatalf("expected 1 column, got %d", len(row))
	}
	got, ok := row[0].(string)
	if !ok {
		t.Fatalf("expected string, got %T", row[0])
	}
	if got != longValue {
		t.Error("long value did not round-trip")
	}
}
