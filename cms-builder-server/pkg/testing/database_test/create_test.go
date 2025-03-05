package database_test

import (
	"context"
	"testing"

	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/stretchr/testify/assert"
)

func TestCreate_Success(t *testing.T) {
	// Setup test environment
	testBed := SetupDatabaseTestBed()
	ctx := context.Background()
	log := testBed.Logger
	db := testBed.Db
	user := testBed.AdminUser

	// Create a mock instance
	instance := MockStruct{
		Field1: "Test Field 1",
		Field2: "Test Field 2",
	}

	// Call the Create function
	err := dbQueries.Create(ctx, log, db, &instance, user, "test-request-id")

	// Assertions
	assert.NoError(t, err)
	assert.NotZero(t, instance.ID) // Ensure the instance has an ID after creation

	// Verify that a history entry was created
	// var historyEntry database.DatabaseLog
	// err = db.DB.Where("action = ? AND resource_id = ?", database.dbTypes.CreateCRUDAction, instance.ID).First(&historyEntry).Error
	// assert.NoError(t, err)
	// assert.Equal(t, user.StringID(), historyEntry.UserId)
	// assert.Equal(t, "test-request-id", historyEntry.TraceId)
}

// func TestCreate_DatabaseError(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupDatabaseTestBed()
// 	db := testBed.Db
// 	user := testBed.AdminUser

// 	// Simulate a database error by closing the connection
// 	db.Close()

// 	// Create a mock instance
// 	instance := MockStruct{
// 		Field1: "Test Field 1",
// 		Field2: "Test Field 2",
// 	}

// 	// Call the Create function
// 	result := dbQueries.Create(db, &instance, user, "test-request-id")

// 	// Assertions
// 	assert.Error(t, result.Error)
// 	assert.Zero(t, instance.ID) // Ensure the instance does not have an ID
// }

// func TestCreate_HistoryEntryFailure(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupDatabaseTestBed()
// 	db := testBed.Db

// 	// Create a mock instance
// 	instance := MockStruct{
// 		Field1: "Test Field 1",
// 		Field2: "Test Field 2",
// 	}

// 	// Simulate a failure in history entry creation by passing an invalid user
// 	invalidUser := &models.User{} // Invalid user with no ID

// 	// Call the Create function
// 	result := dbQueries.Create(db, &instance, invalidUser, "test-request-id")

// 	// Assertions
// 	assert.NoError(t, result.Error) // The Create function should still succeed
// 	assert.NotZero(t, instance.ID)  // Ensure the instance has an ID after creation

// 	// Verify that no history entry was created
// 	var historyEntry database.DatabaseLog
// 	err := db.DB.Where("action = ? AND resource_id = ?", database.dbTypes.CreateCRUDAction, instance.ID).First(&historyEntry).Error
// 	assert.Error(t, err)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)
// }
