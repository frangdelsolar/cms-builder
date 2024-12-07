package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TODO: Needs to figure out how to generate for those urls that are not comming from a model

const (
	PostmanSchemaFilePath = "postman/collection.json"
	PostmanEnvFilePath    = "postman/environment.json"

	PostmanSchemaVersion = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
	PostmanVariableScope = "environment"

	keyFirebaseApiKey  = "FIREBASE_API_KEY"
	keyFirebaseIdToken = "FIREBASE_ID_TOKEN"
	keyBaseUrl         = "URL"
	keyEmail           = "EMAIL"
	keyPassword        = "PASSWORD"
)

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

func (b *Builder) ExportPostman() error {
	schema, err := b.GetPostmanCollection()
	if err != nil {
		// return err
	}
	env, err := b.GetPostmanEnv()
	if err != nil {
		return err
	}

	// write the schema to a file
	err = WriteFile(PostmanSchemaFilePath, schema)
	if err != nil {
		return err
	}

	// write the env to a file
	err = WriteFile(PostmanEnvFilePath, env)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) GetPostmanCollection() (*PostmanCollection, error) {
	appName := config.GetString(EnvKeys.AppName)
	if appName == "" {
		appName = "CollectionName"
	}

	collection := PostmanCollection{
		Info: PostmanCollectionInfo{
			Name:   strings.ToUpper(appName),
			Schema: PostmanSchemaVersion,
		},
		Auth: PostmanRequestAuth{
			Type: "bearer",
			Bearer: []PostmanBearer{
				{
					Key:   "token",
					Value: "{{" + keyFirebaseIdToken + "}}",
					Type:  "string",
				},
			},
		},
		Item:  make([]PostmanCollectionItem, 0),
		Event: make([]PostmanCollectionEvent, 0),
	}

	// Auth
	collection.Item = append(collection.Item, PostmanCollectionItem{
		Name: "Auth",
		Item: []PostmanCollectionItemItem{
			{
				Name: "Register",
				Request: PostmanCollectionItemItemRequest{
					Auth: PostmanRequestAuth{
						Type: "noauth",
					},
					Method: "POST",
					Header: []PostmanHeader{},
					Body: PostmanRequestBody{
						Mode: "raw",
						Raw:  "{\n  \"name\": \"admin\",\n  \"email\": \"{{" + keyEmail + "}}\",\n    \"password\": \"{{" + keyPassword + "}}\"\n}",
						Options: PostmanRequestOptions{
							Raw: PostmanRequestOptionsRaw{
								Language: "json",
							},
						},
					},
					URL: PostmanRequestURL{
						Raw:   "{{URL}}/auth/register",
						Host:  []string{"{{URL}}"},
						Path:  []string{"auth", "register"},
						Query: []PostmanQuery{},
					},
				},
				Response: []interface{}{},
			},
			{
				Name: "Login",
				Event: []PostmanCollectionEvent{
					{
						Listen: "test",
						Script: PostmanCollectionEventScript{
							Type: "text/javascript",
							Exec: []string{
								"pm.environment.set(\"" + keyFirebaseIdToken + "\", pm.response.json().idToken);",
							},
						},
					},
				},
				Request: PostmanCollectionItemItemRequest{
					Auth: PostmanRequestAuth{
						Type: "noauth",
					},
					Method: "POST",
					Header: []PostmanHeader{},
					Body: PostmanRequestBody{
						Mode: "raw",
						Raw:  "{\n    \"email\":\"{{" + keyEmail + "}}\",\n    \"password\":\"{{" + keyPassword + "}}\",\n    \"returnSecureToken\":true\n}\n",
						Options: PostmanRequestOptions{
							Raw: PostmanRequestOptionsRaw{
								Language: "json",
							},
						},
					},
					URL: PostmanRequestURL{
						Raw:      "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword",
						Protocol: "https",
						Host:     []string{"www", "googleapis", "com"},
						Path:     []string{"identitytoolkit", "v3", "relyingparty", "verifyPassword"},
						Query: []PostmanQuery{
							{
								Key:   "key",
								Value: "{{" + keyFirebaseApiKey + "}}",
							},
						},
					},
				},
				Response: []interface{}{},
			},
		},
	})

	// Files
	collection.Item = append(collection.Item, PostmanCollectionItem{
		Name: "File",
		Item: []PostmanCollectionItemItem{
			{
				Name: "Upload File",
				Request: PostmanCollectionItemItemRequest{
					Method: "POST",
					Header: []PostmanHeader{},
					Body: PostmanRequestBody{
						Mode: "formdata",
						FormData: []PostmanFormDataItem{
							{
								Key:   "file",
								Type:  "file",
								Value: "<FILE_NAME>",
							},
						},
					},
					URL: PostmanRequestURL{
						Raw:   "{{URL}}/private/files/upload",
						Host:  []string{"{{URL}}"},
						Path:  []string{"private", "files", "upload"},
						Query: []PostmanQuery{},
					},
				},
				Response: []interface{}{},
			},
			{
				Name: "Download File",
				Request: PostmanCollectionItemItemRequest{
					Method: "GET",
					Header: []PostmanHeader{},
					URL: PostmanRequestURL{
						Raw:   "{{URL}}/private/files/<FILE_NAME>",
						Host:  []string{"{{URL}}"},
						Path:  []string{"private", "files", "<FILE_NAME>"},
						Query: []PostmanQuery{},
					},
				},
				Response: []interface{}{},
			},
			{
				Name: "Delete File",
				Request: PostmanCollectionItemItemRequest{
					Method: "DELETE",
					Header: []PostmanHeader{},
					URL: PostmanRequestURL{
						Raw:   "{{URL}}/private/files/<FILE_NAME>",
						Host:  []string{"{{URL}}"},
						Path:  []string{"private", "files", "<FILE_NAME>"},
						Query: []PostmanQuery{},
					},
				},
				Response: []interface{}{},
			},
		},
	})

	// Iterate over apps to build the collection
	for _, app := range b.Admin.apps {

		path := GetAppPath(&app)
		body := GetBody(&app)
		appId := app.Name() + "Id"
		appIdExpr := "{{" + appId + "}}"

		collection.Item = append(collection.Item, PostmanCollectionItem{
			Name: app.Name(),
			Item: []PostmanCollectionItemItem{
				{
					Name: "Create " + app.Name(),
					Event: []PostmanCollectionEvent{
						{
							Listen: "test",
							Script: PostmanCollectionEventScript{
								Type: "text/javascript",
								Exec: []string{
									"pm.environment.set(\"" + appId + "\", pm.response.json().data.ID);",
								},
							},
						},
					},
					Request: PostmanCollectionItemItemRequest{
						Method: "POST",
						Header: []PostmanHeader{},
						Body: PostmanRequestBody{
							Mode: "raw",
							Raw:  body,
							Options: PostmanRequestOptions{
								Raw: PostmanRequestOptionsRaw{
									Language: "json",
								},
							},
						},
						URL: PostmanRequestURL{
							Raw: strings.Join(path, "/") + "/new",
							Host: []string{
								"{{" + keyBaseUrl + "}}",
							},
							Path:  append(path[1:], "new"),
							Query: make([]PostmanQuery, 0),
						},
					},
				},
				{
					Name: "List " + app.Name(),
					Request: PostmanCollectionItemItemRequest{
						Method: "GET",
						Header: []PostmanHeader{},
						URL: PostmanRequestURL{
							Raw: strings.Join(path, "/"),
							Host: []string{
								"{{" + keyBaseUrl + "}}",
							},
							Path: path[1:],
							Query: []PostmanQuery{
								{
									Key:   "page",
									Value: "1",
								},
								{
									Key:   "limit",
									Value: "10",
								},
							},
						},
					},
				},
				{
					Name: "Update " + app.Name(),
					Request: PostmanCollectionItemItemRequest{
						Method: "PUT",
						Header: []PostmanHeader{},
						Body: PostmanRequestBody{
							Mode: "raw",
							Raw:  body,
							Options: PostmanRequestOptions{
								Raw: PostmanRequestOptionsRaw{
									Language: "json",
								},
							},
						},
						URL: PostmanRequestURL{
							Raw: strings.Join(path, "/") + "/" + appIdExpr + "/update",
							Host: []string{
								"{{" + keyBaseUrl + "}}",
							},
							Path: append(path[1:], []string{
								appIdExpr,
								"update",
							}...),
						},
					},
				},
				{
					Name: "Detail " + app.Name(),
					Request: PostmanCollectionItemItemRequest{
						Method: "GET",
						Header: []PostmanHeader{},
						URL: PostmanRequestURL{
							Raw: strings.Join(path, "/") + "/" + appIdExpr,
							Host: []string{
								"{{" + keyBaseUrl + "}}",
							},
							Path: append(path[1:], []string{
								appIdExpr,
							}...),
						},
					},
				},
				{
					Name: "Delete " + app.Name(),
					Request: PostmanCollectionItemItemRequest{
						Method: "DELETE",
						Header: []PostmanHeader{},
						URL: PostmanRequestURL{
							Raw: strings.Join(path, "/") + "/" + appIdExpr + "/delete",
							Host: []string{
								"{{" + keyBaseUrl + "}}",
							},
							Path: append(path[1:], []string{
								appIdExpr,
								"delete",
							}...),
						},
					},
				},
			},
		})
	}

	return &collection, nil
}

