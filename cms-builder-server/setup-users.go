package orchestrator

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/google/uuid"
)

func (o *Orchestrator) SetupOrchestratorUsers() error {

	o.Logger.Info().Msg("Setting up orchestrator users")

	requestId := uuid.New().String()
	requestId = "automated::" + requestId

	o.Users = &OrchestratorUsers{
		Scheduler: &models.User{},
		God:       &models.User{},
		System:    &models.User{},
		Admin:     &models.User{},
	}

	systemUser, err := o.GetOrCreateSystemUser(requestId)
	if err != nil {
		o.Logger.Error().Err(err).Msg("Error getting or creating system user")
		return err
	}

	o.Users.System = systemUser

	usersData := []models.RegisterUserInput{
		{
			Name:             "God",
			Email:            "god@" + o.Config.GetString(EnvKeys.Domain),
			Password:         uuid.New().String(),
			Roles:            []models.Role{models.AdminRole},
			RegisterFirebase: false,
		},
		{
			Name:             o.Config.GetString(EnvKeys.AdminName),
			Email:            o.Config.GetString(EnvKeys.AdminEmail),
			Password:         o.Config.GetString(EnvKeys.AdminPassword),
			Roles:            []models.Role{models.AdminRole},
			RegisterFirebase: true,
		},
		{
			Name:             "Scheduler",
			Email:            "scheduler@" + o.Config.GetString(EnvKeys.Domain),
			Password:         uuid.New().String(),
			Roles:            []models.Role{models.SchedulerRole},
			RegisterFirebase: false,
		},
	}

	for _, userData := range usersData {
		user, err := server.CreateUserWithRole(userData, o.FirebaseClient, o.DB, o.Users.System, requestId, o.Logger)
		if err != nil {
			o.Logger.Error().Err(err).Interface("user", userData).Msg("Error creating user")
			return err
		}

		if user.Name == "Scheduler" {
			o.Users.Scheduler = user
		} else if user.Name == "God" {
			o.Users.God = user
		} else if user.Name == "Admin" {
			o.Users.Admin = user
		}
	}

	return nil
}

func (o *Orchestrator) GetOrCreateSystemUser(requestId string) (*models.User, error) {

	systemUser := models.User{
		Name:  "System",
		Email: "system@" + o.Config.GetString(EnvKeys.Domain),
		Roles: "admin",
	}

	err := queries.FindOne(o.DB, &systemUser, "email = ?", systemUser.Email).Error
	if err != nil {
		o.Logger.Warn().Err(err).Msg("System User not found")
	}

	if systemUser.ID == 0 || systemUser == (models.User{}) {

		o.Logger.Debug().Interface("user", systemUser).Msg("Creating system user from config")

		err := queries.Create(o.DB, &systemUser, &systemUser, requestId).Error
		if err != nil {
			o.Logger.Error().Err(err).Msg("Error creating system user")
			return nil, err
		}
	}

	return &systemUser, nil
}
