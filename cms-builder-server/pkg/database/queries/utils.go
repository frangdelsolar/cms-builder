package queries

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

// NewDatabaseLogEntry takes an action of type CRUDAction, a user ID, and an object, and returns a pointer to a HistoryEntry and an error.
// The HistoryEntry is generated by marshaling the object to JSON, and extracting the ID from it if it exists.
// The object is expected to be a struct with a JSON tag for the ID field named "ID".
// The function returns an error if the object cannot be marshaled or unmarshaled to JSON.
// The function uses the GetStructName function to get the name of the struct from the object passed in.
func NewDatabaseLogEntry(action CRUDAction, user *models.User, object interface{}, difference interface{}, traceId string) (*DatabaseLog, error) {

	name, err := utils.GetInterfaceName(object)
	if err != nil {
		return nil, err
	}

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

	historyEntry := &DatabaseLog{
		Action:       action,
		UserId:       user.StringID(),
		Username:     user.Email,
		ResourceId:   resourceId,
		ResourceName: name,
		Timestamp:    time.Now().Format(time.RFC3339Nano),
		Detail:       detail,
		TraceId:      traceId,
	}

	return historyEntry, nil
}
