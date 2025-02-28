package testutils

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	"github.com/joho/godotenv"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

type MockStruct struct {
	models.SystemData
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func init() {
	godotenv.Load()
}

func GetMockResourceManager() *mgr.ResourceManager {
	db := GetTestDB()

	return mgr.NewResourceManager(db, logger.Default)
}

func GetMockResource() *mgr.Resource {

	mockManager := GetMockResourceManager()

	resourceConfig := SetupMockResource()
	resource, err := mockManager.AddResource(resourceConfig)

	if err != nil {
		panic(err)
	}

	return resource

}

func SetupMockResource() *mgr.ResourceConfig {

	permissions := server.RolePermissionMap{
		models.AdminRole:   server.AllAllowedAccess,
		models.VisitorRole: []server.CrudOperation{server.OperationRead},
	}

	validators := mgr.ValidatorsMap{
		"Field1": mgr.ValidatorsList{mgr.RequiredValidator},
	}

	config := &mgr.ResourceConfig{
		Model:           MockStruct{},
		SkipUserBinding: false,
		Validators:      validators,
		Permissions:     permissions,
	}

	return config
}
