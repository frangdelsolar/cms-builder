package database

import (
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"gorm.io/gorm"
)

type DatabaseLog struct {
	gorm.Model
	User         *models.User       `json:"user"`
	UserId       string             `gorm:"foreignKey:UserId" json:"userId"`
	Username     string             `json:"username"`
	Action       dbTypes.CRUDAction `json:"action"`
	ResourceName string             `json:"resourceName"`
	ResourceId   string             `json:"resourceId"`
	Timestamp    string             `gorm:"type:timestamp" json:"timestamp"`
	Detail       string             `json:"detail"`
	TraceId      string             `json:"traceId"`
}
