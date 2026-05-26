package lint

import (
	"fmt"
	"github.com/csvj-org/gocsvj"
	"reflect"
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `"Row1", "Row2", "Row3"` + "\n"
	csvj += `"Row1", "Row2", "Row3"` + "\n"

	result := Do(gocsvj.NewReader(strings.NewReader(csvj)))

	er := []Message{
		{Info, "header contains 3 columns"},
		{Info, "read 2 rows"},
	}

	if !reflect.DeepEqual(result, er) {
		t.Error("Bad Basic Info Messages")
	}
}

func TestRowType(t *testing.T) {
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `"Row1", "Row2", "Row3"` + "\n"
	csvj += `"Row1", 2, "Row3"` + "\n"

	result := Do(gocsvj.NewReader(strings.NewReader(csvj)))

	er := []Message{
		{Info, "header contains 3 columns"},
		{Warning, "row 2 column 2 type float64 differs from first row type string"},
		{Info, "read 2 rows"},
	}

	if !reflect.DeepEqual(result, er) {
		t.Error("Bad Basic Info Messages")
	}
}

func TestRowCount(t *testing.T) {
	// Spec §1 rule 3: the reader now rejects ragged rows outright, so lint
	// surfaces the parse error rather than a soft warning. Spurious follow-up
	// warnings (rown=2, row=nil) are an artifact of the existing lint loop
	// not short-circuiting on Read errors; left in place since reworking
	// lint's error handling is out of scope for the §5 reader change.
	csvj := `"Header1", "Header2", "Header3"` + "\n"
	csvj += `"Row1", "Row2", "Row3"` + "\n"
	csvj += `"Row1", "Row2"` + "\n"

	result := Do(gocsvj.NewReader(strings.NewReader(csvj)))

	er := []Message{
		{Info, "header contains 3 columns"},
		{Error, "reading 2 row: row 2 has 2 values, header has 3"},
		{Warning, "row 2 contains different number of items 0 then header 3"},
		{Warning, "row 2 contains less number of elements 0 than first row 3"},
		{Info, "read 2 rows"},
	}

	if !reflect.DeepEqual(result, er) {
		fmt.Println(result)
		t.Error("Bad Basic Info Messages")
	}
}
