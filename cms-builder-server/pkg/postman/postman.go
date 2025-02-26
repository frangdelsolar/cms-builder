package postman

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
)

// TODO: Needs to figure out how to generate for those urls that are not comming from a model

const (
	PostmanSchemaFilePath = "postman-output/collection.json"
	PostmanEnvFilePath    = "postman-output/environment.json"

	PostmanSchemaVersion = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
	PostmanVariableScope = "environment"

	keyFirebaseApiKey  = "FIREBASE_API_KEY"
	keyFirebaseIdToken = "FIREBASE_ID_TOKEN"
	keyBaseUrl         = "URL"
	keyAdminEmail      = "ADMIN_EMAIL"
	keyAdminPassword   = "ADMIN_PASSWORD"
	keyOrigin          = "ORIGIN"
)

var log = logger.Default

func Placeholder(key string) string {
	return "{{" + key + "}}"
}

func ExportPostman(
	appName,
	environment,
	baseUrl,
	adminEmail,
	adminPassword,
	firebaseApiKey string,
	resources []mgr.Resource,
) error {

	log.Debug().Interface("resources", len(resources)).Msg("Exporting postman...")

	schemaCollection, err := GetPostmanCollection(appName, resources)
	if err != nil {
		return err
	}
	envCollection, err := GetPostmanEnv(
		appName,
		environment,
		baseUrl,
		adminEmail,
		adminPassword,
		firebaseApiKey,
	)
	if err != nil {
		return err
	}

	// write the schema to a file
	err = WriteFile(PostmanSchemaFilePath, schemaCollection)
	if err != nil {
		return err
	}

	// write the env to a file
	err = WriteFile(PostmanEnvFilePath, envCollection)
	if err != nil {
		return err
	}

	return nil
}

func NewPostmanCollectionInfo(appName string) PostmanCollectionInfo {
	return PostmanCollectionInfo{
		Name:   strings.ToUpper(appName),
		Schema: PostmanSchemaVersion,
	}
}

var PostmanRequestAuthItem = PostmanRequestAuth{
	Type: "bearer",
	Bearer: []PostmanBearer{
		{
			Key:   "token",
			Value: Placeholder(keyFirebaseIdToken),
			Type:  "string",
		},
	},
}

var AuthCollectionItem = PostmanCollectionItem{
	Name: "Auth",
	Item: []PostmanCollectionItemItem{
		createAuthItem("Register", "POST", "{\n  \"name\": \"admin\",\n  \"email\": \""+Placeholder(keyAdminEmail)+"\",\n    \"password\": \""+Placeholder(keyAdminPassword)+"\"\n}", Placeholder(keyBaseUrl)+"/auth/register"),
		createAuthItem("Login", "POST", "{\n    \"email\":\""+Placeholder(keyAdminEmail)+"\",\n    \"password\":\""+Placeholder(keyAdminPassword)+"\",\n    \"returnSecureToken\":true\n}\n", "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key="+Placeholder(keyFirebaseApiKey)),
	},
}

var FilesCollectionItem = PostmanCollectionItem{
	Name: "File",
	Item: []PostmanCollectionItemItem{
		createFileItem("Create File", "POST", "formdata", Placeholder(keyBaseUrl)+"/private/api/files/new"),
		createFileItem("Download File", "GET", "", Placeholder(keyBaseUrl)+"/private/api/files/"+Placeholder("FILE_ID")+"/download"),
		createFileItem("Delete File", "DELETE", "", Placeholder(keyBaseUrl)+"/private/api/files/"+Placeholder("FILE_ID")+"/delete"),
		createFileItem("List File", "GET", "", Placeholder(keyBaseUrl)+"/private/api/files?page=1&limit=10"),
		createFileItem("Update File", "PUT", "raw", Placeholder(keyBaseUrl)+"/private/api/files/"+Placeholder("FILE_ID")+"/update"),
		createFileItem("Detail File", "GET", "", Placeholder(keyBaseUrl)+"/private/api/files/"+Placeholder("FILE_ID")+""),
	},
}

