package vcsv

type Option func(*CSVReader)

// WithHeader sets the CSV header columns.
func WithHeader(columns []string) Option {
	return func(r *CSVReader) {
		r.headerAtLine = -1
		r.SetHeader(columns)
	}
}

// WithSeparationChar sets the CSV separation character.
func WithSeparationChar(separationChar rune) Option {
	return func(r *CSVReader) {
		r.reader.Comma = separationChar
	}
}

// WithReadHeader sets the line where the CSV header is located. If the value is negative,
// the header is not read. If the value is 0, the header is read from the first line.
// If the value is 1, the header is read from the second line, and so on.
//
// The default value is 0.
func WithReadHeader(line int) Option {
	return func(r *CSVReader) {
		r.headerAtLine = line
	}
}
