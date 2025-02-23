package clients

import (
	"context"
	"encoding/base64"
	"errors"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
)

var (
	ErrFirebaseNotInitialized = errors.New("firebase not initialized")
)

type FirebaseConfig struct {
	Secret string
}

type FirebaseManager struct {
	*firebase.App
	*auth.Client
}

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
type FirebaseUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterUser registers a new user in Firebase with the given name, email, and password.
//
// Parameters:
// - ctx: The context to use for the operation.
// - input: The input data for the user to create, containing the display name, email and password.
//
// Returns:
// - *auth.UserRecord: The user record of the newly created user.
// - error: An error if the user creation fails.
func (fa *FirebaseManager) RegisterUser(ctx context.Context, input FirebaseUserInput) (*auth.UserRecord, error) {
	userToCreate := &auth.UserToCreate{}
	userToCreate.DisplayName(input.Name)
	userToCreate.Email(input.Email)
	userToCreate.Password(input.Password)

	return fa.CreateUser(ctx, userToCreate)
}

// RollbackUserRegistration rolls back a user registration by deleting the user with the given UID in Firebase.
//
// Parameters:
// - ctx: The context to use for the operation.
// - uid: The UID of the user to delete.
//
// Returns:
// - error: An error if the user deletion fails.
func (fa *FirebaseManager) RollbackUserRegistration(ctx context.Context, uid string) error {
	return fa.DeleteUser(ctx, uid)
}

// VerifyIDToken verifies the Firebase ID token provided by the user.
//
// Parameters:
// - ctx: The context to use for the operation.
// - idToken: The Firebase ID token obtained from the user.
//
// Returns:
// - *auth.Token: The decoded Firebase ID token object if valid.
// - error: An error if the token verification fails.
func (fa *FirebaseManager) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return fa.Client.VerifyIDToken(ctx, idToken)
}

// NewFirebaseAdmin creates a new instance of FirebaseAdmin using the provided AuthConfig.
//
// Parameters:
// - config: The AuthConfig containing the credentials file path.
//
// Returns:
// - *FirebaseAdmin: A pointer to the newly created FirebaseAdmin instance.
// - error: An error if there was a problem initializing the Firebase app or the authentication client.
func NewFirebaseAdmin(cfg *FirebaseConfig) (*FirebaseManager, error) {

	var err error
	output := FirebaseManager{}

	decoded, err := base64.StdEncoding.DecodeString(cfg.Secret)
	if err != nil {
		log.Err(err).Msg("error decoding firebase secret")
		return nil, err
	}

	creds := option.WithCredentialsJSON(decoded)

	app, err := firebase.NewApp(context.Background(), nil, creds)
	if err != nil {
		log.Err(err).Msg("error initializing firebase")
		return &output, err
	}

	cli, err := app.Auth(context.Background())
	if err != nil {
		log.Err(err).Msg("error initializing firebase")
		return &output, err
	}

	output.App = app
	output.Client = cli

	return &output, nil
}
