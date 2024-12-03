package builder

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// IMPORTANT: If you ever modify this struct you must also modify
// removeSystemDataFieldsFromRequest from app.go
type SystemData struct {
	gorm.Model
	CreatedByID uint  `gorm:"not null" json:"createdById"`
	CreatedBy   *User `gorm:"foreignKey:CreatedByID" json:"createdBy"`
	UpdatedByID uint  `gorm:"not null" json:"updatedById"`
	UpdatedBy   *User `gorm:"foreignKey:UpdatedByID" json:"updatedBy"`
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
