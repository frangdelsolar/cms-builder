package queries_test

import (
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFindMany(t *testing.T) {
	type TestModel struct {
		gorm.Model
		Name string
	}
	// Initialize in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// AutoMigrate the test model
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// Wrap in our database struct
	testDB := &database.Database{DB: db}

	// Insert test data
	testData := []TestModel{
		{Name: "Alice"},
		{Name: "Bob"},
		{Name: "Charlie"},
	}
	for _, record := range testData {
		db.Create(&record)
	}

	t.Run("Find all without pagination", func(t *testing.T) {
		var results []TestModel
		result := FindMany(testDB, &results, "", nil, "")

		assert.NoError(t, result.Error, "Expected no error in FindMany")
		assert.Equal(t, len(testData), len(results), "Expected all records to be returned")
	})

	t.Run("Find with pagination", func(t *testing.T) {
		var results []TestModel
		pagination := &Pagination{Page: 1, Limit: 2}
		result := FindMany(testDB, &results, "", pagination, "")

		assert.NoError(t, result.Error, "Expected no error in FindMany")
		assert.Equal(t, 2, len(results), "Expected only 2 records due to pagination")
		assert.Equal(t, "Charlie", results[0].Name, "First result should be Charlie")
		assert.Equal(t, "Bob", results[1].Name, "Second result should be Bob")
		assert.Equal(t, int64(len(testData)), pagination.Total, "Pagination total should match total records")
	})

	t.Run("Find with ordering", func(t *testing.T) {
		var results []TestModel
		result := FindMany(testDB, &results, "", nil, "name asc")

		assert.NoError(t, result.Error, "Expected no error in FindMany")
		assert.Equal(t, "Alice", results[0].Name, "First record should be Alice")
		assert.Equal(t, "Charlie", results[len(results)-1].Name, "Last record should be Charlie")
	})

	t.Run("Find with query filter", func(t *testing.T) {
		var results []TestModel
		result := FindMany(testDB, &results, "name = 'Bob'", nil, "")

		assert.NoError(t, result.Error, "Expected no error in FindMany")
		assert.Equal(t, 1, len(results), "Expected one result for name = 'Bob'")
		assert.Equal(t, "Bob", results[0].Name, "Expected the name to be Bob")
	})
}
