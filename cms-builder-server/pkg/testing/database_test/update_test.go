package database_test

// import (
// 	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
// )

// func TestUpdate_Success(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupDatabaseTestBed()
// 	db := testBed.Db
// 	user := testBed.AdminUser

// 	// Create a mock instance
// 	instance := MockStruct{
// 		Field1: "Test Field 1",
// 		Field2: "Test Field 2",
// 	}
// 	db.DB.Create(&instance)

// 	// Update the instance
// 	instance.Field1 = "Updated Field 1"
// 	differences := map[string]interface{}{"field1": "Updated Field 1"}

// 	// Call the Update function
// 	result := dbQueries.Update(db, &instance, user, differences, "test-request-id")

// 	// Assertions
// 	assert.NoError(t, result.Error)

// 	// Verify that the instance was updated
// 	var updatedInstance MockStruct
// 	err := db.DB.First(&updatedInstance, instance.ID).Error
// 	assert.NoError(t, err)
// 	assert.Equal(t, "Updated Field 1", updatedInstance.Field1)

// 	// Verify that a history entry was created
// 	var historyEntry database.DatabaseLog
// 	err = db.DB.Where("action = ? AND resource_id = ?", database.UpdateCRUDAction, instance.ID).First(&historyEntry).Error
// 	assert.NoError(t, err)
// 	assert.Equal(t, user.StringID(), historyEntry.UserId)
// 	assert.Equal(t, "test-request-id", historyEntry.TraceId)
// }

// func TestUpdate_DatabaseError(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupDatabaseTestBed()
// 	db := testBed.Db
// 	user := testBed.AdminUser

// 	// Create a mock instance
// 	instance := MockStruct{
// 		Field1: "Test Field 1",
// 		Field2: "Test Field 2",
// 	}
// 	db.DB.Create(&instance)

// 	// Simulate a database error by closing the connection
// 	db.Close()

// 	// Update the instance
// 	instance.Field1 = "Updated Field 1"
// 	differences := map[string]interface{}{"field1": "Updated Field 1"}

// 	// Call the Update function
// 	result := dbQueries.Update(db, &instance, user, differences, "test-request-id")

// 	// Assertions
// 	assert.Error(t, result.Error)

// 	// Verify that the instance was not updated
// 	var notUpdatedInstance MockStruct
// 	err := db.DB.First(&notUpdatedInstance, instance.ID).Error
// 	assert.NoError(t, err)
// 	assert.Equal(t, "Test Field 1", notUpdatedInstance.Field1)
// }
