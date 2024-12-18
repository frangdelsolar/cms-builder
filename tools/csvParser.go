package tools

import (
	"encoding/csv"
	"encoding/json"
	"os"
)

type CsvParser struct{}

// Parse reads a CSV file from the given path and unmarshals it into the given dataSlice.
//
// The CSV file is expected to have a header row with keys, and each subsequent row
// is a value for each key.
//
// The dataSlice must be a pointer to a slice of structs, where each field in the struct
// corresponds to a key in the CSV header row.
//
// For example, if the CSV file has a header row with keys "Name", "Age", and "Email",
// the dataSlice should be a pointer to a slice of structs with fields Name, Age, and Email.
func (c *CsvParser) Parse(path string, dataSlice interface{}) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.Comma = ','

	records, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	keys := []string{}
	keys = append(keys, records[0]...)

	var jsonString string
	jsonString += "["
	for i := 1; i < len(records); i++ {

		jsonString += "{"
		for j := 0; j < len(records[i]); j++ {
			jsonString += "\"" + keys[j] + "\": \"" + records[i][j] + "\","
		}
		jsonString = jsonString[:len(jsonString)-1]
		jsonString += "},"
	}

	jsonString = jsonString[:len(jsonString)-1]
	jsonString += "]"

	return json.Unmarshal([]byte(jsonString), dataSlice)
}
