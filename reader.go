package vcsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
)

// CSVReader is a CSV reader that supports iterating and reading CSV lines into structs.
// The csv reader is read by default in the first line. If the header is not in the first line,
// you can use the WithReadHeader option to set the line where the header is located.
type CSVReader struct {
	columnIndex  map[string]int
	columns      []string
	reader       *csv.Reader
	headerAtLine int
}

// New creates a new CSVReader.
func New(r io.Reader, options ...Option) (*CSVReader, error) {
	if r == nil {
		return nil, errors.New("reader must not be nil")
	}
	c := CSVReader{}
	c.reader = csv.NewReader(r)
	c.reader.FieldsPerRecord = -1
	c.reader.LazyQuotes = true

	for _, option := range options {
		option(&c)
	}

	if err := c.readHeaderAtLine(c.headerAtLine); err != nil {
		return nil, err
	}
	return &c, nil
}

// Header returns the CSV header columns.
func (r *CSVReader) Header() []string {
	keys := make([]string, 0, len(r.columnIndex))
	for k := range r.columnIndex {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return r.columnIndex[keys[i]] < r.columnIndex[keys[j]]
	})
	return keys
}

// SetHeader sets the CSV header columns.
func (r *CSVReader) SetHeader(columns []string) {
	r.columns = columns
	r.columnIndex = make(map[string]int)
	for i, name := range r.columns {
		r.columnIndex[name] = i
	}
}

// ReadHeader reads the current line, that was already read by Next as the CSV header.
// This method is called automatically when the CSVReader is created, use it only if you want to read the header again.
func (r *CSVReader) ReadHeader() {
	r.SetHeader(r.columns)
}

// Next reads the next CSV line.
func (r *CSVReader) Next(err *error) bool {
	r.columns, *err = r.reader.Read()
	if *err == io.EOF {
		*err = nil
		return false
	}
	if *err != nil {
		return false
	}
	return true
}

// Get returns the value of the given column name.
func (r *CSVReader) Get(columnName string) (string, error) {
	i, columnExists := r.columnIndex[columnName]
	if !columnExists {
		return "", fmt.Errorf("invalid column \"%s\"", columnName)
	}
	if i >= len(r.columns) {
		return "", nil
	}
	return r.columns[i], nil
}

// GetByColumnIndex returns the value by the given column index.
func (r *CSVReader) GetByColumnIndex(columnIndex int) (string, error) {
	if columnIndex >= len(r.columns) {
		return "", nil
	}
	return r.columns[columnIndex], nil
}

// CurrentLineIndex returns the current CSV line index.
func (r *CSVReader) CurrentLineIndex() int {
	lineIndex, _ := r.reader.FieldPos(0)
	return lineIndex
}

// UnmarshalLine fills the given struct with data from the next CSV line.
// The struct fields should be annotated with the `csv` tag to map to CSV column names.
// The struct fields types may be any primitive type or implement encoding.TextUnmarshaler.
//
//
// Supported tag options:
// - `csv:"<column_name>"` - maps the struct field to the given CSV column name.
// - `csv:"index:<column_index>"` - maps the struct field to the given CSV column index.
// - `csv:"format:<time_format>"` - parses the CSV column value as a time.Time using the given format.
//
// Example:
//
//	type Person struct {
//		Name             string           `csv:"name"`
//		Age              int              `csv:"age"`
//		Birthday         time.Time        `csv:"birthdate,format:2006-01-02"`
//		TodayIsBirthday  bool             `csv:"is_birthday"`
//		HeightMeters     decimal.Decimal  `csv:"height_m"`
//	}
// Alternatively, you can use the `index` tag option to map the struct field to the CSV column index instead of the header name.
//	type Person struct {
//		Name             string           `csv:"index:0"`
//		Age              int              `csv:"index:1"`
//		Birthday         time.Time        `csv:"index:2"`
//		TodayIsBirthday  bool             `csv:"index:3"`
//		HeightMeters     decimal.Decimal  `csv:"index:4"`
//	}

func (r *CSVReader) UnmarshalLine(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a non-nil pointer to a struct")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.New("v must be a pointer to a struct")
	}

	rt := rv.Type()
	return r.parseFields(rt, rv)
}

func (r *CSVReader) readHeaderAtLine(line int) (err error) {
	if r.headerAtLine < 0 {
		return nil
	}

	if err := r.skipLines(line + 1); err != nil {
		return err
	}

	r.ReadHeader()
	return nil
}

func (r *CSVReader) skipLines(n int) error {
	var err error
	for i := 0; i < n; i++ {
		if ok := r.Next(&err); !ok {
			return fmt.Errorf("error skipping lines: %w", err)
		}
	}
	return nil
}

func (r *CSVReader) parseFields(rt reflect.Type, rv reflect.Value) error {
	for i := 0; i < rt.NumField(); i++ {
		if err := r.parseFieldValue(rt.Field(i), rv.Field(i)); err != nil {
			return err
		}
	}
	return nil
}
