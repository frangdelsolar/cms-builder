package server

import (
	"encoding/json"
	"fmt"

	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
)

// Following these standards
// https://github.com/omniti-labs/jsend
// https://medium.com/@bojanmajed/standard-json-api-response-format-c6c1aabcaa6d

type Response struct {
	Success    bool                `json:"success"`
	Data       interface{}         `json:"data"`
	Message    string              `json:"message"`
	Pagination *dbTypes.Pagination `json:"pagination"`
}

func (r *Response) ParseResponseData(dataStruct any) error {
	// Convert the Data map to JSON bytes
	dataBytes, err := json.Marshal(r.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal Data map: %v", err)
	}

	// Unmarshal the JSON bytes into the provided struct
	err = json.Unmarshal(dataBytes, dataStruct)
	if err != nil {
		return fmt.Errorf("failed to unmarshal into struct: %v", err)
	}

	return nil
}
