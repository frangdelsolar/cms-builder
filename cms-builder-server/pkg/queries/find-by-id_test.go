package queries_test

import (
	"fmt"
	"testing"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFindById(t *testing.T) {
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
	testRecord := TestModel{Name: "Alice"}
	result := db.Create(&testRecord)
	assert.NoError(t, result.Error)

	t.Run("Find existing record by ID", func(t *testing.T) {
		var foundRecord TestModel
		idStr := fmt.Sprintf("%d", testRecord.ID) // Convert uint to string
		result := FindById(testDB, idStr, &foundRecord, "")

		assert.NoError(t, result.Error, "Expected to find the record without error")
		assert.Equal(t, testRecord.ID, foundRecord.ID, "Expected to retrieve the correct record")
		assert.Equal(t, "Alice", foundRecord.Name, "Expected the correct name")
	})

	t.Run("Find with additional query condition", func(t *testing.T) {
		var foundRecord TestModel
		idStr := fmt.Sprintf("%d", testRecord.ID) // Convert uint to string
		result := FindById(testDB, idStr, &foundRecord, "name = 'Alice'")

		assert.NoError(t, result.Error, "Expected to find the record without error")
		assert.Equal(t, "Alice", foundRecord.Name, "Expected the correct name")
	})

	t.Run("Record not found", func(t *testing.T) {
		var foundRecord TestModel
		result := FindById(testDB, "999", &foundRecord, "")

		assert.Error(t, result.Error, "Expected an error when record does not exist")
		assert.Equal(t, gorm.ErrRecordNotFound, result.Error, "Expected gorm.ErrRecordNotFound")
	})
}
