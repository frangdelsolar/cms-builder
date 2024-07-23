package cms_server

import (
	"fmt"
	"reflect"
)

type Entity struct {
	Model interface{}
}

func (e *Entity) Name() string {
	return fmt.Sprintf("%T", e.Model)
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
