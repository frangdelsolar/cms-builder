package auth

// RegisterUser registers a new user in Firebase with the given name, email, and password.
//
// Parameters:
// - name: the display name of the user.
// - email: the email address of the user.
// - password: the password for the user.
//
// Returns:
// - *auth.UserRecord: the user record of the newly created user.
// - error: an error if the user creation fails.
type RegisterUserInput struct {
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	Roles            []Role `json:"roles"`
	RegisterFirebase bool
}
