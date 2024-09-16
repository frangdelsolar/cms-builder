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

func (user *User) Validate() []error {

	errors := make([]error, 0)

	if err := NameValidator(user.Name); err != nil {
		errors = append(errors, err)
	}
	if err := EmailValidator(user.Email); err != nil {
		errors = append(errors, err)
	}

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

	if err := NameValidator(name); err != nil {
		return nil, err
	}
	if err := EmailValidator(email); err != nil {
		return nil, err
	}

	return &User{Name: name, Email: email}, nil
}

// Update updates the name and email of a user.
//
// Parameters:
// - name: the new name of the user.
// - email: the new email of the user.
// Returns:
// - error: an error if the update fails.
func (user *User) Update(name string, email string) error {
	if err := NameValidator(name); err != nil {
		return err
	}
	if err := EmailValidator(email); err != nil {
		return err
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
func NameValidator(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

// EmailValidator validates the given email.
//
// Parameters:
// - email: the email to be validated.
//
// Returns:
// - error: an error if the email is empty or has an invalid format, otherwise nil.
func EmailValidator(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	emailRegex := `^[a-zA-Z0-9.!#$%&'*+/=?^_` + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`

	match, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		return fmt.Errorf("error compiling email regex: %w", err)
	}

	if !match {
		return fmt.Errorf("invalid email format")
	}

	return nil
}
