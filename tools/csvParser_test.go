package tools_test

import (
	"os"
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