func createAuthItem(name, method, body, url string) PostmanCollectionItemItem {
	return PostmanCollectionItemItem{
		Name: name,
		Request: PostmanCollectionItemItemRequest{
			Auth: PostmanRequestAuth{
				Type: "noauth",
			},
			Method: method,
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
			URL: parseURL(url),
		},
	}
}

func createFileItem(name, method, bodyMode, url string) PostmanCollectionItemItem {
	item := PostmanCollectionItemItem{
		Name: name,
		Request: PostmanCollectionItemItemRequest{
			Method: method,
			Header: []PostmanHeader{},
			URL:    parseURL(url),
		},
	}

	if bodyMode == "formdata" {
		item.Request.Body = PostmanRequestBody{
			Mode: "formdata",
			FormData: []PostmanFormDataItem{
				{
					Key:   "file",
					Type:  "file",
					Value: "<FILE_NAME>",
				},
			},
		}
	} else if bodyMode == "raw" {
		item.Request.Body = PostmanRequestBody{
			Mode: "raw",
			Raw:  "{\n    \"name\":\"<FILE_NAME>\"\n}\n",
			Options: PostmanRequestOptions{
				Raw: PostmanRequestOptionsRaw{
					Language: "json",
				},
			},
		}
	}

	return item
}

func parseURL(rawURL string) PostmanRequestURL {

	log.Debug().Str("rawURL", rawURL).Msg("parseURL")

	u, _ := url.Parse(rawURL)

	host := strings.Split(u.Host, ".")
	path := strings.Split(u.Path, "/")

	log.Debug().Strs("host", host).Strs("path", path).Str("zero", path[0]).Msg("parseURL")

	if path[0] == Placeholder(keyBaseUrl) {
		path = path[1:]
		host = []string{Placeholder(keyBaseUrl)}
	}

	log.Debug().Strs("host", host).Strs("path", path).Msg("parseURL")

	query := []PostmanQuery{}
	for key, values := range u.Query() {
		for _, value := range values {
			query = append(query, PostmanQuery{Key: key, Value: value})
		}
	}

	return PostmanRequestURL{
		Raw:   rawURL,
		Host:  host,
		Path:  path,
		Query: query,
	}
}

var preRequestScript = PostmanCollectionEvent{
	Listen: "prerequest",
	Script: PostmanCollectionEventScript{
		Type: "text/javascript",
		Exec: []string{
			"var origin = pm.environment.get(\"" + keyOrigin + "\")",
			"",
			"pm.request.headers.add({",
			"    key: \"Origin\",",
			"    value: origin",
			"});",
		},
	},
}

func GetPostmanCollection(appName string, resources []mgr.Resource) (*PostmanCollection, error) {
	if appName == "" {
		appName = "CollectionName"
	}

	collection := PostmanCollection{
		Info:  NewPostmanCollectionInfo(appName),
		Auth:  PostmanRequestAuthItem,
		Item:  []PostmanCollectionItem{AuthCollectionItem, FilesCollectionItem},
		Event: []PostmanCollectionEvent{preRequestScript},
	}

	for _, src := range resources {
		resourceName, _ := src.GetName()
		body, _ := mgr.InterfaceToMap(src.Model)
		bodyJSON, _ := json.MarshalIndent(body, "", "   ")

		resourceId := strings.ToUpper(resourceName) + "_ID"
		resourceIdPlaceHolder := Placeholder(resourceId)

		resourceItem := createResourceItem(resourceName)

		for _, route := range src.Routes {
			path := route.Path
			path = strings.Replace(path, "{id}", resourceIdPlaceHolder, 1)

			if route.RequiresAuth {
				path = "/private" + path
			}
			url := Placeholder(keyBaseUrl) + path

			for _, method := range route.Methods {

				isCreate := false
				if strings.Contains(path, "new") {
					isCreate = true
				}

				item := createResourceSubItem(route.Name, method, string(bodyJSON), url, resourceIdPlaceHolder, route.RequiresAuth, isCreate)
				resourceItem.Item = append(resourceItem.Item, item)
			}
		}

		collection.Item = append(collection.Item, resourceItem)
	}

	return &collection, nil
}

