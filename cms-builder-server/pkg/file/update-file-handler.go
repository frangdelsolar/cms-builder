package file

import (
	"net/http"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	manager "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
)

func UpdateStoredFilesHandler(a *manager.Resource, db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := ValidateRequestMethod(r, http.MethodPut)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		SendJsonResponse(w, http.StatusMethodNotAllowed, nil, "You cannot update a file. You may delete and create a new one.")
	}
}
