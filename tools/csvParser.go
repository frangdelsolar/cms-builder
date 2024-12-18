package tools

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type CsvParser struct{}

// Parse reads a CSV file and fills a slice of structs with the data.
// The CSV header must match the field names in the struct.
// The CSV data is JSON encoded and then unmarshaled into the slice.
// Validation is done on the CSV header and the struct fields to ensure
// all fields are present.
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

	keyMap := make(map[string]int)
	for i := 0; i < len(keys); i++ {
		keyMap[keys[i]] = i
	}

	// Validate keys in JSON with struct fields (and map them)
	structValue := reflect.ValueOf(dataSlice).Elem() // Dereference slice pointer
	structType := structValue.Type().Elem()          // Get underlying struct type

	// Map CSV header names to struct field names
	fieldMap := make(map[string]int)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}
		jsonTag, ok := field.Tag.Lookup("json")
		if !ok {
			jsonTag = field.Name
		}
		fieldMap[jsonTag] = i
	}

	for k := range keyMap {
		if _, ok := fieldMap[k]; !ok {
			return fmt.Errorf("key %s not found in struct", k)
		}
	}

	for k := range fieldMap {
		if _, ok := keyMap[k]; !ok {
			return fmt.Errorf("key %s not found in CSV header", k)
		}
	}

	err = json.Unmarshal([]byte(jsonString), dataSlice)
	if err != nil {
		return err
	}

	return nil
}
