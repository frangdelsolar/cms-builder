package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const GodTokenHeader = "X-God-Token"

// TODO: THIS FILE NEEDS SERIOUS REFACTORING

func VerifyUser(userIdToken string, firebase *clients.FirebaseManager, db *database.Database, systemUser *models.User, requestId string) (*models.User, error) {
	accessToken, err := firebase.VerifyIDToken(context.Background(), userIdToken)
	if err != nil {
		log.Error().Err(err).Msg("Error verifying token")
		return nil, err
	}

	var localUser models.User = models.User{}

	if err != nil {
		log.Error().Err(err).Msg("Error finding user in database")
	}

	if localUser.ID == 0 {
		claims := accessToken.Claims

		// Safer way to get the name, handling missing claims and type issues
		name, ok := claims["name"].(string)
		if !ok {
			name, ok = claims["displayName"].(string) // Try displayName as fallback
			if !ok {
				log.Warn().Msg("Name claim not found in token")
				name = "" // Or a suitable default like "Unknown models.User"
			}
		}

		// Safer way to get the email, handling missing claims and type issues
		email, ok := claims["email"].(string)
		if !ok {
			log.Warn().Msg("Email claim not found in token")
			email = "" //  Consider a default like "no-email@example.com" or handle differently
		}

		localUser.Name = name
		localUser.Email = strings.ToLower(email)
		localUser.FirebaseId = accessToken.UID
		localUser.Roles = string(models.VisitorRole)

		res := queries.Create(db, &localUser, systemUser, requestId)
		if res.Error != nil { // Check for errors during creation
			err = fmt.Errorf("error creating user in database: %v", res.Error)
			return nil, err // Return the error if user creation fails
		}
	}

	return &localUser, nil
}

func VerifyGodUser(envToken string, requestToken string) bool {
	return requestToken != envToken && requestToken != ""
}

// AuthMiddleware is a middleware function that verifies the user based on the
// access token provided in the Authorization header of the request. If the
// verification fails, it will return a 401 error. If the verification is
// successful, it will continue to the next handler in the chain, setting a
// "requested_by" header in the request with the ID of the verified user.
func AuthMiddleware(envGodToken string, godUser *models.User, firebase *clients.FirebaseManager, db *database.Database, systemUser *models.User) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Clear the headers in case someone else set them
			r.Header.Del(RequestedByParamKey.S())
			r.Header.Del(RolesParamKey.S())

			// Check if the request has a god token
			headerGodToken := r.Header.Get(GodTokenHeader)

			accessToken := GetRequestAccessToken(r)
			requestId := GetRequestId(r)
			log := GetRequestLogger(r)

			var localUser *models.User
			if headerGodToken != "" && VerifyGodUser(envGodToken, headerGodToken) {
				localUser = godUser
			}

			var err error
			if localUser.ID == 0 && accessToken != "" {
				localUser, err = VerifyUser(accessToken, firebase, db, systemUser, requestId)
				if err != nil {
					log.Error().Err(err).Msg("Error verifying user. models.User may not be authenticated")
				}
			}

			if localUser != nil {
				r = r.WithContext(context.WithValue(r.Context(), CtxRequestIsAuth, true))
				r = r.WithContext(context.WithValue(r.Context(), CtxRequestUser, localUser))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TODO: this is disgusting -> REFACTOR
// CreateUserWithRole creates a new user in Firebase with the given name, email, and password, and also
// creates a new user in the local database with the given role. If the user already exists in Firebase,
// it will add the user to the local database. If the user already exists in the local database, it will
// return an error.
func CreateUserWithRole(input models.RegisterUserInput, firebase *clients.FirebaseManager, db *database.Database, systemUser *models.User, requestId string) (*models.User, error) {
	ctx := context.Background()

	registerFirebase := input.RegisterFirebase
	firebaseInput := clients.FirebaseUserInput{
		Email:    input.Email,
		Name:     input.Name,
		Password: input.Password,
	}

	roles := ""
	for _, role := range input.Roles {
		roles += string(role) + ","
	}
	roles = roles[:len(roles)-1]

	var fbUserId string
	if registerFirebase {
		fbUser, err := firebase.RegisterUser(ctx, firebaseInput)
		if err != nil {
			msg := fmt.Sprintf("Error registering user: %s", err.Error())

			if strings.Contains(err.Error(), "EMAIL_EXISTS") {
				existingFbUser, err := firebase.Client.GetUserByEmail(ctx, input.Email)
				if err != nil {
					msg := fmt.Sprintf("Error getting user by email: %s", err.Error())
					return nil, fmt.Errorf("%s", msg)
				}
				// log.Warn().Msg("models.User already exists in Firebase. Will add it to database")
				fbUserId = existingFbUser.UID
			} else {
				return nil, fmt.Errorf("%s", msg)
			}
		} else {
			fbUserId = fbUser.UID
		}

		// Check if there is a user with the same fbUserId in the database
		var existingUser models.User
		q := "firebase_id = '" + fbUserId + "'"
		err = db.DB.Where(q).First(&existingUser).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("error getting user from database")
			}
		}

		if existingUser != (models.User{}) {
			// log.Warn().Msg("models.User already exists in database.")
			return &existingUser, nil
		}
	}

	// Check if there is a user with the same email in the database
	var existingUser models.User
	q := "email = '" + strings.ToLower(input.Email) + "'"
	err := db.DB.Where(q).First(&existingUser).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("error getting user from database")
		}
	}

	if existingUser != (models.User{}) {
		// log.Warn().Msg("models.User already exists in database.")
		return &existingUser, nil
	}

	// Create user in database
	user := models.User{
		Name:       input.Name,
		Email:      strings.ToLower(input.Email),
		FirebaseId: fbUserId,
		Roles:      roles,
	}

	err = queries.Create(db, &user, systemUser, requestId).Error
	if err != nil {
		return nil, fmt.Errorf("error creating user in database")
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("error creating user in database")
	}

	return &user, nil
}
