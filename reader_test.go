package vcsv

import (
	"bytes"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
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
	StringField       string           `csv:"field1"`
	IntField          int              `csv:"field2"`
	BoolField         bool             `csv:"field3"`
	DecimalFieldPtr   *decimal.Decimal `csv:"field4"`
	DecimalField      decimal.Decimal  `csv:"field5"`
	TimeField         RecordTime       `csv:"field6"`
	TimeFieldAlias    AliasTime        `csv:"field7"`
	NoTagFieldIgnored string
}

var decPtr = decimal.NewFromFloat(123.456)
var dec = decimal.NewFromFloat(321.456)

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
			expect:  TestStruct{StringField: "hello", IntField: 42, BoolField: true, DecimalFieldPtr: &decPtr, DecimalField: dec, TimeField: mockTimeDate(t), TimeFieldAlias: AliasTime(mockTimeDate(t).Time)},
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
			expect:    TestStruct{StringField: "hello", IntField: 42, BoolField: true, DecimalFieldPtr: nil, DecimalField: dec, TimeField: mockTimeDate(t), TimeFieldAlias: AliasTime(mockTimeDate(t).Time)},
		},
		{
			name:      "Empty Struct (decimal.Decimal)",
			csvData:   csvRow("hello,42,true,123.456,,2023/12/04 00:00:27 +0100,2023/12/04 00:00:27 +0100"),
			expectErr: false,
			expect:    TestStruct{StringField: "hello", IntField: 42, BoolField: true, DecimalFieldPtr: &decPtr, DecimalField: decimal.Decimal{}, TimeField: mockTimeDate(t), TimeFieldAlias: AliasTime(mockTimeDate(t).Time)},
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
			expect:    TestStruct{StringField: "hello", IntField: 42, BoolField: true, DecimalFieldPtr: &decPtr, DecimalField: dec, TimeField: mockTimeDate(t), TimeFieldAlias: AliasTime(mockTimeDate(t).Time)},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bytes.NewBufferString(tc.csvData)
			csvReader, err := New(reader, WithSeparationChar(','))
			require.NoError(t, err)

			// first line is the header
			var loopErr error
			csvReader.Next(&loopErr)
			require.NoError(t, loopErr)

			var result TestStruct
			err = csvReader.UnmarshalLine(&result)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect, result)
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
			require.NoError(t, err)

			// first line is the header
			var loopErr error
			csvReader.Next(&loopErr)
			require.NoError(t, loopErr)

			var result data
			err = csvReader.UnmarshalLine(&result)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect, result)
			}
		})
	}
}

func TestHeaderSettingAndGetting(t *testing.T) {
	header := []string{"column1", "column2", "column3"}
	reader := bytes.NewBufferString("data1,data2,data3\ndata4,data5,data6")
	csvReader, err := New(reader, WithSeparationChar(','), WithHeader(header))
	require.NoError(t, err)

	retrievedHeader := csvReader.Header()
	require.ElementsMatch(t, header, retrievedHeader, "Headers should match the input")
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
	require.NoError(t, err)

	var result TestStructByIndex
	// Assuming that the first line is the data, as no headers are set
	err = csvReader.UnmarshalLine(&result)
	require.NoError(t, err)

	expected := TestStructByIndex{StringField: "hello", IntField: 42, BoolField: true, TimeField: mockTimeDate(t).Time}
	require.Equal(t, expected, result)
}
