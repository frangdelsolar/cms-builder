package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func TestCreate_Success(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupDatabaseTestBed()
	ctx := context.Background()
	log := testBed.Logger
	db := testBed.Db
	user := testBed.AdminUser

	// Create a mock instance
	instance := testPkg.MockStruct{
		Field1: "Test Field 1",
		Field2: "Test Field 2",
	}

	// Call the Create function
	err := dbQueries.Create(ctx, log, db, &instance, user, "test-request-id")

	// Assertions
	assert.NoError(t, err)
	assert.NotZero(t, instance.ID) // Ensure the instance has an ID after creation

	// Verify that a history entry was created
	var historyEntry dbModels.DatabaseLog
	err = db.DB.Where("action = ? AND resource_id = ?", dbTypes.CreateCRUDAction, instance.ID).First(&historyEntry).Error
	assert.NoError(t, err)
	assert.Equal(t, user.StringID(), historyEntry.UserId)
	assert.Equal(t, "test-request-id", historyEntry.TraceId)
}
