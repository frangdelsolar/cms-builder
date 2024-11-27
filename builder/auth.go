package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// VerifyUser verifies the user based on the access token provided in the userIdToken parameter.
//
// The method verifies the token by calling VerifyIDToken on the Firebase Admin instance.
// If the token is valid, it retrieves the user record from the database and returns it.
// If the token is invalid, it returns an error.
func (b *Builder) VerifyUser(userIdToken string) (*User, error) {
	// verify token
	firebase, err := b.GetFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Error getting firebase")
		return nil, err
	}

	accessToken, err := firebase.VerifyIDToken(context.Background(), userIdToken)
	if err != nil {
		log.Error().Err(err).Msg("Error verifying token")
		return nil, err
	}

	var localUser User

	q := "firebase_id = '" + accessToken.UID + "'"
	b.db.Find(&localUser, q)

	// Create user if firebase has it but not in database
	if localUser.ID == 0 {
		log.Info().Msg("User exists in Firebase but not in database. Will create it now")

		// create user in database
		localUser.Name = accessToken.Claims["name"].(string)
		localUser.Email = accessToken.Claims["email"].(string)
		localUser.FirebaseId = accessToken.UID
		b.db.Create(&localUser)
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

		r.Header.Set("auth", "true") // this is set only for testing purposes

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
func (b *Builder) RegisterUserController(w http.ResponseWriter, r *http.Request) {
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
	fb, err := b.GetFirebase()
	if err != nil {
		msg := fmt.Sprintf("Error getting firebase: %s", err.Error())
		SendJsonResponse(w, http.StatusInternalServerError, nil, msg)
		return
	}
	var fbUserId string
	fbUser, err := fb.RegisterUser(r.Context(), input)
	if err != nil {
		msg := fmt.Sprintf("Error registering user: %s", err.Error())

		if strings.Contains(err.Error(), "EMAIL_EXISTS") {
			existingFbUser, err := b.firebase.Client.GetUserByEmail(r.Context(), input.Email)
			if err != nil {
				msg := fmt.Sprintf("Error getting user by email: %s", err.Error())
				SendJsonResponse(w, http.StatusInternalServerError, nil, msg)
				return
			}
			log.Debug().Msg("User already exists in Firebase. Will add it to database")
			fbUserId = existingFbUser.UID
		} else {
			SendJsonResponse(w, http.StatusInternalServerError, nil, msg)
			return
		}
	} else {
		fbUserId = fbUser.UID
	}

	// Check if there is a user with the same fbUserId in the database
	var existingUser User
	q := "firebase_id = '" + fbUserId + "'"
	b.db.Find(&existingUser, q)
	if existingUser.ID != 0 {
		log.Debug().Msg("User already exists in database.")
		// Prevent sending the firebaseId to the client
		existingUser.FirebaseId = ""
		SendJsonResponse(w, http.StatusOK, existingUser, "User already registered")
		return
	}

	// Create user in database
	user := User{Name: input.Name, Email: input.Email, FirebaseId: fbUserId}
	b.db.Create(&user)

	if user.ID == 0 {
		SendJsonResponse(w, http.StatusInternalServerError, nil, "Error creating user in database")
		return
	}

	// Prevent sending the firebaseId to the client
	user.FirebaseId = ""
	SendJsonResponse(w, http.StatusOK, user, "User registered successfully")
}

// GetAccessTokenFromRequest extracts the access token from the Authorization header of the given request.
// The header should be in the format "Bearer <token>".
// If the token is not found, it returns an empty string.
func GetAccessTokenFromRequest(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header != "" {
		token := strings.Split(header, " ")[1]
		if token != "" {
			return token
		}
	}
	return ""
}
