package builder

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

type SystemData struct {
	gorm.Model
	CreatedByID uint  `gorm:"not null" json:"createdById" jsonschema:"title=Created By Id,description=Id of the user who created this record"`
	CreatedBy   *User `gorm:"foreignKey:CreatedByID" json:"createdBy" jsonschema:"title=Created By,description=User who created this record"`
	UpdatedByID uint  `gorm:"not null" json:"updatedById" jsonschema:"title=Updated By Id,description=Id of the user who updated this record"`
	UpdatedBy   *User `gorm:"foreignKey:UpdatedByID" json:"updatedBy" jsonschema:"title=Updated By,description=User who updated this record"`
}

// Returns a map with the json representation of the fields
func (s *SystemData) Keys() []string {

	var keys = []string{
		"id",
	}
	rt := reflect.TypeOf(SystemData{})
	if rt.Kind() != reflect.Struct {
		panic("bad type")
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		v := strings.Split(f.Tag.Get("json"), ",")[0] // use split to ignore tag "options" like omitempty, etc.
		if v != "" {
			keys = append(keys, v)
		}
	}

	return keys
}

// ID returns the ID of the SystemData as a string.
//
// Returns:
// - string: the ID of the SystemData.
func (s *SystemData) GetIDString() string {
	return fmt.Sprint(s.ID)
}
