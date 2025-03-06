package resourcemanager

import (
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
)

func NewFieldValidationError(fieldName string) rmTypes.ValidationError {
	return rmTypes.ValidationError{
		Field: fieldName,
		Error: "",
	}
}
