package testing

import (
	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	authTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	rmValidators "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/validators"
)

const (
	AllAllowedRole authTypes.Role = "all-allowed"
	JustReadRole   authTypes.Role = "just-read"
)

type MockStruct struct {
	authModels.SystemData
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func CreateMockResourceInstance(createdByID uint) *MockStruct {
	if createdByID == 0 {
		createdByID = RandomUint()
	}

	return &MockStruct{
		SystemData: authModels.SystemData{
			CreatedByID: createdByID,
			UpdatedByID: createdByID,
		},
		Field1: RandomString(10),
		Field2: RandomString(10),
	}
}

func SetupMockResource() *rmTypes.ResourceConfig {

	permissions := authTypes.RolePermissionMap{
		AllAllowedRole:            authConstants.AllAllowedAccess,
		authConstants.AdminRole:   authConstants.AllAllowedAccess,
		authConstants.VisitorRole: []authTypes.CrudOperation{authConstants.OperationRead},
	}

	validators := rmTypes.ValidatorsMap{
		"Field1": rmTypes.ValidatorsList{rmValidators.RequiredValidator},
	}

	config := &rmTypes.ResourceConfig{
		Model:           MockStruct{},
		SkipUserBinding: false,
		Validators:      validators,
		Permissions:     permissions,
	}

	return config
}
