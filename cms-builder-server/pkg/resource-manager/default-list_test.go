package resourcemanager_test

import (
	"testing"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
)

func TestDefaultListHandler(t *testing.T) {

	type MockResource struct {
		Field1 string
	}

	resourceConfig := &ResourceConfig{
		Model:           MockResource{},
		SkipUserBinding: false,
	}

	manager := NewResourceManager()

	resource, err := manager.AddResource(resourceConfig)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resource)

}
