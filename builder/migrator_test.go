package builder_test

import (
	"testing"

	"github.com/frangdelsolar/cms/builder"
	th "github.com/frangdelsolar/cms/builder/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	dbConfig := e.DB.Config

	t.Log("Migrating the database")
	err = builder.Migrate(dbConfig)
	assert.NoError(t, err, "Migrate should not return an error")
}
