package gocsvj

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// A Writer writes records to a CSVJ encoded file.
type Writer struct {
	StrictHeaders bool

	w    *bufio.Writer
	hlen int
}

// NewWriter returns a new Writer that writes to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:             bufio.NewWriter(w),
		StrictHeaders: true, // True to check number of records in regular rows and header (default)
	}
}

// WriteHeader writes first CSVJ header-record
func (w *Writer) WriteHeader(header []string) error {
	seen := make(map[string]struct{}, len(header))
	for _, name := range header {
		if _, dup := seen[name]; dup {
			return fmt.Errorf("duplicate header name %q", name)
		}
		seen[name] = struct{}{}
	}
	w.hlen = len(header)
	return w.writeRaw(header)
}

// Writer writes a single CSVJ record to w along with any necessary quoting.
// A record is a slice of supported objects (see above).
func (w *Writer) Write(records interface{}) error {
	val := reflect.ValueOf(records)
	kind := val.Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		break
	default:
		return errors.New(fmt.Sprint("CSVJ records must be array or slice, not ", kind))
	}

	err := checkCSVJWritable(records)
	if err != nil {
		return err
	}

	if w.StrictHeaders && w.hlen != val.Len() {
		return errors.New("Record does not match header")
	}

	return w.writeRaw(records)
}

func (w *Writer) writeRaw(records interface{}) error {
	bytes, err := json.Marshal(records)
	bytes = bytes[1 : len(bytes)-1]

	if err != nil {
		return err
	}

	bytes = append(bytes, 13, 10)
	if _, err = w.w.Write(bytes); err != nil {
		return err
	}

	return nil
}

// Flush writes any buffered data to the underlying io.Writer.
// To check if an error occurred during the Flush, call Error.
func (w *Writer) Flush() {
	w.w.Flush()
}

// Error reports any error that has occurred during a previous Write or Flush.
func (w *Writer) Error() error {
	_, err := w.w.Write(nil)
	return err
}

func checkCSVJWritable(ar interface{}) error {

	val := reflect.ValueOf(ar)
	n := val.Len()

	for idx := 0; idx < n; idx++ {
		el := val.Index(idx)
		kind := el.Type().Kind()

		if kind == reflect.Interface {
			if el.IsNil() {
				continue
			}

			el = el.Elem()
			kind = el.Type().Kind()
		}

		switch kind {
		case reflect.Ptr, reflect.Interface, reflect.Slice,
			reflect.Struct, reflect.Map, reflect.Array:
			return fmt.Errorf("item %d is not CSVJ type-safe: %v", idx, kind)
		}
	}

	return nil
}