// GetPostmanEnv returns a PostmanEnv struct with the variables from the
// corresponding .env file. If the file does not exist or there is an error
// reading it, an error is returned.
func (b *Builder) GetPostmanEnv() (*PostmanEnv, error) {
	appName := config.GetString(EnvKeys.AppName)
	if appName == "" {
		appName = "CollectionName"
	}

	environment := config.GetString(EnvKeys.Environment)
	if environment == "" {
		return nil, fmt.Errorf("unable to get environment")
	}
	envSchema := PostmanEnv{
		PostmanVariableScope: PostmanVariableScope,
		Values:               make([]PostmanEnvValue, 0),
	}
	envSchema.Name = strings.ToUpper(appName) + "_" + strings.ToUpper(environment)

	// Url
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyBaseUrl,
		Value:   config.GetString(EnvKeys.BaseUrl),
		Enabled: true,
	})

	// Email
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyEmail,
		Value:   "admin@admin.com",
		Enabled: true,
	})

	// Password
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyPassword,
		Value:   "admin123",
		Enabled: true,
	})

	// FirebaseAPIkey
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyFirebaseApiKey,
		Value:   config.GetString(EnvKeys.FirebaseApiKey),
		Enabled: true,
	})

	// FirebaseIdToken
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyFirebaseIdToken,
		Value:   "",
		Enabled: true,
	})

	return &envSchema, nil
}

// WriteFile writes data to a file.
//
// It takes two arguments:
//
//   - filePath: The path to the file to write to.
//   - data: The data to write to the file. It must be a type that can be encoded
//     to JSON.
//
// The function creates the file if it does not exist, and overwrites it if it does.
// It also sets the indentation of the JSON encoder to 4 spaces.
//
// The function returns an error if it is unable to write to the file.
func WriteFile(filePath string, data interface{}) error {

	// Mkdir
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	log.Info().Str("filePath", filePath).Msg("Wrote file")

	return nil
}

func GetAppPath(app *App) []string {
	path := []string{"{{" + keyBaseUrl + "}}"}
	if !app.skipUserBinding {
		path = append(path, "private")
	}
	path = append(path, []string{
		"api",
		app.PluralName(),
	}...)

	return path
}

func GetBody(app *App) string {

	data, err := JsonifyInterface(app.model)
	if err != nil {
		return ""
	}

	jsonData, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		return ""
	}

	return string(jsonData)
}
