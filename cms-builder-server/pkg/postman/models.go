package postman

import "time"

type PostmanEnvValue struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Enabled bool   `json:"enabled"`
}

type PostmanEnv struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Values               []PostmanEnvValue `json:"values"`
	PostmanVariableScope string            `json:"_postman_variable_scope"`
	PostmanExportedAt    time.Time         `json:"_postman_exported_at"`
	PostmanExportedUsing string            `json:"_postman_exported_using"`
}

type PostmanCollectionInfo struct {
	PostmanID  string `json:"_postman_id"`
	Name       string `json:"name"`
	Schema     string `json:"schema"`
	ExporterID string `json:"_exporter_id"`
}

// PostmanHeader struct for request headers
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"` // Optional type for extensibility
}

// Raw struct for raw data
type PostmanRequestOptionsRaw struct {
	Language string `json:"language"`
}
type PostmanRequestOptions struct {
	Raw PostmanRequestOptionsRaw `json:"options"`
}

type PostmanFormDataItem struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Src   string `json:"src"`
	Value string `json:"value"`
}

// PostmanRequestBody struct for request body content
type PostmanRequestBody struct {
	Mode     string                `json:"mode"`
	Raw      string                `json:"raw"`
	Options  PostmanRequestOptions `json:"options"`
	FormData []PostmanFormDataItem `json:"formdata"`
}

type PostmanBearer struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

// PostmanRequestAuth struct for request authentication details
type PostmanRequestAuth struct {
	Type   string          `json:"type"`
	Bearer []PostmanBearer `json:"bearer"`
}

type PostmanQuery struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled"`
}

// PostmanRequestURL struct for request URL components
type PostmanRequestURL struct {
	Raw      string         `json:"raw"`
	Host     []string       `json:"host"`
	Path     []string       `json:"path"`
	Query    []PostmanQuery `json:"query"`
	Protocol string         `json:"protocol"`
}

// PostmanCollectionItemItemRequest struct for individual request details
type PostmanCollectionItemItemRequest struct {
	Auth   PostmanRequestAuth `json:"auth"`
	Method string             `json:"method"`
	Header []PostmanHeader    `json:"header"`
	Body   PostmanRequestBody `json:"body"`
	URL    PostmanRequestURL  `json:"url"`
}

type PostmanCollectionEventScript struct {
	Exec     []string `json:"exec"`
	Type     string   `json:"type"`
	Packages struct {
	} `json:"packages"`
}

type PostmanCollectionEvent struct {
	Listen string                       `json:"listen"`
	Script PostmanCollectionEventScript `json:"script"`
}

type PostmanCollectionItemItem struct {
	Name     string                           `json:"name"`
	Request  PostmanCollectionItemItemRequest `json:"request"`
	Response []interface{}                    `json:"response"`
	Event    []PostmanCollectionEvent         `json:"event,omitempty"`
}

type PostmanCollectionItem struct {
	Name string                      `json:"name"`
	Item []PostmanCollectionItemItem `json:"item"`
}

type PostmanCollection struct {
	Info  PostmanCollectionInfo    `json:"info"`
	Item  []PostmanCollectionItem  `json:"item"`
	Auth  PostmanRequestAuth       `json:"auth"`
	Event []PostmanCollectionEvent `json:"event"`
}
