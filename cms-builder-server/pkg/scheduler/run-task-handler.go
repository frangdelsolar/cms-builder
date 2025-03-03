package scheduler

import (
	"errors"
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"gorm.io/gorm"
)

// DefaultCreateHandler handles the creation of a new resource.
// SchedulerJobDefinition
var RunSchedulerTaskHandler = func(manager *mgr.ResourceManager, db *database.Database, s JobRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		requestId := requestCtx.RequestId

		// 1. Validate Request Method
		err := ValidateRequestMethod(r, http.MethodPost)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, "Method not allowed")
			return
		}

		a, err := manager.GetResource(&SchedulerJobDefinition{})
		if err != nil {
			log.Error().Err(err).Msgf("Error getting resource")
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Error getting resource")
			return
		}

		// 2. Check Permissions
		if !UserIsAllowed(a.Permissions, user.GetRoles(), OperationCreate, a.ResourceNames.Singular, log) {
			SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to create this resource")
			return
		}

		// 4. Retrieve Instance ID and Fetch Instance
		jobDefinitionName := r.URL.Query().Get("job_definition_name")
		instance := a.GetOne()

		query := "name = ?"
		res := queries.FindOne(db, instance, query, jobDefinitionName, user.StringID())
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
				return
			}

			log.Error().Err(res.Error).Msgf("Error finding instance")
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Error finding resource")
			return
		}

		jd := instance.(*SchedulerJobDefinition)

		_, err = s.RunJob(jd, requestId, user, log, db)
		if err != nil {
			log.Error().Err(err).Msg("Error running task")
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Error running task")
			return
		}

		log.Debug().Interface("job_definition", jd).Msg("Job definition found in database")

		var task SchedulerTask
		err = db.DB.Where("job_definition_name = ?", jd.Name).Order("created_at DESC").First(&task).Error
		if err != nil {
			log.Error().Err(err).Msg("Error finding task")
			SendJsonResponse(w, http.StatusInternalServerError, nil, "Error finding task")
			return
		}
		// 9. Generate Success Message
		msg := jobDefinitionName + " has been triggered"

		// 10. Send Success Response
		SendJsonResponse(w, http.StatusCreated, task, msg)
	}
}
