package cms_server

import (
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
	t := reflect.TypeOf(e.Model)
	// Check if it's a pointer and dereference it
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Ensure it's a struct type
	if t.Kind() != reflect.Struct {
		return nil
	}

	var fieldNames []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldNames = append(fieldNames, field.Name)
	}
	return fieldNames
}
