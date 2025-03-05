package validators

import (
	"fmt"
	"regexp"

	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
)

func EmailValidator(fieldName string, instance rmTypes.EntityData, output *rmTypes.ValidationError) *rmTypes.ValidationError {
	email := fmt.Sprint(instance[fieldName])
	if email == "" {
		return output
	}
	emailRegex := `^[a-zA-Z0-9.!#$%&'*+/=?^_` + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`

	match, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		output.Error = err.Error()
	}

	if !match {
		output.Error = "email has an invalid format"
	}

	return output
}
