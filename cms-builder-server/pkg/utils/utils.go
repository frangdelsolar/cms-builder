package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gertd/go-pluralize"
)

func GetInterfaceName(input interface{}) (string, error) {
	modelType := reflect.TypeOf(input)

	if modelType == nil {
		return "", fmt.Errorf("model cannot be nil")
	}

	// If it's a pointer, get the element type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Ensure it's a struct before returning its name
	if modelType.Kind() != reflect.Struct {
		return "", fmt.Errorf("input must be a struct or a pointer to a struct")
	}

	return modelType.Name(), nil
}

// Pluralize returns the plural form of the given word.
//
// Parameters:
// - word: The word to pluralize.
//
// Returns:
// - string: The plural form of the word.
func Pluralize(word string) string {
	p := pluralize.NewClient()
	return p.Plural(word)
}

func SnakeCase(s string) string {
	re := regexp.MustCompile("(.)([A-Z][a-z]+)")
	s = re.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}

func KebabCase(s string) string {
	re := regexp.MustCompile("(.)([A-Z][a-z]+)")
	s = re.ReplaceAllString(s, "${1}-${2}")
	return strings.ToLower(s)
}

// CompareInterfaces takes two objects a and b, and returns a map of their differences.
// The function is used to compare two objects and return a map of the differences.
// The returned map will have keys that are the names of the fields in the object,
// and values that are slices of two elements: the value of the field in the first object,
// and the value of the field in the second object.
func CompareInterfaces(a, b interface{}) interface{} {
	if a == nil && b == nil {
		return map[string]interface{}{}
	}

	if a == nil || b == nil {
		return map[string]interface{}{"value": []interface{}{a, b}}
	}

	aJSON, err := json.Marshal(a)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	bJSON, err := json.Marshal(b)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	var aMap map[string]interface{}
	err = json.Unmarshal(aJSON, &aMap)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	var bMap map[string]interface{}
	err = json.Unmarshal(bJSON, &bMap)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	// Initialize the result map
	res := make(map[string]interface{})

	// Iterate over all keys in the first map
	for k, va := range aMap {
		// If the key is present in the second map
		if vb, ok := bMap[k]; ok {
			// If the values are not equal, add the difference to the result map
			if !reflect.DeepEqual(va, vb) {

				// Accont for time fields
				if reflect.TypeOf(va).Kind() == reflect.String && reflect.TypeOf(vb).Kind() == reflect.String {
					timeA, errA := time.Parse(time.RFC3339Nano, va.(string)) // Parse with nanosecond precision
					timeB, errB := time.Parse(time.RFC3339Nano, vb.(string))

					if errA == nil && errB == nil { // Both are valid times
						if !timeA.Equal(timeB) { // Use time.Equal for time comparison
							res[k] = []interface{}{va, vb}
						}
					} else { // Handle parsing errors or non-time strings
						if !reflect.DeepEqual(va, vb) { //Fallback to DeepEqual
							res[k] = []interface{}{va, vb}
						}
					}
				} else {
					// If the values are both maps, recursively call GetDiff
					// Add nil map checks here
					if va != nil && vb != nil && reflect.TypeOf(va).Kind() == reflect.Map && reflect.TypeOf(vb).Kind() == reflect.Map {
						nestedDiff := CompareInterfaces(va, vb)
						// Add interface{} nil check
						if len(nestedDiff.(map[string]interface{})) > 0 {
							res[k] = nestedDiff
						}
					} else {
						res[k] = []interface{}{va, vb}
					}
				}

			}
		} else {
			// If the key is not present in the second map, add the value from the first map to the result map
			res[k] = []interface{}{va, nil}
		}
	}

	// Iterate over all keys in the second map
	for k, vb := range bMap {
		// If the key is not present in the first map, add the value from the second map to the result map
		if _, ok := aMap[k]; !ok {
			res[k] = []interface{}{nil, vb}
		}
	}

	return res
}
