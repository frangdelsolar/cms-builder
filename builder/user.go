package builder

import (
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name       string `json:"name"`
	Email      string `json:"email"`
	FirebaseId string `json:"firebase_id"`
}

// ID returns the ID of the SystemData as a string.
//
// Returns:
// - string: the ID of the SystemData.
func (u *User) GetIDString() string {
	return fmt.Sprint(u.ID)
}

// NameValidator validates the given name.
//
// Parameters:
// - name: the name to be validated.
//
// Returns:
// - error: an error if the name is empty, otherwise nil.
func NameValidator(fieldName string, instance map[string]interface{}, output *FieldValidationError) *FieldValidationError {
	name := fmt.Sprint(instance[fieldName])
	if name == "" {
		output.Error = fieldName + " cannot be empty"
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
func EmailValidator(fieldName string, instance map[string]interface{}, output *FieldValidationError) *FieldValidationError {
	email := fmt.Sprint(instance[fieldName])
	if email == "" {
		output.Error = fieldName + " cannot be empty"
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
