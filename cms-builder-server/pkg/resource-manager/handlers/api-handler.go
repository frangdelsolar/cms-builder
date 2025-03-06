package resourcemanager

import (
	"net/http"

	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

type Endpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}
type appInfo struct {
	Name        string              `json:"name"`
	Plural      string              `json:"pluralName"`
	Snake       string              `json:"snakeName"`
	Kebab       string              `json:"kebabName"`
	SnakePlural string              `json:"snakePluralName"`
	KebabPlural string              `json:"kebabPluralName"`
	Endpoints   map[string]Endpoint `json:"endpoints"`
}

func ApiHandler(resources map[string]*rmTypes.Resource, apiBaseUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := svrUtils.ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		output := make([]appInfo, 0)

		for _, rsc := range resources {

			url := apiBaseUrl + "/api/" + rsc.ResourceNames.KebabPlural + "/schema"

			data := appInfo{
				Name:        rsc.ResourceNames.Singular,
				Plural:      rsc.ResourceNames.Plural,
				Snake:       rsc.ResourceNames.SnakeSingular,
				Kebab:       rsc.ResourceNames.KebabSingular,
				SnakePlural: rsc.ResourceNames.SnakePlural,
				KebabPlural: rsc.ResourceNames.KebabPlural,
				Endpoints: map[string]Endpoint{
					"schema": {
						Method: http.MethodGet,
						Path:   url,
					},
				},
			}

			output = append(output, data)
		}

		svrUtils.SendJsonResponse(w, http.StatusOK, output, "api")
	}
}
