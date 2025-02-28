package database_test

import (
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestFindOne_Success(t *testing.T) {
	// Setup test environment
	testBed := SetupDatabaseTestBed()
	db := testBed.Db

	// Create a mock instance
	instance := MockStruct{
		Field1: "Test Field 1",
		Field2: "Test Field 2",
	}
	db.DB.Create(&instance)

	// Call the FindOne function
	var foundInstance MockStruct
	result := queries.FindOne(db, &foundInstance, "id = ?", instance.ID)

	// Assertions
	assert.NoError(t, result.Error)
	assert.Equal(t, instance.ID, foundInstance.ID)
	assert.Equal(t, instance.Field1, foundInstance.Field1)
	assert.Equal(t, instance.Field2, foundInstance.Field2)
}

func TestFindOne_ResourceNotFound(t *testing.T) {
	// Setup test environment
	testBed := SetupDatabaseTestBed()
	db := testBed.Db

	// Call the FindOne function for a non-existent resource
	var foundInstance MockStruct
	result := queries.FindOne(db, &foundInstance, "id = ?", 99999)

	// Assertions
	assert.Error(t, result.Error)
	assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
}

func TestFindOne_DatabaseError(t *testing.T) {
	// Setup test environment
	testBed := SetupDatabaseTestBed()
	db := testBed.Db

	// Simulate a database error by closing the connection
	db.Close()

	// Call the FindOne function
	var foundInstance MockStruct
	result := queries.FindOne(db, &foundInstance, "id = ?", 1)

	// Assertions
	assert.Error(t, result.Error)
}
