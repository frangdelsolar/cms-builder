package testing

import (
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	schPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	schModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler/models"
)

func SetupSchedulerTestBed() TestUtils {
	db := NewTestDB()
	err := db.DB.AutoMigrate(authModels.User{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(dbModels.DatabaseLog{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(schModels.SchedulerJobDefinition{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(schModels.SchedulerTask{})
	if err != nil {
		panic(err)
	}

	log := NewTestLogger()

	schedulerUser := CreateSchedulerUser()
	err = db.DB.Create(schedulerUser).Error
	if err != nil {
		panic(err)
	}

	scheduler, err := schPkg.NewScheduler(db, schedulerUser, log)
	if err != nil {
		panic(err)
	}

	return TestUtils{
		Scheduler:     scheduler,
		SchedulerUser: schedulerUser,
		Db:            db,
		Logger:        log,
	}
}
