package tools

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type CsvParser struct{}

// GenerateCsvTemplate takes a struct and returns a CSV string with the JSON tags
// as column headers. If a JSON tag is not present, the field name is used as the
// column header instead. The returned CSV string is empty except for the header
// row.
func (c *CsvParser) GenerateCsvTemplate(model interface{}) []byte {

	keys := []string{}
	// Get the struct type
	structType := reflect.TypeOf(model)
	// Iterate over the struct fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}
		jsonTag, ok := field.Tag.Lookup("json")
		if !ok {
			jsonTag = field.Name
		}
		keys = append(keys, jsonTag)
	}
	// Convert the slice of keys to a CSV string
	csvString := fmt.Sprintf("%s\n", strings.Join(keys, ","))
	return []byte(csvString)
}

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

	err = ValidateKeys(keys, dataSlice)
	if err != nil {
		return err
	}

	var jsonString string
	jsonString += "["
	for i := 1; i < len(records); i++ {

		jsonString += "{"
		for j := 0; j < len(records[i]); j++ {
			jsonString += GetValueForFieldType(records[i][j], keys[j], dataSlice)
		}
		jsonString = jsonString[:len(jsonString)-1]
		jsonString += "},"
	}

	jsonString = jsonString[:len(jsonString)-1]
	jsonString += "]"

	err = json.Unmarshal([]byte(jsonString), dataSlice)
	if err != nil {
		return err
	}

	return nil
}

// ValidateKeys checks that the keys from a CSV header match the field names
// in a provided slice of structs. It ensures that each key in the CSV header
// corresponds to a field in the struct and vice versa. The function returns
// an error if there is a mismatch between the CSV keys and struct field names.
// The keys are compared against the struct's JSON tag, if present, otherwise
// the struct's field name is used.
func ValidateKeys(keys []string, dataSlice interface{}) error {
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

	return nil
}

// GetValueForFieldType is a helper function that takes a value, key and dataSlice
// and returns a JSON formatted string of the given value and key.
// The function is used in the Parse function to convert a CSV file to JSON.
// It iterates over the fields of the given struct and checks if the current
// field name matches the key. If it does, it appends the value to the output
// string with the correct JSON formatting. If the field is a string, it is
// enclosed in double quotes.
func GetValueForFieldType(value string, key string, dataSlice interface{}) string {
	output := ""

	structValue := reflect.ValueOf(dataSlice).Elem() // Dereference slice pointer
	structType := structValue.Type().Elem()          // Get underlying struct type

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}
		jsonTag, ok := field.Tag.Lookup("json")
		if !ok {
			jsonTag = field.Name
		}

		if jsonTag == key {
			if field.Type.Kind() == reflect.String {
				output += "\"" + key + "\": \"" + value + "\","
			} else {
				output += "\"" + key + "\": " + value + ","
			}
		}

	}

	return output
}
