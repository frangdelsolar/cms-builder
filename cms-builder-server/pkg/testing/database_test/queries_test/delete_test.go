package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	dbModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

// TestDelete_SingleEntity_Success tests the deletion of a single entity.
func TestDelete_SingleEntity_Success(t *testing.T) {
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

	// Call the Delete function
	err := dbQueries.Delete(context.Background(), testBed.Logger, db, &instance, user, "test-request-id")

	// Assertions
	assert.NoError(t, err)

	// Verify that the instance was deleted
	var deletedInstance testPkg.MockStruct
	err = db.DB.First(&deletedInstance, instance.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Verify that a history entry was created
	var historyEntry dbModels.DatabaseLog
	err = db.DB.Where("action = ? AND resource_id = ?", dbTypes.DeleteCRUDAction, instance.ID).First(&historyEntry).Error
	assert.NoError(t, err)
	assert.Equal(t, user.StringID(), historyEntry.UserId)
	assert.Equal(t, "test-request-id", historyEntry.TraceId)
}

// TestDelete_Slice_Success tests the deletion of a slice of entities.
func TestDelete_Slice_Success(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupDatabaseTestBed()
	db := testBed.Db
	user := testBed.AdminUser

	// Create mock instances
	instances := []testPkg.MockStruct{
		{Field1: "Test Field 1", Field2: "Test Field 2"},
		{Field1: "Test Field 3", Field2: "Test Field 4"},
		{Field1: "Test Field 5", Field2: "Test Field 6"},
	}
	for i := range instances {
		dbQueries.Create(context.Background(), testBed.Logger, db, &instances[i], user, "test-request-id")
	}

	// Call the Delete function
	err := dbQueries.Delete(context.Background(), testBed.Logger, db, instances, user, "test-request-id")

	// Assertions
	assert.NoError(t, err)

	// Verify that all instances were deleted
	for _, instance := range instances {
		var deletedInstance testPkg.MockStruct
		err := db.DB.First(&deletedInstance, instance.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		// Verify that a history entry was created for each instance
		var historyEntry dbModels.DatabaseLog
		err = db.DB.Where("action = ? AND resource_id = ?", dbTypes.DeleteCRUDAction, instance.ID).First(&historyEntry).Error
		assert.NoError(t, err)
		assert.Equal(t, user.StringID(), historyEntry.UserId)
		assert.Equal(t, "test-request-id", historyEntry.TraceId)
	}
}
