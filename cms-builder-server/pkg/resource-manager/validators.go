package resourcemanager

import (
	"fmt"
	"regexp"
)

type EntityData map[string]interface{}

type Validator func(fieldName string, entity EntityData, output *ValidationError) *ValidationError

type ValidatorsList []Validator

type ValidatorsMap map[string]ValidatorsList

type ValidationError struct {
	Field string // The name of the field that failed validation
	Error string // The error message
}

func NewFieldValidationError(fieldName string) ValidationError {
	return ValidationError{
		Field: fieldName,
		Error: "",
	}
}

type ValidationResult struct {
	Errors []ValidationError // A list of field validation errors
}

func RequiredValidator(fieldName string, instance EntityData, output *ValidationError) *ValidationError {
	value := instance[fieldName]

	if value == nil || value == "" {
		output.Error = fmt.Sprintf("%s is required", fieldName)
	}

	return output
}

func EmailValidator(fieldName string, instance EntityData, output *ValidationError) *ValidationError {
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
