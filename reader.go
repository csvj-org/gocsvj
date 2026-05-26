// Copyright CSVJ.org. All rights reserved.
// Use of this source code is governed by
// MIT license that can be found in the LICENSE file.

// Package github.com/csvj-org/gocsvj reads and writes comma-separated values files of CSVJ type.
// CSVJ is a csv-like format for storing tables that follows certain JSON encoding rules.

package gocsvj

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// A Reader reads records from a CSVJ-encoded file.
type Reader struct {
	row        int
	headerRead bool
	header     []string
	r          *bufio.Reader
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: bufio.NewReader(r),
	}
}

// Headers reads the header record from the input and caches it so subsequent
// calls return the same slice.
func (r *Reader) Headers() ([]string, error) {
	if r.headerRead {
		return r.header, nil
	}

	raw, err := r.readLine()
	if err != nil {
		return nil, err
	}

	values, err := parseLine(raw, "header")
	if err != nil {
		return nil, err
	}

	header, err := valuesAsStrings(values)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(header))
	for _, name := range header {
		if _, dup := seen[name]; dup {
			return nil, fmt.Errorf("duplicate header name %q", name)
		}
		seen[name] = struct{}{}
	}

	r.header = header
	r.headerRead = true
	return r.header, nil
}

// Read reads one data record from the input. It reads the header first if
// that has not happened yet. io.EOF is returned at end of input.
func (r *Reader) Read() ([]interface{}, error) {
	if !r.headerRead {
		if _, err := r.Headers(); err != nil {
			return nil, err
		}
	}

	r.row++
	raw, err := r.readLine()
	if err != nil {
		return nil, err
	}

	values, err := parseLine(raw, fmt.Sprintf("row %d", r.row))
	if err != nil {
		return nil, err
	}

	if len(values) != len(r.header) {
		return nil, fmt.Errorf("row %d has %d values, header has %d", r.row, len(values), len(r.header))
	}

	return values, nil
}

// readLine reads one CSVJ line and strips its terminator. A file ending
// without a final \n or \r\n is rejected per spec §1 rule 2. A clean EOF
// (no trailing partial line) is returned as io.EOF.
func (r *Reader) readLine() (string, error) {
	line, err := r.r.ReadString('\n')
	if err == io.EOF {
		if line == "" {
			return "", io.EOF
		}
		return "", errors.New("file does not end with newline")
	}
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line, nil
}

func parseLine(body, label string) ([]interface{}, error) {
	var values []interface{}
	if err := json.Unmarshal([]byte("["+body+"]"), &values); err != nil {
		return nil, fmt.Errorf("%s parse error: %s", label, err.Error())
	}
	if ok, idx := checkCSVJTypes(values); !ok {
		return nil, fmt.Errorf("%s parse error at item %d", label, idx)
	}
	return values, nil
}

func valuesAsStrings(vs []interface{}) ([]string, error) {
	strs := make([]string, len(vs))
	for i, v := range vs {
		w, ok := v.(string)
		if !ok {
			return nil, errors.New("non-string item at csvj header")
		}
		strs[i] = w
	}
	return strs, nil
}

func checkCSVJTypes(ar []interface{}) (bool, int) {
	for idx, el := range ar {
		if el == nil {
			continue
		}
		switch el.(type) {
		case float64:
		case string:
		case bool:
		default:
			return false, idx
		}
	}
	return true, -1
}
