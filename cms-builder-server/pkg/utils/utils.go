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

				// Handle time fields with proper nil checks
				if va != nil && vb != nil {
					vaStr, aIsStr := va.(string)
					vbStr, bIsStr := vb.(string)
					if aIsStr && bIsStr {
						timeA, errA := time.Parse(time.RFC3339Nano, vaStr)
						timeB, errB := time.Parse(time.RFC3339Nano, vbStr)

						if errA == nil && errB == nil { // Both are valid times
							if !timeA.Equal(timeB) {
								res[k] = []interface{}{va, vb}
							}
							continue
						}
					}
				}

				// If the values are both maps, recursively call CompareInterfaces
				if va != nil && vb != nil && reflect.TypeOf(va).Kind() == reflect.Map && reflect.TypeOf(vb).Kind() == reflect.Map {
					nestedDiff := CompareInterfaces(va, vb)
					if nestedMap, ok := nestedDiff.(map[string]interface{}); ok && len(nestedMap) > 0 {
						res[k] = nestedDiff
					}
				} else {
					res[k] = []interface{}{va, vb}
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

func Deepcopy(a interface{}) interface{} {
	if a == nil {
		return nil
	}

	original := reflect.ValueOf(a)
	if !original.IsValid() { // Check if the reflect.Value is valid.
		return nil
	}

	originalType := original.Type()

	// Handle pointers by copying the underlying value
	if originalType.Kind() == reflect.Ptr {
		if original.IsNil() { // Check if the pointer is nil.
			return nil
		}
		original = original.Elem()
		originalType = original.Type()
	}

	switch originalType.Kind() {
	case reflect.Slice, reflect.Array:
		newSlice := reflect.MakeSlice(originalType, original.Len(), original.Cap())
		reflect.Copy(newSlice, original)
		return newSlice.Interface()
	case reflect.Map:
		newMap := reflect.MakeMap(originalType)
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			// Recursively copy map values.
			newMap.SetMapIndex(key, reflect.ValueOf(Deepcopy(originalValue.Interface())))
		}
		return newMap.Interface()
	case reflect.Struct:
		newStruct := reflect.New(originalType).Elem()
		for i := 0; i < original.NumField(); i++ {
			field := original.Field(i)
			// Recursively copy struct fields.
			newStruct.Field(i).Set(reflect.ValueOf(Deepcopy(field.Interface())))
		}
		return newStruct.Interface()
	case reflect.Ptr:
		//This case should not be reached due to the initial pointer handling.
		// However, it is included for completeness.
		return Deepcopy(original.Interface())
	default:
		// For basic types, just return the original value.
		return a
	}
}
