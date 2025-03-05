package scheduler

import (
	"net/http"

	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

// DefaultCreateHandler handles the creation of a new resource.
// SchedulerJobDefinition
var RunSchedulerTaskHandler = func(manager *mgr.ResourceManager, db *dbTypes.DatabaseConnection, s JobRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := svrUtils.GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		requestId := requestCtx.RequestId

		// 1. Validate Request Method
		err := svrUtils.ValidateRequestMethod(r, http.MethodPost)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, "Method not allowed")
			return
		}

		a, err := manager.GetResource(&SchedulerJobDefinition{})
		if err != nil {
			log.Error().Err(err).Msgf("Error getting resource")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error getting resource")
			return
		}

		// 2. Check Permissions
		if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationCreate, a.ResourceNames.Singular, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		// 4. Retrieve Instance ID and Fetch Instance
		jobDefinitionName := r.URL.Query().Get("job_definition_name")
		instance := a.GetOne()

		filters := map[string]interface{}{
			"name": jobDefinitionName,
		}

		err = dbQueries.FindOne(r.Context(), log, db, instance, filters)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		jd := instance.(*SchedulerJobDefinition)

		_, err = s.RunJob(jd, requestId, user, log, db)
		if err != nil {
			log.Error().Err(err).Msg("Error running task")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error running task")
			return
		}

		var task SchedulerTask
		err = db.DB.Where("job_definition_name = ?", jd.Name).Order("created_at DESC").First(&task).Error
		if err != nil {
			log.Error().Err(err).Msg("Error finding task")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error finding task")
			return
		}
		// 9. Generate Success Message
		msg := jobDefinitionName + " has been triggered"

		// 10. Send Success Response
		svrUtils.SendJsonResponse(w, http.StatusCreated, task, msg)
	}
}
