package scheduler

import (
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	schTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/types"
	"github.com/go-co-op/gocron/v2"
)

// NewScheduler initializes a new scheduler instance.
// Parameters:
//   - db: Database connection.
//   - schedulerUser: User associated with the scheduler.
//   - log: Logger instance.
//
// Returns:
//   - *Scheduler: Initialized scheduler instance.
//   - error: Error if initialization fails.
func NewScheduler(db *dbTypes.DatabaseConnection, schedulerUser *authModels.User, log *loggerTypes.Logger) (*Scheduler, error) {
	log.Info().Msg("Initializing scheduler")

	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	s.Start()
	return &Scheduler{
		Cron:        s,
		User:        schedulerUser,
		DB:          db,
		Logger:      log,
		TaskManager: schTypes.TaskManager{Tasks: map[string]string{}},
		JobRegistry: schTypes.JobRegistry{Jobs: map[string]schTypes.JobRegistryTaskDefinition{}},
	}, nil
}
