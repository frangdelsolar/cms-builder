package database_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
)

// Re-run existing successful tests
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

// --- NEW TEST FOR THE IN CLAUSE FIX ---

func TestFindMany_INClauseFilter(t *testing.T) {
	// Setup test environment
	testBed := testPkg.SetupDatabaseTestBed()
	db := testBed.Db

	// 1. Create 5 mock instances and capture their IDs
	var createdIDs []uint
	for i := 0; i < 5; i++ {
		instance := testPkg.MockStruct{
			Field1: fmt.Sprintf("Filterable %d", i),
			Field2: "Test Field 2",
		}
		db.DB.Create(&instance)
		// Assuming MockStruct has an ID field that GORM populates (like 'ID uint')
		// We rely on the internal structure of MockStruct to get the ID.
		// For testing, we'll try to find the last inserted ID or query them back
		// since we don't have the struct definition. Let's rely on creation order.

		// This relies on GORM populating the ID field after creation.
		createdIDs = append(createdIDs, instance.ID)
	}

	// Sanity check: Ensure we captured 5 IDs
	assert.Equal(t, 5, len(createdIDs), "Should have created 5 records")

	// 2. Define the IDs we want to filter for (e.g., the 2nd, 3rd, and 4th created records)
	// We use the ID values, not the indices.
	idsToFilter := []uint{}
	if len(createdIDs) >= 4 {
		idsToFilter = append(idsToFilter, createdIDs[1], createdIDs[2], createdIDs[3]) // IDs 1, 2, 3 (index 1 to 3)
	} else {
		t.Skip("Skipping IN Clause test due to unexpected ID creation count.")
		return
	}

	// 3. Define the filter map: column name ("id") and slice of values ([]uint)
	filters := map[string]interface{}{
		"id": idsToFilter,
	}

	// 4. Call FindMany with the IN clause filter
	var instances []testPkg.MockStruct
	err := dbQueries.FindMany(context.Background(), testBed.Logger, db, &instances, nil, "id asc", filters, []string{})

	// 5. Assertions
	assert.NoError(t, err, "FindMany should not return an error with the IN clause filter")
	assert.Equal(t, 3, len(instances), "Should retrieve exactly 3 records matching the filtered IDs")

	// Optional: Assert that the retrieved IDs match the filtered IDs (e.g., check if the first retrieved ID is one of the desired IDs)
	if len(instances) > 0 {
		retrievedIDs := make([]uint, len(instances))
		for i, inst := range instances {
			retrievedIDs[i] = inst.ID // Assumes MockStruct has ID field
		}

		// Simple assertion to check if the set of retrieved IDs matches the set of filtered IDs
		assert.ElementsMatch(t, idsToFilter, retrievedIDs, "Retrieved IDs should match the filtered IDs")
	}
}
