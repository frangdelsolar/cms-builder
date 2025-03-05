package database_test

// import (
// 	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
// )

// TestDelete_SingleEntity_Success tests the deletion of a single entity.
// func TestDelete_SingleEntity_Success(t *testing.T) {
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

// 	// Call the Delete function
// 	result := dbQueries.Delete(db, &instance, user, "test-request-id")

// 	// Assertions
// 	assert.NoError(t, result.Error)

// 	// Verify that the instance was deleted
// 	var deletedInstance MockStruct
// 	err := db.DB.First(&deletedInstance, instance.ID).Error
// 	assert.Error(t, err)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)

// 	// Verify that a history entry was created
// 	var historyEntry database.DatabaseLog
// 	err = db.DB.Where("action = ? AND resource_id = ?", database.dbTypes.DeleteCRUDAction, instance.ID).First(&historyEntry).Error
// 	assert.NoError(t, err)
// 	assert.Equal(t, user.StringID(), historyEntry.UserId)
// 	assert.Equal(t, "test-request-id", historyEntry.TraceId)
// }

// // TestDelete_Slice_Success tests the deletion of a slice of entities.
// func TestDelete_Slice_Success(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupDatabaseTestBed()
// 	db := testBed.Db
// 	user := testBed.AdminUser

// 	// Create mock instances
// 	instances := []MockStruct{
// 		{Field1: "Test Field 1", Field2: "Test Field 2"},
// 		{Field1: "Test Field 3", Field2: "Test Field 4"},
// 		{Field1: "Test Field 5", Field2: "Test Field 6"},
// 	}
// 	for i := range instances {
// 		dbQueries.Create(db, &instances[i], user, "test-request-id")
// 	}

// 	// Call the Delete function
// 	result := dbQueries.Delete(db, instances, user, "test-request-id")

// 	// Assertions
// 	assert.NoError(t, result.Error)

// 	// Verify that all instances were deleted
// 	for _, instance := range instances {
// 		var deletedInstance MockStruct
// 		err := db.DB.First(&deletedInstance, instance.ID).Error
// 		assert.Error(t, err)
// 		assert.Equal(t, gorm.ErrRecordNotFound, err)

// 		// Verify that a history entry was created for each instance
// 		var historyEntry database.DatabaseLog
// 		err = db.DB.Where("action = ? AND resource_id = ?", database.dbTypes.DeleteCRUDAction, instance.ID).First(&historyEntry).Error
// 		assert.NoError(t, err)
// 		assert.Equal(t, user.StringID(), historyEntry.UserId)
// 		assert.Equal(t, "test-request-id", historyEntry.TraceId)
// 	}
// }

// // TestDelete_SingleEntity_Error tests error handling when deleting a single entity fails.
// func TestDelete_SingleEntity_Error(t *testing.T) {
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

// 	// Simulate a database error by closing the database connection
// 	db.Close()

// 	// Call the Delete function
// 	result := dbQueries.Delete(db, &instance, user, "test-request-id")

// 	// Assertions
// 	assert.Error(t, result.Error)

// 	// Verify that the instance was not deleted
// 	var deletedInstance MockStruct
// 	err := db.DB.First(&deletedInstance, instance.ID).Error
// 	assert.NoError(t, err)
// 	assert.Equal(t, instance.ID, deletedInstance.ID)

// 	// Verify that no history entry was created
// 	var historyEntry database.DatabaseLog
// 	err = db.DB.Where("action = ? AND resource_id = ?", database.dbTypes.DeleteCRUDAction, instance.ID).First(&historyEntry).Error
// 	assert.Error(t, err)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)
// }

// // TestDelete_Slice_Error tests error handling when deleting a slice of entities fails.
// func TestDelete_Slice_Error(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupDatabaseTestBed()
// 	db := testBed.Db
// 	user := testBed.AdminUser

// 	// Create mock instances
// 	instances := []MockStruct{
// 		{Field1: "Test Field 1", Field2: "Test Field 2"},
// 		{Field1: "Test Field 3", Field2: "Test Field 4"},
// 		{Field1: "Test Field 5", Field2: "Test Field 6"},
// 	}
// 	for i := range instances {
// 		db.DB.Create(&instances[i])
// 	}

// 	// Simulate a database error by closing the database connection after the first deletion
// 	db.Close()

// 	// Call the Delete function
// 	result := dbQueries.Delete(db, instances, user, "test-request-id")

// 	// Assertions
// 	assert.Error(t, result.Error)

// 	// Verify that no instances were deleted
// 	for _, instance := range instances {
// 		var deletedInstance MockStruct
// 		err := db.DB.First(&deletedInstance, instance.ID).Error
// 		assert.NoError(t, err)
// 		assert.Equal(t, instance.ID, deletedInstance.ID)

// 		// Verify that no history entry was created
// 		var historyEntry database.DatabaseLog
// 		err = db.DB.Where("action = ? AND resource_id = ?", database.dbTypes.DeleteCRUDAction, instance.ID).First(&historyEntry).Error
// 		assert.Error(t, err)
// 		assert.Equal(t, gorm.ErrRecordNotFound, err)
// 	}
// }
