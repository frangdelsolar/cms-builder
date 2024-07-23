package cms_server

import (
	"fmt"
	"reflect"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var log *zerolog.Logger

type Config struct {
	Logger *zerolog.Logger
	DB     *gorm.DB
}

func Setup(config *Config) {
	log = config.Logger
	log.Info().Msg("Setting up CMS server")
}

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

func Register(model interface{}) {
	log.Info().Msgf("Registering model: %s", model)

	entity := Entity{
		Model: model,
	}

	log.Info().Interface("entity", entity).Msgf("Model registered")
	log.Debug().Msgf("Name: %s", entity.Name())

	log.Debug().Msgf("Fields: %s", entity.Fields())

}
