package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

type CRUDAction string

const (
	CreateCRUDAction CRUDAction = "created"
	UpdateCRUDAction CRUDAction = "updated"
	DeleteCRUDAction CRUDAction = "deleted"
)

type HistoryEntry struct {
	gorm.Model
	User         *User      `json:"user"`
	UserId       string     `gorm:"foreignKey:UserId" json:"userId"`
	Username     string     `json:"username"`
	Action       CRUDAction `json:"action"`
	ResourceName string     `json:"resourceName"`
	ResourceId   string     `json:"resourceId"`
	Timestamp    string     `gorm:"type:timestamp" json:"timestamp"`
	Detail       string     `json:"detail"`
	RequestId    string     `json:"requestId"`
}

// NewLogHistoryEntry takes an action of type CRUDAction, a user ID, and an object, and returns a pointer to a HistoryEntry and an error.
// The HistoryEntry is generated by marshaling the object to JSON, and extracting the ID from it if it exists.
// The object is expected to be a struct with a JSON tag for the ID field named "ID".
// The function returns an error if the object cannot be marshaled or unmarshaled to JSON.
// The function uses the GetStructName function to get the name of the struct from the object passed in.
func NewLogHistoryEntry(action CRUDAction, user *User, object interface{}, difference interface{}, requestId string) (*HistoryEntry, error) {

	name := utils.GetStructName(object)
	jsonData, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	var objJson map[string]interface{}
	err = json.Unmarshal(jsonData, &objJson)
	if err != nil {
		return nil, err
	}

	resourceId := ""
	if objJson["ID"] != nil {
		resourceId = fmt.Sprintf("%v", objJson["ID"])
	}

	detail := ""
	if action == CreateCRUDAction {
		detail = string(jsonData)
	} else if action == UpdateCRUDAction {

		differenceJSON, err := json.Marshal(difference)
		if err != nil {
			return nil, err
		}

		detail = string(differenceJSON)
	}

	historyEntry := &HistoryEntry{
		Action:       action,
		UserId:       user.StringID(),
		Username:     user.Email,
		ResourceId:   resourceId,
		ResourceName: name,
		Timestamp:    time.Now().Format(time.RFC3339Nano),
		Detail:       detail,
		RequestId:    requestId,
	}

	return historyEntry, nil
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
