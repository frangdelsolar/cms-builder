package resourcemanager

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/queries"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

var DefaultListHandler ApiFunction = func(a *Resource, db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User

		isAllowed := a.Permissions.HasPermission(user.GetRoles(), OperationRead)
		if !isAllowed {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		queryParams, err := GetRequestQueryParams(r)
		if err != nil {
			SendJsonResponse(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		// Create slice to store the model instances.
		instances, err := a.GetSlice()
		if err != nil {
			log.Error().Err(err).Msgf("Error creating slice for model")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		pagination := &queries.Pagination{
			Total: 0,
			Page:  queryParams.Page,
			Limit: queryParams.Limit,
		}
		query := ""
		order := queryParams.Order
		isAdmin := user.HasRole(models.AdminRole)

		appName, err := a.GetName()
		if err != nil {
			log.Error().Err(err).Msgf("Error getting app name")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		msg := appName + "-list"

		if a.SkipUserBinding || isAdmin {
			err = queries.FindMany(db, instances, query, pagination, order).Error
			if err != nil {
				log.Error().Err(err).Msgf("Error finding instances")
				SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
				return
			}
		}

		query = "created_by_id = '" + user.StringID() + "'"
		err = queries.FindMany(db, instances, query, pagination, order).Error
		if err != nil {
			log.Error().Err(err).Msgf("Error finding instances")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		SendJsonResponseWithPagination(w, http.StatusOK, instances, msg, pagination)
	}
}
