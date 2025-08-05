package file

import (
	"fmt"
	"net/http"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	rmPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
	utilsPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

func DownloadStoredFileHandler(mgr *rmPkg.ResourceManager, db *dbTypes.DatabaseConnection, st storeTypes.Store) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "bytes")
			w.WriteHeader(http.StatusOK)
			return
		}

		requestCtx := svrUtils.GetRequestContext(r)
		log := requestCtx.Logger
		user := requestCtx.User
		isAdmin := user.HasRole(authConstants.AdminRole)

		err := svrUtils.ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		a, err := mgr.GetResource(fileModels.File{})
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		// 3. Check Permissions
		if !authUtils.UserIsAllowed(a.Permissions, user.GetRoles(), authConstants.OperationRead, a.ResourceNames.Singular, log) {
			svrUtils.SendJsonResponse(w, http.StatusForbidden, nil, "User is not allowed to access this resource")
			return
		}

		filters := map[string]interface{}{
			"id": svrUtils.GetUrlParam("id", r),
		}

		if !(a.SkipUserBinding || isAdmin) {
			filters["created_by_id"] = user.ID
		}

		instance := fileModels.File{}
		err = dbQueries.FindOne(r.Context(), log, db, &instance, filters, []string{})
		if err != nil {
			log.Error().Err(err).Msgf("Instance not found")
			svrUtils.SendJsonResponse(w, http.StatusNotFound, nil, "Instance not found")
			return
		}

		previousState := a.GetOne()
		_ = dbQueries.FindOne(r.Context(), log, db, &previousState, filters, []string{})

		// Update download count
		instance.DownloadCount++
		differences := utilsPkg.CompareInterfaces(previousState, instance)
		err = dbQueries.Update(r.Context(), log, db, &instance, user, differences, requestCtx.RequestId)
		if err != nil {
			log.Error().Err(err).Msg("Error updating instance")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error updating instance")
			return
		}

		// Open the file
		content, err := st.ReadFile(&instance, log)
		if err != nil {
			log.Error().Err(err).Msg("Error streaming file to response")
			svrUtils.SendJsonResponse(w, http.StatusInternalServerError, nil, "Error reading file")
			return
		}

		// Set headers for file download
		w.Header().Set("Content-Type", instance.MimeType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", instance.Name))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", instance.Size))
		w.Write(content)
	}
}
