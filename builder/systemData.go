package builder

import (
	"fmt"

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

// ID returns the ID of the SystemData as a string.
//
// Returns:
// - string: the ID of the SystemData.
func (s *SystemData) GetIDString() string {
	return fmt.Sprint(s.ID)
}
