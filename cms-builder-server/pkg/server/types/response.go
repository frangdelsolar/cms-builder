package types

import (
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
