package tools_test

import (
	"os"
	"strings"
	"testing"

	"github.com/frangdelsolar/cms/tools"
	"github.com/stretchr/testify/assert"
)

type TestCsvInterface struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
	Field3 string `json:"field3"`
}

func TestCSVParser(t *testing.T) {
	// Create a temporary file
	file, err := os.Create("test.csv")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Write some data to the file
	_, err = file.WriteString("field1,field2,field3\n\"value1,9,0\",value2,value3\nvalue1,value2,value3\nvalue1,value2,value3")
	assert.NoError(t, err)

	// Create a new CSV parser
	parser := tools.CsvParser{}

	// Create a slice to hold the data
	var dataSlice []TestCsvInterface = []TestCsvInterface{}

	// Parse the file
	err = parser.Parse(file.Name(), &dataSlice)
	assert.NoError(t, err)

	// Assert that the data was parsed correctly
	assert.Equal(t, 3, len(dataSlice))
	assert.Equal(t, "value1,9,0", dataSlice[0].Field1)

	// Close the file
	err = file.Close()
	assert.NoError(t, err)
}

type TestStruct struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func TestCsvParser_Parse(t *testing.T) {
	t.Skip("Make sure it works at some point")
	tests := []struct {
		name       string
		path       string
		dataSlice  interface{}
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "valid CSV file with matching struct fields",
			path: "testdata/valid.csv",
			dataSlice: &[]TestStruct{
				{Field1: "value1", Field2: "value2"},
				{Field1: "value3", Field2: "value4"},
			},
			wantErr: false,
		},
		{
			name: "CSV file with missing struct fields",
			path: "testdata/missing_fields.csv",
			dataSlice: &[]TestStruct{
				{Field1: "value1", Field2: "value2"},
				{Field1: "value3", Field2: "value4"},
			},
			wantErr:    true,
			wantErrMsg: "key field3 not found in struct",
		},
		{
			name: "CSV file with extra struct fields",
			path: "testdata/extra_fields.csv",
			dataSlice: &[]TestStruct{
				{Field1: "value1", Field2: "value2"},
				{Field1: "value3", Field2: "value4"},
			},
			wantErr:    true,
			wantErrMsg: "key field3 not found in CSV header",
		},
		{
			name: "invalid CSV file (non-existent file)",
			path: "testdata/non_existent.csv",
			dataSlice: &[]TestStruct{
				{Field1: "value1", Field2: "value2"},
				{Field1: "value3", Field2: "value4"},
			},
			wantErr:    true,
			wantErrMsg: "open testdata/non_existent.csv: no such file or directory",
		},
		{
			name: "invalid CSV file (empty file)",
			path: "testdata/empty.csv",
			dataSlice: &[]TestStruct{
				{Field1: "value1", Field2: "value2"},
				{Field1: "value3", Field2: "value4"},
			},
			wantErr:    true,
			wantErrMsg: "EOF",
		},
		{
			name: "invalid CSV file (invalid JSON data)",
			path: "testdata/invalid_json.csv",
			dataSlice: &[]TestStruct{
				{Field1: "value1", Field2: "value2"},
				{Field1: "value3", Field2: "value4"},
			},
			wantErr:    true,
			wantErrMsg: "invalid character '}' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvParser := &tools.CsvParser{}
			err := csvParser.Parse(tt.path, tt.dataSlice)
			if (err != nil) != tt.wantErr {
				t.Errorf("CsvParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("CsvParser.Parse() error message = %v, wantErrMsg %v", err, tt.wantErrMsg)
			}
			if err == nil {
				data, ok := tt.dataSlice.(*[]TestStruct)
				if !ok {
					t.Errorf("dataSlice is not a *[]TestStruct")
				}
				if len(*data) != 2 {
					t.Errorf("expected 2 records, got %d", len(*data))
				}
				if (*data)[0].Field1 != "value1" || (*data)[0].Field2 != "value2" {
					t.Errorf("expected record 1 to be {Field1: value1, Field2: value2}, got %+v", (*data)[0])
				}
				if (*data)[1].Field1 != "value3" || (*data)[1].Field2 != "value4" {
					t.Errorf("expected record 2 to be {Field1: value3, Field2: value4}, got %+v", (*data)[1])
				}
			}
		})
	}
}
