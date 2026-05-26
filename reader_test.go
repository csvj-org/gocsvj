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
	// Spec §1 rule 2: every file MUST end with \n or \r\n. A data row
	// without a trailing terminator is rejected.
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `42, 42, false`

	r := NewReader(strings.NewReader(csvj))

	if _, err := r.Headers(); err != nil {
		t.Fatal(err)
	}

	if _, err := r.Read(); err == nil {
		t.Error("expected rejection of file without trailing newline")
	}
}

func TestReaderEmptyLineInMiddle(t *testing.T) {
	// Spec §1 rule 3: every data line MUST contain exactly as many values
	// as the header line. An empty line in a file with a non-zero header
	// is a 0-value row and gets rejected as ragged.
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += "\n"
	csvj += `null, null, true` + "\n"

	r := NewReader(strings.NewReader(csvj))

	if _, err := r.Read(); err == nil {
		t.Error("expected rejection of ragged (zero-value) row")
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

// The tests below assert the strict §1 rules now enforced by the reader.

func TestReaderEmptyHeaderLine(t *testing.T) {
	// Spec §1 rule 1: a single `\n` is the minimum valid CSVJ file. It
	// represents an empty header (zero columns) and zero data rows.
	r := NewReader(strings.NewReader("\n"))

	hdr, err := r.Headers()
	if err != nil {
		t.Fatal(err)
	}
	if len(hdr) != 0 {
		t.Errorf("expected empty header, got %v", hdr)
	}

	if _, err := r.Read(); err != io.EOF {
		t.Errorf("expected io.EOF after empty header, got %v", err)
	}
}

func TestReaderEmptyHeaderLineCRLF(t *testing.T) {
	r := NewReader(strings.NewReader("\r\n"))
	hdr, err := r.Headers()
	if err != nil {
		t.Fatal(err)
	}
	if len(hdr) != 0 {
		t.Errorf("expected empty header, got %v", hdr)
	}
}

func TestReaderTrailingNewlineRequired(t *testing.T) {
	// Header line is well-formed but the file lacks a final \n.
	r := NewReader(strings.NewReader(`"h1"`))

	if _, err := r.Headers(); err == nil {
		t.Error("expected rejection of header without trailing newline")
	}
}

func TestReaderRaggedShort(t *testing.T) {
	// Spec §1 rule 3: a row with fewer values than the header is rejected.
	csvj := `"a", "b", "c"` + "\n"
	csvj += `"x", "y"` + "\n"

	r := NewReader(strings.NewReader(csvj))
	if _, err := r.Read(); err == nil {
		t.Error("expected rejection of short ragged row")
	}
}

func TestReaderRaggedLong(t *testing.T) {
	// Spec §1 rule 3: a row with more values than the header is rejected.
	csvj := `"a", "b"` + "\n"
	csvj += `"x", "y", "z"` + "\n"

	r := NewReader(strings.NewReader(csvj))
	if _, err := r.Read(); err == nil {
		t.Error("expected rejection of long ragged row")
	}
}

func TestReaderDuplicateHeader(t *testing.T) {
	// Spec §1 rule 4: duplicate column names in the header are rejected.
	csvj := `"a", "b", "a"` + "\n"

	r := NewReader(strings.NewReader(csvj))
	if _, err := r.Headers(); err == nil {
		t.Error("expected rejection of duplicate header name")
	}
}

func TestReaderDuplicateEmptyHeader(t *testing.T) {
	// Spec §1 rule 4: the duplicate check applies to empty strings too;
	// only the all-empty-line case (zero columns) is exempt.
	csvj := `"", ""` + "\n"

	r := NewReader(strings.NewReader(csvj))
	if _, err := r.Headers(); err == nil {
		t.Error("expected rejection of duplicate empty-string headers")
	}
}
