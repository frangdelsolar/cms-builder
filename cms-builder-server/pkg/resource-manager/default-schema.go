package resourcemanager

import (
	"net/http"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	"github.com/invopop/jsonschema"
)

func DefaultSchemaHandler(app *Resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		appName, err := app.GetName()
		if err != nil {
			log.Error().Err(err).Msgf("Error getting app name")
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		msg := "Schema for " + appName

		schema := jsonschema.Reflect(app.Model)
		SendJsonResponse(w, http.StatusOK, schema, msg)
	}
}
