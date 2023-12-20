# VCSV - Vertigo CSV Reader for Go

vcsv is a Go package providing a flexible and powerful way to read and 
unmarshal CSV data into Go structs. It supports custom CSV headers, various data 
types, and user-defined parsing rules through struct tags.

## Features

- Read CSV data and unmarshal into Go structs.
- Support for custom CSV headers.
- Handle various primitive types and custom types implementing `encoding.TextUnmarshaler`.
- Options such as `format` to specify the date format for `time.Time` fields.
- Flexible configuration options for CSV parsing.
- No external dependencies. Only uses the standard library.

## Installation

To install VCSV, use the following command:

```bash
go get -u github.com/fond-of-vertigo/vcsv
```

## Usage
Below are some examples of how to use the VCSV package.

### Basic Usage
First, define a struct that maps to your CSV format:

```go
import "time"

type Person struct {
    Name          string    `csv:"name"`
    Age           int       `csv:"age"`
    Birthdate     time.Time `csv:"birthdate,format:2006-01-02"`
    IsBirthday    bool      `csv:"is_birthday"`
}
```

When a CSV file contains no header, you can use the `index` tag to specify the
column index instead of the column name:
```go
type Person struct {
    Name          string    `csv:"index:0"`
    Age           int       `csv:"index:1"`
    Birthdate     time.Time `csv:"index:2,format:2006-01-02"`
    IsBirthday    bool      `csv:"index:3"`
}
```



Then, use VCSV to read and unmarshal data:

```go
package main

import (
	"fmt"
	"github.com/fond-of-vertigo/vcsv"
	"os"
)

func main() {
	file, err := os.Open("data.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader, err := vcsv.New(file, vcsv.WithSeparationChar(','))
	if err != nil {
		panic(err)
	}

	var p Person
	for reader.Next(&err) {
        // Unmarshal the current line into the Person struct.
		if err := reader.UnmarshalLine(&p); err != nil { 
			panic(err)
		}
        
		fmt.Printf("%+v\n", p)
	}
}
```

## Configuration Options
VCSV provides several options to configure the CSV reader:

- `WithHeader([]string)`: Sets the CSV header columns manually.
- `WithSeparationChar(rune)`: Sets a custom column separation character.
- `WithReadHeader(int)`: Specifies which line of the CSV file contains the header.

Example:
```go
reader, err := vcsv.New(file, vcsv.WithSeparationChar(';'), vcsv.WithHeader([]string{"Name", "Age", "Birthdate", "IsBirthday"}))
```