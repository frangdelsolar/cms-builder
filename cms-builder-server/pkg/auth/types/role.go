package types

// Role represents a user role.
type Role string

// S converts a Role to its string representation.
func (r Role) S() string {
	return string(r)
}
