package builder

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	*gorm.Model
	Name       string `json:"name"`
	Email      string `gorm:"unique" json:"email"`
	FirebaseId string `json:"firebase_id"`
	Roles      string `json:"roles"` // comma-separated list of roles
}

// ID returns the ID of the SystemData as a string.
//
// Returns:
// - string: the ID of the SystemData.
func (u *User) GetIDString() string {
	return fmt.Sprint(u.ID)
}

// GetRoles parses the comma-separated list of roles from the User's Roles field and
// returns a slice of Role objects.
//
// Returns:
// - []Role: a slice of Role objects, or an empty slice if the Roles field is empty.
func (u *User) GetRoles() []Role {
	roles := []Role{}

	for _, role := range strings.Split(u.Roles, ",") {
		role = strings.TrimSpace(role)
		roles = append(roles, Role(role))
	}
	return roles
}
