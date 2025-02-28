package database_test

import (
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestDelete_Success(t *testing.T) {
	// Setup test environment
	testBed := SetupDatabaseTestBed()
	db := testBed.Db
	user := testBed.AdminUser

	// Create a mock instance
	instance := MockStruct{
		Field1: "Test Field 1",
		Field2: "Test Field 2",
	}
	db.DB.Create(&instance)

	// Call the Delete function
	result := queries.Delete(db, &instance, user, "test-request-id")

	// Assertions
	assert.NoError(t, result.Error)

	// Verify that the instance was deleted
	var deletedInstance MockStruct
	err := db.DB.First(&deletedInstance, instance.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Verify that a history entry was created
	var historyEntry database.DatabaseLog
	err = db.DB.Where("action = ? AND resource_id = ?", database.DeleteCRUDAction, instance.ID).First(&historyEntry).Error
	assert.NoError(t, err)
	assert.Equal(t, user.StringID(), historyEntry.UserId)
	assert.Equal(t, "test-request-id", historyEntry.TraceId)
}

// func TestDelete_DatabaseError(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupServerTestBed()
// 	db := testBed.Db
// 	user := testBed.AdminUser

// 	// Create a mock instance
// 	instance := models.MockStruct{
// 		Field1: "Test Field 1",
// 		Field2: "Test Field 2",
// 	}
// 	db.DB.Create(&instance)

// 	// Simulate a database error by closing the connection
// 	db.Close()

// 	// Call the Delete function
// 	result := queries.Delete(db, &instance, user, "test-request-id")

// 	// Assertions
// 	assert.Error(t, result.Error)

// 	// Verify that the instance was not deleted
// 	var notDeletedInstance models.MockStruct
// 	err := db.DB.First(&notDeletedInstance, instance.ID).Error
// 	assert.NoError(t, err)
// 	assert.Equal(t, instance.ID, notDeletedInstance.ID)
// }

// func TestDelete_HistoryEntryFailure(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupServerTestBed()
// 	db := testBed.Db
// 	user := testBed.AdminUser

// 	// Create a mock instance
// 	instance := models.MockStruct{
// 		Field1: "Test Field 1",
// 		Field2: "Test Field 2",
// 	}
// 	db.DB.Create(&instance)

// 	// Simulate a failure in history entry creation by passing an invalid user
// 	invalidUser := &models.User{} // Invalid user with no ID

// 	// Call the Delete function
// 	result := queries.Delete(db, &instance, invalidUser, "test-request-id")

// 	// Assertions
// 	assert.NoError(t, result.Error) // The Delete function should still succeed

// 	// Verify that the instance was deleted
// 	var deletedInstance models.MockStruct
// 	err := db.DB.First(&deletedInstance, instance.ID).Error
// 	assert.Error(t, err)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)

// 	// Verify that no history entry was created
// 	var historyEntry models.DatabaseLog
// 	err = db.DB.Where("action = ? AND resource_id = ?", database.DeleteCRUDAction, instance.ID).First(&historyEntry).Error
// 	assert.Error(t, err)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)
// }
