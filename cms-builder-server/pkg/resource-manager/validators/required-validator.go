package resourcemanager

import (
	"fmt"

	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
)

func RequiredValidator(fieldName string, instance rmTypes.EntityData, output *rmTypes.ValidationError) *rmTypes.ValidationError {
	value := instance[fieldName]

	if value == nil || value == "" {
		output.Error = fmt.Sprintf("%s is required", fieldName)
	}

	return output
}
