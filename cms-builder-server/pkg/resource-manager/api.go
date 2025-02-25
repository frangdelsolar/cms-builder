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

		requestCtx := GetRequestContext(r)
		log := requestCtx.Logger

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		output := make([]appInfo, 0)

		for _, rsc := range mgr.Resources {

			name, err := rsc.GetName()
			if err != nil {
				log.Error().Err(err).Msg("Error getting name")
				SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			}

			plural, _ := rsc.GetPluralName()
			snake, _ := rsc.GetSnakeCaseName()
			snakes, _ := rsc.GetSnakeCasePluralName()
			kebab, _ := rsc.GetKebabCaseName()
			kebabs, _ := rsc.GetKebabCasePluralName()

			url := apiBaseUrl + "/api/" + kebab + "/schema"

			data := appInfo{
				Name:        name,
				Plural:      plural,
				Snake:       snake,
				Kebab:       kebab,
				SnakePlural: snakes,
				KebabPlural: kebabs,
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
