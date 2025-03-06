package utils

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	firebaseAuth "firebase.google.com/go/auth"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

func VerifyUser(userIdToken string, firebase *cliPkg.FirebaseManager, db *dbTypes.DatabaseConnection, systemUser *authModels.User, requestId string, log *loggerTypes.Logger) (*authModels.User, error) {

	accessToken, err := firebase.VerifyIDToken(context.Background(), userIdToken)
	if err != nil {
		log.Error().Err(err).Msg("Error verifying token")
		return nil, err
	}

	localUser := authModels.User{}
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

func RegisterFirebaseUserInDatabase(accessToken *firebaseAuth.Token, firebase *cliPkg.FirebaseManager, db *dbTypes.DatabaseConnection, systemUser *authModels.User, requestId string, log *loggerTypes.Logger) (*authModels.User, error) {
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

	roles := string(authConstants.VisitorRole)

	// Create a local user object
	localUser := &authModels.User{
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
