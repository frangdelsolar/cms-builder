package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	utilsPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

func CreateUserWithRole(input authTypes.RegisterUserInput, firebase *cliPkg.FirebaseManager, db *dbTypes.DatabaseConnection, systemUser *authModels.User, requestId string, log *loggerTypes.Logger) (*authModels.User, error) {

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

	// Handle existing user (including soft-deleted ones)
	if existingUser != nil {
		if existingUser.DeletedAt.Valid {
			// Permanently delete the soft-deleted user first
			if err := dbQueries.HardDelete(context.Background(), log, db, existingUser, systemUser, requestId); err != nil {
				return nil, fmt.Errorf("failed to permanently delete previous user: %w", err)
			}
			// Then proceed to create new user
		} else {
			// Active user exists - handle normally
			return handleExistingUser(existingUser, fbUserId, db, systemUser, requestId, log)
		}
	}

	// Create the user in the database
	newUser := authModels.User{
		FirstName:  input.FirstName,
		LastName:   input.LastName,
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
func registerOrGetFirebaseUser(ctx context.Context, firebase *cliPkg.FirebaseManager, input authTypes.RegisterUserInput, log *loggerTypes.Logger) (string, error) {
	if !input.RegisterFirebase {
		return "", nil
	}

	log.Info().Str("email", input.Email).Msg("Registering user in Firebase")

	firebaseInput := cliPkg.FirebaseUserInput{
		Email:    input.Email,
		Name:     input.FirstName + " " + input.LastName,
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
func findUserByEmail(db *dbTypes.DatabaseConnection, email string, log *loggerTypes.Logger) (*authModels.User, error) {
	var user authModels.User

	// Use Unscoped() to include soft-deleted records in the query
	err := db.DB.Unscoped().Where("email = ?", email).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found at all
		}
		log.Error().Err(err).Msg("Failed to query user from database")
		return nil, fmt.Errorf("failed to query user from database: %w", err)
	}

	// Check if the user is soft-deleted
	if user.DeletedAt.Valid {
		log.Debug().Str("email", email).Time("deleted_at", user.DeletedAt.Time).
			Msg("Found soft-deleted user with this email")
		return &user, nil
	}

	// User exists and is not deleted
	return &user, nil
}

// Helper function to handle existing user (update Firebase ID if necessary)
func handleExistingUser(existingUser *authModels.User, fbUserId string, db *dbTypes.DatabaseConnection, systemUser *authModels.User, requestId string, log *loggerTypes.Logger) (*authModels.User, error) {
	if existingUser.FirebaseId == fbUserId {
		if fbUserId != "" {
			log.Info().Msg("User already exists in database with matching Firebase ID")
		}
		return existingUser, nil
	}

	previousState := *existingUser
	differences := utilsPkg.CompareInterfaces(previousState, existingUser)

	existingUser.FirebaseId = fbUserId

	if err := dbQueries.Update(context.Background(), log, db, existingUser, systemUser, differences, requestId); err != nil {
		log.Error().Err(err).Msg("Failed to update user in database")
		return nil, fmt.Errorf("failed to update user in database: %w", err)
	}

	log.Info().Interface("user", existingUser).Msg("User updated successfully")
	return existingUser, nil
}

func FormatRoles(roles []authTypes.Role) string {
	rolesStr := ""
	for _, role := range roles {
		rolesStr += role.S() + ","
	}

	if len(rolesStr) > 0 {
		rolesStr = rolesStr[:len(rolesStr)-1]
	}

	return rolesStr
}
