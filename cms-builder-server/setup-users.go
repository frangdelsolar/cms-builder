package orchestrator

import (
	"context"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/google/uuid"
)

func (o *Orchestrator) SetupOrchestratorUsers() error {

	o.Logger.Info().Msg("Setting up orchestrator users")

	requestId := uuid.New().String()
	requestId = "automated::" + requestId

	o.Users = &OrchestratorUsers{
		Scheduler: &authModels.User{},
		God:       &authModels.User{},
		System:    &authModels.User{},
		Admin:     &authModels.User{},
	}

	systemUser, err := o.GetOrCreateSystemUser(requestId)
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error getting or creating system user")
		return err
	}

	o.Users.System = systemUser

	usersData := []authTypes.RegisterUserInput{
		{
			FirstName:        "God",
			Email:            "god@" + o.Config.GetString(EnvKeys.Domain),
			Password:         uuid.New().String(),
			Roles:            []authTypes.Role{authConstants.AdminRole},
			RegisterFirebase: false,
		},
		{
			FirstName:        o.Config.GetString(EnvKeys.AdminName),
			Email:            o.Config.GetString(EnvKeys.AdminEmail),
			Password:         o.Config.GetString(EnvKeys.AdminPassword),
			Roles:            []authTypes.Role{authConstants.AdminRole},
			RegisterFirebase: true,
		},
		{
			FirstName:        "Scheduler",
			Email:            "scheduler@" + o.Config.GetString(EnvKeys.Domain),
			Password:         uuid.New().String(),
			Roles:            []authTypes.Role{authConstants.SchedulerRole},
			RegisterFirebase: false,
		},
	}

	for _, userData := range usersData {
		user, err := authUtils.CreateUserWithRole(userData, o.FirebaseClient, o.DB, o.Users.System, requestId, o.Logger)
		if err != nil {
			o.Logger.Error().Err(err).Interface("user", userData).Msg("Error creating user")
			return err
		}

		if user.FirstName == "Scheduler" {
			o.Users.Scheduler = user
		} else if user.FirstName == "God" {
			o.Users.God = user
		} else if user.FirstName == "Admin" {
			o.Users.Admin = user
		}
	}

	return nil
}

func (o *Orchestrator) GetOrCreateSystemUser(requestId string) (*authModels.User, error) {

	systemUser := authModels.User{
		FirstName: "System",
		Email:     "system@" + o.Config.GetString(EnvKeys.Domain),
		Roles:     "admin",
	}

	filters := map[string]interface{}{
		"email": systemUser.Email,
	}

	ctx := context.Background()

	err := dbQueries.FindOne(ctx, o.Logger, o.DB, &systemUser, filters, []string{})
	if err != nil {
		o.Logger.Warn().Err(err).Msg("System User not found")
	}

	if systemUser.ID == 0 || systemUser == (authModels.User{}) {
		o.Logger.Debug().Interface("user", systemUser).Msg("Creating system user from config")
		err := dbQueries.Create(ctx, o.Logger, o.DB, &systemUser, &systemUser, requestId)
		if err != nil {
			o.Logger.Error().Err(err).Msg("Error creating system user")
			return nil, err
		}
	}

	return &systemUser, nil
}
