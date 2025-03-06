package resourcemanager

import (
	"net/http"

	"github.com/invopop/jsonschema"

	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

func DefaultSchemaHandler(app *rmTypes.Resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := svrUtils.ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		msg := "Schema for " + app.ResourceNames.Singular

		schema := jsonschema.Reflect(app.Model)
		svrUtils.SendJsonResponse(w, http.StatusOK, schema, msg)
	}
}
