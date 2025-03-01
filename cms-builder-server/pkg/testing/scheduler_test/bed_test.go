package scheduler_test

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/scheduler"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func SetupSchedulerTestBed() TestUtils {
	db := NewTestDB()
	err := db.DB.AutoMigrate(models.User{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(database.DatabaseLog{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(SchedulerJobDefinition{})
	if err != nil {
		panic(err)
	}

	err = db.DB.AutoMigrate(SchedulerTask{})
	if err != nil {
		panic(err)
	}

	log := NewTestLogger()

	schedulerUser := CreateSchedulerUser()
	err = db.DB.Create(schedulerUser).Error
	if err != nil {
		panic(err)
	}

	scheduler, err := NewScheduler(db, schedulerUser, log)
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
