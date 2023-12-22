package vcsv

import (
	"bytes"
	"math/big"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

type RecordTime struct {
	time.Time
}

// "2023/12/04 00:00:27 +0100"
// time with format YYYY/MM/DD HH:MM:SS offset
func (t *RecordTime) UnmarshalText(b []byte) (err error) {
	t.Time, err = time.Parse("2006/01/02 15:04:05 -0700", string(b))
	return err
}

type AliasTime time.Time

func (t *AliasTime) UnmarshalText(b []byte) (err error) {
	tt, err := time.Parse("2006/01/02 15:04:05 -0700", string(b))
	*t = AliasTime(tt)
	return err
}

type TestStruct struct {
	StringField       string     `csv:"field1"`
	IntField          int        `csv:"field2"`
	BoolField         bool       `csv:"field3"`
	DecimalFieldPtr   *big.Float `csv:"field4"`
	DecimalField      big.Float  `csv:"field5"`
	TimeField         RecordTime `csv:"field6"`
	TimeFieldAlias    AliasTime  `csv:"field7"`
	NoTagFieldIgnored string
}

func mockDecPtr() *big.Float {
	var decPtr, _, _ = big.ParseFloat("123.456", 10, 64, big.ToNearestEven)
	return decPtr
}

func mockDec() big.Float {
	var dec, _, _ = big.ParseFloat("321.456", 10, 64, big.ToNearestEven)
	return *dec
}

func mockTimeDate(t *testing.T) RecordTime {
	var tf RecordTime
	var err error
	tf.Time, err = time.Parse("2006/01/02 15:04:05 -0700", "2023/12/04 00:00:27 +0100")
	if err != nil {
		t.Fatalf("Failed to parse time for test fixture: %v", err)
	}
	return tf
}

func csvRow(vals string) string {
	return "field1,field2,field3,field4,field5,field6,field7\n" + vals
}

func TestReadIntoStruct(t *testing.T) {
	testCases := []struct {
		name      string
		csvData   string
		expect    TestStruct
		expectErr bool
	}{
		{
			name:    "Valid Data",
			csvData: csvRow("hello,42,true,123.456,321.456,2023/12/04 00:00:27 +0100,2023/12/04 00:00:27 +0100"),
			expect:  TestStruct{StringField: "hello", IntField: 42, BoolField: true, DecimalFieldPtr: mockDecPtr(), DecimalField: mockDec(), TimeField: mockTimeDate(t), TimeFieldAlias: AliasTime(mockTimeDate(t).Time)},
		},
		{
			name:      "Invalid Int",
			csvData:   csvRow("hello,notanint,true,123.456, 321.456,,2023/12/04 00:00:27 +0100,2023/12/04 00:00:27 +0100"),
			expectErr: true,
		},
		{
			name:      "Invalid Bool",
			csvData:   csvRow("hello,42,notabool,123.456,321.456,2023/12/04 00:00:27 +0100,2023/12/04 00:00:27 +0100"),
			expectErr: true,
		},
		{
			name:      "Nilled Struct (*decimal.Decimal)",
			csvData:   csvRow("hello,42,true,,321.456,2023/12/04 00:00:27 +0100,2023/12/04 00:00:27 +0100"),
			expectErr: false,
			expect:    TestStruct{StringField: "hello", IntField: 42, BoolField: true, DecimalFieldPtr: nil, DecimalField: mockDec(), TimeField: mockTimeDate(t), TimeFieldAlias: AliasTime(mockTimeDate(t).Time)},
		},
		{
			name:      "Empty Struct (decimal.Decimal)",
			csvData:   csvRow("hello,42,true,123.456,,2023/12/04 00:00:27 +0100,2023/12/04 00:00:27 +0100"),
			expectErr: false,
			expect:    TestStruct{StringField: "hello", IntField: 42, BoolField: true, DecimalFieldPtr: mockDecPtr(), DecimalField: big.Float{}, TimeField: mockTimeDate(t), TimeFieldAlias: AliasTime(mockTimeDate(t).Time)},
		},
		{
			name:      "Invalid StructPtr (*decimal.Decimal)",
			csvData:   csvRow("hello,42,true,notadec,321.456,2023/12/04 00:00:27 +0100,2023/12/04 00:00:27 +0100"),
			expectErr: true,
		},
		{
			name:      "Invalid Struct (decimal.Decimal)",
			csvData:   csvRow("hello,42,true,123.456,notadec,2023/12/04 00:00:27 +0100,2023/12/04 00:00:27 +0100"),
			expectErr: true,
		},
		{
			name:      "Valid Data with quotes",
			csvData:   csvRow(`"hello","42","true","123.456","321.456","2023/12/04 00:00:27 +0100","2023/12/04 00:00:27 +0100"`),
			expect:    TestStruct{StringField: "hello", IntField: 42, BoolField: true, DecimalFieldPtr: mockDecPtr(), DecimalField: mockDec(), TimeField: mockTimeDate(t), TimeFieldAlias: AliasTime(mockTimeDate(t).Time)},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bytes.NewBufferString(tc.csvData)
			csvReader, err := New(reader, WithSeparationChar(','))

			MustNoError(t, err)

			// first line is the header
			var loopErr error
			csvReader.Next(&loopErr)
			MustNoError(t, loopErr)

			var result TestStruct
			err = csvReader.UnmarshalLine(&result)

			if tc.expectErr {
				MustError(t, err)
			} else {
				MustNoError(t, err)
				if ok := reflect.DeepEqual(tc.expect, result); !ok {
					t.Fatalf("Expected %+v but got %+v", tc.expect, result)
				}
			}
		})
	}
}

