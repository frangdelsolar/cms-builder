package testing

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

const (
	AllAllowedRole models.Role = "all-allowed"
	JustReadRole   models.Role = "just-read"
)

type MockStruct struct {
	models.SystemData
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func CreateMockResourceInstance(createdByID uint) *MockStruct {
	if createdByID == 0 {
		createdByID = RandomUint()
	}

	return &MockStruct{
		SystemData: models.SystemData{
			CreatedByID: createdByID,
			UpdatedByID: createdByID,
		},
		Field1: RandomString(10),
		Field2: RandomString(10),
	}
}

func SetupMockResource() *mgr.ResourceConfig {

	permissions := server.RolePermissionMap{
		AllAllowedRole:     server.AllAllowedAccess,
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
