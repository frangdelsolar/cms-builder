package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

func TestFindMany_Success(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupDatabaseTestBed()
	db := testBed.Db

	// Create multiple mock instances
	for i := 0; i < 15; i++ {
		instance := testPkg.MockStruct{
			Field1: "Test Field 1",
			Field2: "Test Field 2",
		}
		db.DB.Create(&instance)
	}

	// Call the FindMany function with pagination
	var instances []testPkg.MockStruct
	pagination := &dbTypes.Pagination{
		Page:  1,
		Limit: 10,
	}
	err := dbQueries.FindMany(context.Background(), testBed.Logger, db, &instances, pagination, "id desc", map[string]interface{}{}, []string{})

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 10, len(instances))
	assert.GreaterOrEqual(t, pagination.Total, int64(15))
}

func TestFindMany_NoPagination(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupDatabaseTestBed()
	db := testBed.Db

	// Create multiple mock instances
	for i := 0; i < 15; i++ {
		instance := testPkg.MockStruct{
			Field1: "Test Field 1",
			Field2: "Test Field 2",
		}
		db.DB.Create(&instance)
	}

	// Call the FindMany function without pagination
	var instances []testPkg.MockStruct
	err := dbQueries.FindMany(context.Background(), testBed.Logger, db, &instances, nil, "id desc", map[string]interface{}{}, []string{})

	// Assertions
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(instances), 15)
}

// TODO: TEST FILTERS PARAM