func createResourceItem(resourceName string) PostmanCollectionItem {
	return PostmanCollectionItem{
		Name: resourceName,
		Item: []PostmanCollectionItemItem{},
	}
}

func createResourceSubItem(routeName, method, body, url, idPlaceHolder string, requiresAuth bool, isCreate bool) PostmanCollectionItemItem {
	item := PostmanCollectionItemItem{
		Name: routeName,
		Request: PostmanCollectionItemItemRequest{
			Method: method,
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
			URL: parseURL(url),
		},
	}

	// Add authentication if required
	if requiresAuth {
		item.Request.Auth = PostmanRequestAuth{
			Type: "bearer",
			Bearer: []PostmanBearer{
				{
					Key:   "token",
					Value: Placeholder(keyFirebaseIdToken),
					Type:  "string",
				},
			},
		}
	} else {
		item.Request.Auth = PostmanRequestAuth{
			Type: "noauth",
		}
	}

	// Add event scripts for "Create" requests to set environment variables
	if isCreate {
		item.Event = []PostmanCollectionEvent{
			{
				Listen: "test",
				Script: PostmanCollectionEventScript{
					Type: "text/javascript",
					Exec: []string{
						"pm.environment.set(\"" + idPlaceHolder + "\", pm.response.json().data.ID);",
					},
				},
			},
		}
	}

	// Handle different request types
	switch method {
	case "POST":
		item.Request.Body = PostmanRequestBody{
			Mode: "raw",
			Raw:  body,
			Options: PostmanRequestOptions{
				Raw: PostmanRequestOptionsRaw{
					Language: "json",
				},
			},
		}
	case "PUT":
		item.Request.Body = PostmanRequestBody{
			Mode: "raw",
			Raw:  body,
			Options: PostmanRequestOptions{
				Raw: PostmanRequestOptionsRaw{
					Language: "json",
				},
			},
		}
	case "DELETE":
		// No body for DELETE requests
		item.Request.Body = PostmanRequestBody{}
	case "GET":
		// No body for GET requests
		item.Request.Body = PostmanRequestBody{}
	}

	return item
}

func GetPostmanEnv(
	appName,
	environment,
	baseUrl,
	adminEmail,
	adminPassword,
	firebaseApiKey string,

) (*PostmanEnv, error) {
	if appName == "" {
		appName = "CollectionName"
	}

	if environment == "" {
		return nil, fmt.Errorf("missing environment")
	}
	envSchema := PostmanEnv{
		PostmanVariableScope: PostmanVariableScope,
		Values:               make([]PostmanEnvValue, 0),
	}
	envSchema.Name = strings.ToUpper(appName) + "_" + strings.ToUpper(environment)

	// Url
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyBaseUrl,
		Value:   baseUrl,
		Enabled: true,
	})

	// Email
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyAdminEmail,
		Value:   adminEmail,
		Enabled: true,
	})

	// Password
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyAdminPassword,
		Value:   adminPassword,
		Enabled: true,
	})

	// FirebaseAPIkey
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyFirebaseApiKey,
		Value:   firebaseApiKey,
		Enabled: true,
	})

	// FirebaseIdToken
	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyFirebaseIdToken,
		Value:   "",
		Enabled: true,
	})

	envSchema.Values = append(envSchema.Values, PostmanEnvValue{
		Key:     keyOrigin,
		Value:   baseUrl,
		Enabled: true,
	})

	return &envSchema, nil
}

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

	return nil
}
