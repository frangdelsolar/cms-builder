package cms_admin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	pluralize "github.com/gertd/go-pluralize"
)

type Entity struct {
	Model interface{}
}

func (e *Entity) Name() string {
	modelName := fmt.Sprintf("%T", e.Model)
	name := modelName[strings.LastIndex(modelName, ".")+1:]
	name = strings.ToLower(name)
	return name
}

func (e *Entity) Plural() string {
	p := pluralize.NewClient()
	return p.Plural(e.Name())
}

func (e *Entity) Fields() []string {

	instanceType := reflect.TypeOf(e.Model)
	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}
	instance := reflect.New(instanceType).Interface()

	jsonDict, err := json.Marshal(instance)
	if err != nil {
		log.Error().Err(err).Msgf("Error marshalling %s record to JSON", e.Name())
		return nil
	}
	log.Debug().Interface("response", string(jsonDict)).Msg("Fields")

	// Iterate over the fields
	fields := make(map[string]interface{})
	err = json.Unmarshal(jsonDict, &fields)
	if err != nil {
		log.Error().Err(err).Msgf("Error unmarshalling %s record to JSON", e.Name())
		return nil
	}

	var fieldNames []string
	for k := range fields {
		fieldNames = append(fieldNames, k)
	}

	return fieldNames
}
