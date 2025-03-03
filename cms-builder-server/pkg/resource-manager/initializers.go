package resourcemanager

import (
	"fmt"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
	"github.com/rs/zerolog/log"
)

func InitializeResourceNames(r *Resource) (ResourceNames, error) {

	singular, err := utils.GetInterfaceName(r.Model)
	if err != nil {
		return ResourceNames{}, err
	}

	plural := utils.Pluralize(singular)
	kebab := utils.KebabCase(singular)
	kebabs := utils.KebabCase(plural)
	snake := utils.SnakeCase(singular)
	snakes := utils.SnakeCase(plural)

	return ResourceNames{
		Singular:      singular,
		Plural:        plural,
		SnakeSingular: snake,
		SnakePlural:   snakes,
		KebabSingular: kebab,
		KebabPlural:   kebabs,
	}, nil
}

func InitializeRoutes(r *Resource, input []server.Route, db *database.Database) error {
	baseRoute := "/api/" + r.ResourceNames.KebabPlural

	routes := []server.Route{
		{
			Path:         baseRoute,
			Handler:      r.Api.List(r, db),
			Name:         fmt.Sprintf("%s:list", r.ResourceNames.Singular),
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
		{
			Path:         baseRoute + "/schema",
			Handler:      r.Api.Schema(r),
			Name:         fmt.Sprintf("%s Schema", r.ResourceNames.Singular),
			RequiresAuth: false,
			Methods:      []string{http.MethodGet},
		},
		{
			Path:         baseRoute + "/new",
			Handler:      r.Api.Create(r, db),
			Name:         fmt.Sprintf("Create %s", r.ResourceNames.Singular),
			RequiresAuth: true,
			Methods:      []string{http.MethodPost},
		},
		{
			Path:         baseRoute + "/{id}/delete",
			Handler:      r.Api.Delete(r, db),
			Name:         fmt.Sprintf("Delete %s", r.ResourceNames.Singular),
			RequiresAuth: true,
			Methods:      []string{http.MethodDelete},
		},
		{
			Path:         baseRoute + "/{id}/update",
			Handler:      r.Api.Update(r, db),
			Name:         fmt.Sprintf("Update %s", r.ResourceNames.Singular),
			RequiresAuth: true,
			Methods:      []string{http.MethodPut},
		},
		{
			Path:         baseRoute + "/{id}",
			Handler:      r.Api.Detail(r, db),
			Name:         fmt.Sprintf("%s Detail", r.ResourceNames.Singular),
			RequiresAuth: true,
			Methods:      []string{http.MethodGet},
		},
	}

	for _, route := range routes {
		err := r.AddRoute(route)
		if err != nil {
			return err
		}
	}

	for _, route := range input {
		err := r.AddRoute(route)
		if err != nil {
			return err
		}
	}

	return nil
}

// InitializeValidators adds validators to the given resource.
// It loops through the given map of validators and adds each one to the resource's Validators map.
func InitializeValidators(r *Resource, input *ValidatorsMap) error {
	for fieldName, validators := range *input {
		for _, validator := range validators {
			err := r.AddValidator(fieldName, validator)
			if err != nil {
				log.Error().Err(err).Str("field", fieldName).Msg("Failed to add validator")
				return err
			}
		}
	}
	return nil
}

// InitializeHandlers returns a new ApiHandlers struct with default handlers.
// If the given input is not nil, it overwrites the default handlers with the given functions.
func InitializeHandlers(input *ApiHandlers) *ApiHandlers {
	handlers := &ApiHandlers{
		List:   DefaultListHandler,
		Detail: DefaultDetailHandler,
		Create: DefaultCreateHandler,
		Update: DefaultUpdateHandler,
		Delete: DefaultDeleteHandler,
		Schema: DefaultSchemaHandler,
	}

	if input != nil {
		if input.List != nil {
			handlers.List = input.List
		}

		if input.Detail != nil {
			handlers.Detail = input.Detail
		}

		if input.Create != nil {
			handlers.Create = input.Create
		}

		if input.Update != nil {
			handlers.Update = input.Update
		}

		if input.Delete != nil {
			handlers.Delete = input.Delete
		}

		if input.Schema != nil {
			handlers.Schema = input.Schema
		}
	}

	return handlers
}
