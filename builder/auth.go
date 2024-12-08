package builder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// VerifyUser verifies the user based on the access token provided in the userIdToken parameter.
//
// The method verifies the token by calling VerifyIDToken on the Firebase Admin instance.
// If the token is valid, it retrieves the user record from the database and returns it.
// If the token is invalid, it returns an error.
func (b *Builder) VerifyUser(userIdToken string) (*User, error) {
	accessToken, err := b.Firebase.VerifyIDToken(context.Background(), userIdToken)
	if err != nil {
		log.Error().Err(err).Msg("Error verifying token")
		return nil, err
	}

	var localUser User

	q := "firebase_id = '" + accessToken.UID + "'"
	b.DB.DB.Where(q).First(&localUser)

	// Create user if firebase has it but not in database
	if localUser.ID == 0 {
		log.Info().Msg("User exists in Firebase but not in database. Will create it now")

		// create user in database
		localUser.Name = accessToken.Claims["name"].(string)
		localUser.Email = strings.ToLower(accessToken.Claims["email"].(string))
		localUser.FirebaseId = accessToken.UID
		localUser.Roles = string(VisitorRole)
		b.DB.Create(&localUser)
	}

	return &localUser, nil
}

// AuthMiddleware is a middleware function that verifies the user based on the
// access token provided in the Authorization header of the request. If the
// verification fails, it will return a 401 error. If the verification is
// successful, it will continue to the next handler in the chain, setting a
// "requested_by" header in the request with the ID of the verified user.
func (b *Builder) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Clear the headers in case someone else set them
		deleteHeader(requestedByParamKey, r)
		deleteHeader(authParamKey, r)

		accessToken := GetAccessTokenFromRequest(r)
		localUser, err := b.VerifyUser(accessToken)
		if err != nil {
			log.Error().Err(err).Msg("Error verifying user")
			SendJsonResponse(w, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		if localUser == nil {
			SendJsonResponse(w, http.StatusUnauthorized, fmt.Errorf("User not found"), "Unauthorized")
			return
		}

		setHeader(requestedByParamKey, localUser.GetIDString(), r)
		setHeader(authParamKey, "true", r)

		log.Info().Interface("User", localUser).Msg("Logging in user")

		next.ServeHTTP(w, r)
	})
}

// RegisterUserController handles the endpoint to register a new user. The endpoint
// expects a POST request with a JSON body containing the name, email and password
// of the user to register. The function will return a 400 error if the request body
// is not valid JSON, or if the request method is not POST.
//
// The function will also return a 500 error if there is an error registering the user
// in Firebase, or if there is an error creating the user in the local database.
//
// The function will also set the requested_by header to the ID of the newly created
// user.
func (b *Builder) RegisterVisitorController(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendJsonResponse(w, http.StatusMethodNotAllowed, nil, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Error reading request body: %s", err.Error())
		SendJsonResponse(w, http.StatusInternalServerError, nil, msg)
		return
	}

	var input RegisterUserInput
	err = json.Unmarshal(body, &input)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling request body: %s", err.Error())
		SendJsonResponse(w, http.StatusBadRequest, nil, msg)
		return
	}

	user, err := b.CreateUserWithRole(input, VisitorRole)
	if err != nil {
		msg := fmt.Sprintf("Error creating user: %s", err.Error())
		SendJsonResponse(w, http.StatusInternalServerError, nil, msg)
		return
	}

	// Prevent sending the firebaseId to the client
	user.FirebaseId = ""
	SendJsonResponse(w, http.StatusOK, user, "User registered successfully")
}

// CreateUserWithRole creates a new user in Firebase with the given name, email, and password, and also
// creates a new user in the local database with the given role. If the user already exists in Firebase,
// it will add the user to the local database. If the user already exists in the local database, it will
// return an error.
func (b *Builder) CreateUserWithRole(input RegisterUserInput, role Role) (*User, error) {
	ctx := context.Background()
	var fbUserId string

	fbUser, err := b.Firebase.RegisterUser(ctx, input)
	if err != nil {
		msg := fmt.Sprintf("Error registering user: %s", err.Error())

		if strings.Contains(err.Error(), "EMAIL_EXISTS") {
			existingFbUser, err := b.Firebase.Client.GetUserByEmail(ctx, input.Email)
			if err != nil {
				msg := fmt.Sprintf("Error getting user by email: %s", err.Error())
				return nil, fmt.Errorf("%s", msg)
			}
			log.Warn().Msg("User already exists in Firebase. Will add it to database")
			fbUserId = existingFbUser.UID
		} else {
			return nil, fmt.Errorf("%s", msg)
		}
	} else {
		fbUserId = fbUser.UID
	}

	// Check if there is a user with the same fbUserId in the database
	var existingUser User
	q := "firebase_id = '" + fbUserId + "'"
	err = b.DB.DB.Where(q).First(&existingUser).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("error getting user from database")
		}
	}

	if existingUser != (User{}) {
		log.Warn().Msg("User already exists in database.")
		return &existingUser, nil
	}

	// Create user in database
	user := User{
		Name:       input.Name,
		Email:      strings.ToLower(input.Email),
		FirebaseId: fbUserId,
		Roles:      string(role),
	}
	b.DB.Create(&user)

	if user.ID == 0 {
		return nil, fmt.Errorf("error creating user in database")
	}

	return &user, nil
}
