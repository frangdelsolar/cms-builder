package resourcemanager

import (
	"net/http"

	. "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
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

func ApiHandler(mgr *ResourceManager, apiBaseUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		output := make([]appInfo, 0)

		for _, rsc := range mgr.Resources {

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

		SendJsonResponse(w, http.StatusOK, output, "api")
	}
}
