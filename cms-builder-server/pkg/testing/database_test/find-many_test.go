package database_test

// import (
// 	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
// )

// func TestFindMany_Success(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupDatabaseTestBed()
// 	db := testBed.Db

// 	// Create multiple mock instances
// 	for i := 0; i < 15; i++ {
// 		instance := MockStruct{
// 			Field1: "Test Field 1",
// 			Field2: "Test Field 2",
// 		}
// 		db.DB.Create(&instance)
// 	}

// 	// Call the FindMany function with pagination
// 	var instances []MockStruct
// 	pagination := &queries.Pagination{
// 		Page:  1,
// 		Limit: 10,
// 	}
// 	result := queries.FindMany(db, &instances, pagination, "id desc", "")

// 	// Assertions
// 	assert.NoError(t, result.Error)
// 	assert.Equal(t, 10, len(instances))
// 	assert.GreaterOrEqual(t, pagination.Total, int64(15))
// }

// func TestFindMany_NoPagination(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupDatabaseTestBed()
// 	db := testBed.Db

// 	// Create multiple mock instances
// 	for i := 0; i < 15; i++ {
// 		instance := MockStruct{
// 			Field1: "Test Field 1",
// 			Field2: "Test Field 2",
// 		}
// 		db.DB.Create(&instance)
// 	}

// 	// Call the FindMany function without pagination
// 	var instances []MockStruct
// 	result := queries.FindMany(db, &instances, nil, "id desc", "")

// 	// Assertions
// 	assert.NoError(t, result.Error)
// 	assert.GreaterOrEqual(t, len(instances), 15)
// }

// func TestFindMany_DatabaseError(t *testing.T) {
// 	// Setup test environment
// 	testBed := SetupDatabaseTestBed()
// 	db := testBed.Db

// 	// Simulate a database error by closing the connection
// 	db.Close()

// 	// Call the FindMany function
// 	var instances []MockStruct
// 	pagination := &queries.Pagination{
// 		Page:  1,
// 		Limit: 10,
// 	}
// 	result := queries.FindMany(db, &instances, pagination, "id desc", "")

// 	// Assertions
// 	assert.Error(t, result.Error)
// }
