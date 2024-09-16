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

func (user *User) Validate() ValidationResult {

	var errors ValidationResult

	errors.Execute(NameValidator, user.Name)
	errors.Execute(EmailValidator, user.Email)

	return errors
}

// ID returns the ID of the user as a string.
//
// No parameters.
// Returns a string.
func (user *User) GetIDString() string {
	return fmt.Sprint(user.ID)
}

// NewUser creates a new user with the given name and email.
//
// Parameters:
// - name: the name of the user.
// - email: the email of the user.
// Returns:
// - *User: the newly created user.
// - error: an error if the user creation fails.
func NewUser(name string, email string) (*User, error) {

	user := &User{Name: name, Email: email}

	validationErrors := user.Validate()
	if len(validationErrors.Errors) > 0 {
		err := fmt.Errorf("validation errors: %v", validationErrors)
		return nil, err
	}

	return user, nil
}

// Update updates the name and email of a user.
//
// Parameters:
// - name: the new name of the user.
// - email: the new email of the user.
// Returns:
// - error: an error if the update fails.
func (user *User) Update(name string, email string) error {

	if err := NameValidator(name); err != (FieldValidationError{}) {
		return fmt.Errorf("validation errors: %v", err)
	}

	if err := EmailValidator(email); err != (FieldValidationError{}) {
		return fmt.Errorf("validation errors: %v", err)
	}

	user.Name = name
	user.Email = email

	return nil
}

// NameValidator validates the given name.
//
// Parameters:
// - name: the name to be validated.
//
// Returns:
// - error: an error if the name is empty, otherwise nil.
func NameValidator(name string) FieldValidationError {
	output := NewFieldValidationError(name)
	if name == "" {
		output.Error = "name cannot be empty"
		return output
	}

	return FieldValidationError{}
}

// EmailValidator validates the given email.
//
// Parameters:
// - email: the email to be validated.
//
// Returns:
// - error: an error if the email is empty or has an invalid format, otherwise nil.
func EmailValidator(email string) FieldValidationError {
	output := NewFieldValidationError(email)

	if email == "" {
		output.Error = "email cannot be empty"
		return output
	}

	emailRegex := `^[a-zA-Z0-9.!#$%&'*+/=?^_` + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`

	match, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		output.Error = err.Error()
		return output
	}

	if !match {
		output.Error = "email has an invalid format"
		return output
	}

	return FieldValidationError{}
}
