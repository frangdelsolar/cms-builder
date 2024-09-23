package test_helpers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/frangdelsolar/cms/builder"
	"github.com/stretchr/testify/assert"
)

type MockStruct struct {
	*builder.SystemData
	Field string
}

// FieldValidator validates the given field.
//
// It checks if the field value is empty. If the field value is empty, it returns an error with the message
// "{field name} cannot be empty". Otherwise, it returns nil.
//
// Parameters:
// - fieldName: the name of the field to be validated.
// - instance: a map[string]interface{} representing the instance to be validated.
//
// Returns:
// - error: an error if the field value is empty, otherwise nil.
func FieldValidator(fieldName string, instance map[string]interface{}, output *builder.FieldValidationError) *builder.FieldValidationError {
	fieldValue := fmt.Sprint(instance[fieldName])
	if fieldValue == "" {
		output.Error = fieldName + " cannot be empty"
	}

	return output
}

type EngineServices struct {
	Engine   *builder.Builder
	Admin    *builder.Admin
	DB       *builder.Database
	Server   *builder.Server
	Firebase *builder.FirebaseAdmin
	App      *builder.App
}

// GetDefaultEngine returns a default Builder instance, Admin, Database, Server, and App instances,
// and a callback to be used to deregister the App after the test is finished.
//
// It sets up a default Builder instance with the given configuration, an Admin instance,
// a Database instance, a Server instance, and an App instance with a MockStruct type.
// The function also sets up a field validator for the "field" key on the MockStruct type.
// The function returns the instances and a callback to be used to deregister the App after the test is finished.
func GetDefaultEngine() EngineServices {
	input := &builder.NewBuilderInput{
		ReadConfigFromFile: true,
		ConfigFilePath:     "config.yaml", // Replace with a valid config file path
		InitializeLogger:   true,
		InitiliazeDB:       true,
		InitiliazeServer:   true,
		InitiliazeAdmin:    true,
		InitiliazeFirebase: true,
	}

	var err error
	engine, err := builder.NewBuilder(input)
	admin, err := engine.GetAdmin()
	db, err := engine.GetDatabase()
	server, err := engine.GetServer()
	firebase, err := engine.GetFirebase()

	app, err := admin.Register(MockStruct{}, false)
	app.RegisterValidator("field", FieldValidator)
	defer admin.Unregister(app.Name())

	if err != nil {
		panic(err)
	}
	return EngineServices{engine, admin, db, server, firebase, &app}
}

// createMockResource creates a new resource for the given user and returns the created resource, the user, and a function to roll back the resource creation.
//
// It creates a new request with a random string as the value for the "field" key, then calls the app's ApiNew method to create a new resource. It then unmarshalls the response into the createdItem variable and returns it along with the user and the rollback function.
func CreateMockResource(t *testing.T, db *builder.Database, app *builder.App, user *builder.User) (*MockStruct, *builder.User, func()) {

	data := "{\"field\": \"" + RandomString(10) + "\"}"
	request, user, rollback := NewRequest(http.MethodPost, data, true, user, nil)

	t.Logf("Creating new resource for user: %v", user.ID)

	var createdItem MockStruct
	response, err := ExecuteApiCall(t, app.ApiNew(db), request, &createdItem)

	assert.NoError(t, err, "ApiNew should not return an error")
	assert.True(t, response.Success, "ApiNew should return a success response")
	return &createdItem, user, rollback
}

// executeApiCall executes the given API call handler function with the given request and stores the response in the given value.
//
// It logs the request method and body, creates a new MockWriter, and calls the API call handler function with the MockWriter and the request. It then parses the response from the MockWriter and stores it in the given value. Finally, it asserts that the parsing did not return an error.
//
// Parameters:
// - t: the testing.T instance.
// - apiCall: the API call handler function to be executed.
// - request: the request to be passed to the API call handler function.
// - v: the value to store the response in.
//
// Returns:
// - builder.Response: the parsed response from the API call handler function.
func ExecuteApiCall(t *testing.T, apiCall builder.HandlerFunc, request *http.Request, v interface{}) (builder.Response, error) {
	t.Log("Executing API call", request.Method, request.Body)
	writer := MockWriter{}
	apiCall(&writer, request)

	return builder.ParseResponse(writer.Buffer.Bytes(), v)
}
