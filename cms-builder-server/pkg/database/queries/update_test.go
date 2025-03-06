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

func TestUpdate_Success(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupDatabaseTestBed()
	db := testBed.Db
	user := testBed.AdminUser

	// Create a mock instance
	instance := testPkg.MockStruct{
		Field1: "Test Field 1",
		Field2: "Test Field 2",
	}
	db.DB.Create(&instance)

	// Update the instance
	instance.Field1 = "Updated Field 1"
	differences := map[string]interface{}{"field1": "Updated Field 1"}

	// Call the Update function
	err := dbQueries.Update(context.Background(), testBed.Logger, db, &instance, user, differences, "test-request-id")

	// Assertions
	assert.NoError(t, err)

	// Verify that the instance was updated
	var updatedInstance testPkg.MockStruct
	err = db.DB.First(&updatedInstance, instance.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Field 1", updatedInstance.Field1)

	// Verify that a history entry was created
	var historyEntry dbModels.DatabaseLog
	err = db.DB.Where("action = ? AND resource_id = ?", dbTypes.UpdateCRUDAction, instance.ID).First(&historyEntry).Error
	assert.NoError(t, err)
	assert.Equal(t, user.StringID(), historyEntry.UserId)
	assert.Equal(t, "test-request-id", historyEntry.TraceId)
}
