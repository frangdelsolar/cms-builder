package auth

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
)

var (
	ErrorRoleAlreadyAssigned = fmt.Errorf("role already assigned to user")
)

type User struct {
	gorm.Model
	ID         uint   `gorm:"primaryKey" json:"ID"`
	Name       string `json:"name"`
	Email      string `gorm:"unique" json:"email"`
	FirebaseId string `json:"firebaseId"`
	Roles      string `json:"roles"` // comma-separated list of roles e.g. "admin,visitor"
}

// ID returns the ID of the SystemData as a string.
//
// Returns:
// - string: the ID of the SystemData.
func (u *User) StringID() string {
	return fmt.Sprint(u.ID)
}

// GetRoles parses the comma-separated list of roles from the User's Roles field and
// returns a slice of authPkg.Role objects.
//
// Returns:
// - []authPkg.Role: a slice of authPkg.Role objects, or an empty slice if the Roles field is empty.
func (u *User) GetRoles() []authTypes.Role {
	roles := []authTypes.Role{}

	// Trim leading/trailing spaces and check if the Roles field is empty
	trimmedRoles := strings.TrimSpace(u.Roles)
	if trimmedRoles == "" {
		return roles
	}

	// Split the roles string and process each role
	for _, role := range strings.Split(trimmedRoles, ",") {
		role = strings.TrimSpace(role)
		if role != "" { // Skip empty roles
			roles = append(roles, authTypes.Role(role))
		}
	}
	return roles
}

// SetRole adds a role to the User's Roles field. If the role is already present,
// it does nothing. If the Roles field is empty, it sets the field to the given role.
//
// Parameters:
// - role: the authPkg.Role to be added to the User's Roles field.
func (u *User) SetRole(role authTypes.Role) error {
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
func (u *User) RemoveRole(role authTypes.Role) {
	roles := strings.Split(u.Roles, ",")
	for i, r := range roles {
		if r == string(role) {
			roles = append(roles[:i], roles[i+1:]...)
			break
		}
	}
	u.Roles = strings.Join(roles, ",")
}

func (u *User) HasRole(role authTypes.Role) bool {
	roles := u.GetRoles()
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
