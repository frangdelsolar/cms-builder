package builder

import (
	"fmt"
	"regexp"
)

// TODO: Create tests for validators

type EntityData map[string]interface{}

// Validator is a function that validates a field value.
type Validator func(fieldName string, entity EntityData, output *ValidationError) *ValidationError

type ValidatorsList []Validator

type ValidatorsMap map[string]ValidatorsList

type ValidationError struct {
	Field string // The name of the field that failed validation
	Error string // The error message
}

// NewFieldValidationError creates a new FieldValidationError with the given field name and an empty error string.
func NewFieldValidationError(fieldName string) ValidationError {
	return ValidationError{
		Field: fieldName,
		Error: "",
	}
}

type ValidationResult struct {
	Errors []ValidationError // A list of field validation errors
}

// RequiredValidator is a validator that checks if a field value is not nil.
//
// Parameters:
//   - fieldName: The name of the field to be validated.
//   - instance: The instance to be validated.
//   - output: The output validation error.
//
// Returns:
//   - output: The output validation error with an error message if the value is nil.
func RequiredValidator(fieldName string, instance EntityData, output *ValidationError) *ValidationError {
	value := instance[fieldName]

	if value == nil || value == "" {
		output.Error = fmt.Sprintf("%s is required", fieldName)
	}

	return output
}

// EmailValidator validates the given email.
//
// Parameters:
// - email: the email to be validated.
//
// Returns:
// - error: an error if the email is empty or has an invalid format, otherwise nil.
func EmailValidator(fieldName string, instance EntityData, output *ValidationError) *ValidationError {
	email := fmt.Sprint(instance[fieldName])
	if email == "" {
		return nil
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
