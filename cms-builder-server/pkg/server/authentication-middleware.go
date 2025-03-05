package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"

	firebaseAuth "firebase.google.com/go/auth"

	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"gorm.io/gorm"
)

const GodTokenHeader = "X-God-Token"

func VerifyUser(userIdToken string, firebase *clients.FirebaseManager, db *database.Database, systemUser *models.User, requestId string, log *loggerTypes.Logger) (*models.User, error) {

	accessToken, err := firebase.VerifyIDToken(context.Background(), userIdToken)
	if err != nil {
		log.Error().Err(err).Msg("Error verifying token")
		return nil, err
	}

	localUser := models.User{}
	filters := map[string]interface{}{
		"firebase_id": accessToken.UID,
	}

	err = dbQueries.FindOne(context.Background(), log, db, &localUser, filters)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Need
			log.Warn().Msg("User is firebase user but not in database")
			return RegisterFirebaseUserInDatabase(accessToken, firebase, db, systemUser, requestId, log)
		} else {
			log.Error().Err(err).Msg("Error finding user in database")
			return nil, err
		}
	}

	return &localUser, nil
}

func RegisterFirebaseUserInDatabase(accessToken *firebaseAuth.Token, firebase *clients.FirebaseManager, db *database.Database, systemUser *models.User, requestId string, log *loggerTypes.Logger) (*models.User, error) {
	// Extract name and email from the access token's claims
	name, _ := accessToken.Claims["name"].(string)    // Name might not always be present
	email, ok := accessToken.Claims["email"].(string) // Email is usually required

	// Check if email is present and valid
	if !ok || email == "" {
		return nil, fmt.Errorf("email claim is missing or invalid in the access token")
	}

	// Set default name if not provided
	if name == "" {
		name = "No Name" // Or any other default value
	}

	roles := string(models.VisitorRole)

	// Create a local user object
	localUser := &models.User{
		Name:  name,
		Email: email,
		Roles: roles, // Assign roles if applicable
	}

	// Save the user to the database
	err := dbQueries.Create(context.Background(), log, db, localUser, systemUser, requestId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	return localUser, nil
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

			localUser := &models.User{}
			if headerGodToken != "" && VerifyGodUser(envGodToken, headerGodToken) {
				localUser = godUser
			}

			var err error
			if localUser.ID == 0 && accessToken != "" {
				localUser, err = VerifyUser(accessToken, firebase, db, systemUser, requestId, log)
				if err != nil {
					log.Error().Err(err).Msg("Error verifying user. User may not be authenticated")
				}
			}

			if localUser != nil {

				// Create a new context with both values
				ctx := r.Context()
				ctx = context.WithValue(ctx, CtxRequestIsAuth, true)
				ctx = context.WithValue(ctx, CtxRequestUser, localUser)

				// Update the request with the new context
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func FormatRoles(roles []models.Role) string {
	rolesStr := ""
	for _, role := range roles {
		rolesStr += role.S() + ","
	}

	if len(rolesStr) > 0 {
		rolesStr = rolesStr[:len(rolesStr)-1]
	}

	return rolesStr
}

func CreateUserWithRole(input models.RegisterUserInput, firebase *clients.FirebaseManager, db *database.Database, systemUser *models.User, requestId string, log *loggerTypes.Logger) (*models.User, error) {

	log.Debug().Interface("input", input).Msg("Creating user with role")

	ctx := context.Background()

	// Normalize email to lowercase
	input.Email = strings.ToLower(input.Email)

	// Convert roles slice to a comma-separated string
	roles := FormatRoles(input.Roles)

	// Register user in Firebase if required
	fbUserId, err := registerOrGetFirebaseUser(ctx, firebase, input, log)
	if err != nil {
		return nil, fmt.Errorf("failed to register or get Firebase user: %w", err)
	}

	// Check if a user with the same email already exists in the database
	existingUser, err := findUserByEmail(db, input.Email, log)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user by email: %w", err)
	}

	// If the user exists, update Firebase ID if necessary and return
	if existingUser != nil {
		return handleExistingUser(existingUser, fbUserId, db, systemUser, requestId, log)
	}

	// Create the user in the database
	newUser := models.User{
		Name:       input.Name,
		Email:      input.Email,
		FirebaseId: fbUserId,
		Roles:      roles,
	}

	if err := dbQueries.Create(context.Background(), log, db, &newUser, systemUser, requestId); err != nil {
		log.Error().Err(err).Msg("Failed to create user in database")
		return nil, fmt.Errorf("failed to create user in database: %w", err)
	}

	log.Info().Interface("user", newUser).Msg("User created successfully")
	return &newUser, nil
}

// Helper function to register or get an existing Firebase user
func registerOrGetFirebaseUser(ctx context.Context, firebase *clients.FirebaseManager, input models.RegisterUserInput, log *loggerTypes.Logger) (string, error) {
	if !input.RegisterFirebase {
		return "", nil
	}

	log.Info().Str("email", input.Email).Msg("Registering user in Firebase")

	firebaseInput := clients.FirebaseUserInput{
		Email:    input.Email,
		Name:     input.Name,
		Password: input.Password,
	}

	fbUser, err := firebase.GetOrCreateUser(ctx, firebaseInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to register or get Firebase user")
		return "", fmt.Errorf("failed to register or get Firebase user: %w", err)
	}

	log.Info().Interface("user", fbUser).Msg("User registered in Firebase")
	return fbUser.UID, nil
}

// Helper function to find a user by email
func findUserByEmail(db *database.Database, email string, log *loggerTypes.Logger) (*models.User, error) {
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		log.Error().Err(err).Msg("Failed to query user from database")
		return nil, fmt.Errorf("failed to query user from database: %w", err)
	}

	return &user, nil
}

// Helper function to handle existing user (update Firebase ID if necessary)
func handleExistingUser(existingUser *models.User, fbUserId string, db *database.Database, systemUser *models.User, requestId string, log *loggerTypes.Logger) (*models.User, error) {
	if existingUser.FirebaseId == fbUserId {
		if fbUserId != "" {
			log.Info().Msg("User already exists in database with matching Firebase ID")
		}
		return existingUser, nil
	}

	previousState := *existingUser
	differences := utils.CompareInterfaces(previousState, existingUser)

	existingUser.FirebaseId = fbUserId

	if err := dbQueries.Update(context.Background(), log, db, existingUser, systemUser, differences, requestId); err != nil {
		log.Error().Err(err).Msg("Failed to update user in database")
		return nil, fmt.Errorf("failed to update user in database: %w", err)
	}

	log.Info().Interface("user", existingUser).Msg("User updated successfully")
	return existingUser, nil
}
