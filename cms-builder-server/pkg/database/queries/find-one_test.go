package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func TestFindOne_Success(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupDatabaseTestBed()
	db := testBed.Db

	// Create a mock instance
	instance := testPkg.MockStruct{
		Field1: "Test Field 1",
		Field2: "Test Field 2",
	}
	db.DB.Create(&instance)

	// Call the FindOne function
	var foundInstance testPkg.MockStruct
	filters := map[string]interface{}{
		"id": instance.ID,
	}

	err := dbQueries.FindOne(context.Background(), testBed.Logger, db, &foundInstance, filters, []string{})

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, instance.ID, foundInstance.ID)
	assert.Equal(t, instance.Field1, foundInstance.Field1)
	assert.Equal(t, instance.Field2, foundInstance.Field2)
}

func TestFindOne_ResourceNotFound(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupDatabaseTestBed()
	db := testBed.Db

	// Call the FindOne function for a non-existent resource
	var foundInstance testPkg.MockStruct
	filters := map[string]interface{}{
		"id": 99999,
	}
	err := dbQueries.FindOne(context.Background(), testBed.Logger, db, &foundInstance, filters, []string{})

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