func TestTimeParsingWithFormat(t *testing.T) {
	type data struct {
		TimeField time.Time `csv:"field1,format:2006/01/02 15:04:05 -0700"`
	}

	testCases := []struct {
		name      string
		csvData   string
		expect    data
		expectErr bool
	}{
		{
			name:      "Valid Data",
			csvData:   "field1\n2023/12/04 00:00:27 +0100",
			expect:    data{TimeField: mockTimeDate(t).Time},
			expectErr: false,
		},
		{
			name:      "Invalid Data",
			csvData:   "field1\nnotadate",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bytes.NewBufferString(tc.csvData)
			csvReader, err := New(reader, WithSeparationChar(','))
			MustNoError(t, err)

			// first line is the header
			var loopErr error
			csvReader.Next(&loopErr)
			MustNoError(t, loopErr)

			var result data
			err = csvReader.UnmarshalLine(&result)

			if tc.expectErr {
				MustError(t, err)
			} else {
				MustNoError(t, err)
				if ok := reflect.DeepEqual(tc.expect, result); !ok {
					t.Fatalf("Expected %+v but got %+v", tc.expect, result)
				}
			}
		})
	}
}

func TestHeaderSettingAndGetting(t *testing.T) {
	header := []string{"column1", "column2", "column3"}
	reader := bytes.NewBufferString("data1,data2,data3\ndata4,data5,data6")
	csvReader, err := New(reader, WithSeparationChar(','), WithHeader(header))
	MustNoError(t, err)

	retrievedHeader := csvReader.Header()
	if ok := slices.Equal(header, retrievedHeader); !ok {
		t.Fatalf("Expected %+v but got %+v", header, retrievedHeader)
	}
}

func TestReadIntoStructByColumnIndex(t *testing.T) {
	type TestStructByIndex struct {
		StringField string    `csv:"index:0"`
		IntField    int       `csv:"index:1"`
		BoolField   bool      `csv:"index:2"`
		TimeField   time.Time `csv:"index:3,format:2006/01/02 15:04:05 -0700"`
	}
	csvData := "hello,42,true,2023/12/04 00:00:27 +0100"
	reader := bytes.NewBufferString(csvData)
	csvReader, err := New(reader, WithSeparationChar(','))
	MustNoError(t, err)

	var result TestStructByIndex
	// Assuming that the first line is the data, as no headers are set
	err = csvReader.UnmarshalLine(&result)
	MustNoError(t, err)

	expected := TestStructByIndex{StringField: "hello", IntField: 42, BoolField: true, TimeField: mockTimeDate(t).Time}
	if ok := reflect.DeepEqual(expected, result); !ok {
		t.Fatalf("Expected %+v but got %+v", expected, result)
	}
}

func Test_skipBOM(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No BOM", "Hello, World!", "Hello, World!"},
		{"UTF-8 BOM", "\xEF\xBB\xBFHello, World!", "Hello, World!"},
		{"UTF-16LE BOM", "\xFF\xFEHello, World!", "Hello, World!"},
		{"UTF-16BE BOM", "\xFE\xFFHello, World!", "Hello, World!"},
		{"UTF-32LE BOM", "\xFF\xFE\x00\x00Hello, World!", "Hello, World!"},
		{"UTF-32BE BOM", "\x00\x00\xFE\xFFHello, World!", "Hello, World!"},
		{"Empty String", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Convert the input string to a Reader
			r := strings.NewReader(tc.input)

			// Execute the skipBOM function and capture the new reader
			newReader, err := skipBOM(r)
			if err != nil {
				t.Fatalf("skipBOM() returned an error: %v", err)
			}

			// Read the remaining data from the new reader
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(newReader)
			if err != nil {
				t.Fatalf("Reading from new reader returned an error: %v", err)
			}

			// Check if the remaining data is as expected
			if got := buf.String(); got != tc.expected {
				t.Errorf("Expected remaining data to be %q, got %q", tc.expected, got)
			}
		})
	}
}

func MustNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func MustError(t *testing.T, err error) {
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
}
