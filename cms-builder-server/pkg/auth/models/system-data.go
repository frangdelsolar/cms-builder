package models

import (
	"fmt"

	"gorm.io/gorm"
)

type SystemData struct {
	gorm.Model
	CreatedByID uint  `gorm:"not null" json:"createdById" jsonschema:"title=Created By Id,description=Id of the user who created this record"`
	CreatedBy   *User `gorm:"foreignKey:CreatedByID" json:"createdBy" jsonschema:"title=Created By,description=User who created this record"`
	UpdatedByID uint  `gorm:"not null" json:"updatedById" jsonschema:"title=Updated By Id,description=Id of the user who updated this record"`
	UpdatedBy   *User `gorm:"foreignKey:UpdatedByID" json:"updatedBy" jsonschema:"title=Updated By,description=User who updated this record"`
}

func (s *SystemData) StringID() string {
	return fmt.Sprint(s.ID)
}
