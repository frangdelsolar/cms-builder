package builder

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrorRoleAlreadyAssigned = fmt.Errorf("role already assigned to user")
)

type User struct {
	*gorm.Model
	ID   uint   `gorm:"primaryKey" json:"ID"`
	Name string `json:"name"`
	// Email      string `gorm:"unique" json:"email"`
	Email      string `json:"email"` //FIXME
	FirebaseId string `json:"firebaseId"`
	Roles      string `json:"roles"` // comma-separated list of roles e.g. "admin,visitor"
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

// SetRole adds a role to the User's Roles field. If the role is already present,
// it does nothing. If the Roles field is empty, it sets the field to the given role.
//
// Parameters:
// - role: the Role to be added to the User's Roles field.
func (u *User) SetRole(role Role) error {
	if u.Roles == "" {
		u.Roles = string(role)
	} else {
		if strings.Contains(u.Roles, string(role)) {
			return ErrorRoleAlreadyAssigned
		}
		u.Roles += "," + string(role)
	}
	return nil
}

// RemoveRole removes a role from the User's Roles field. If the role is not present, it has no effect.
func (u *User) RemoveRole(role Role) {
	roles := strings.Split(u.Roles, ",")
	for i, r := range roles {
		if r == string(role) {
			roles = append(roles[:i], roles[i+1:]...)
			break
		}
	}
	u.Roles = strings.Join(roles, ",")
}
