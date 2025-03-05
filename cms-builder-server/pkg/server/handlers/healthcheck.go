package handlers

import (
	"net/http"

	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := svrUtils.ValidateRequestMethod(r, "GET"); err != nil {
		svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
		return
	}

	svrUtils.SendJsonResponse(w, http.StatusOK, nil, "OK")
}
