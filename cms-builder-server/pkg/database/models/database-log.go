package models

import (
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	"gorm.io/gorm"
)

type DatabaseLog struct {
	gorm.Model
	User         *authModels.User   `json:"user"`
	UserId       string             `gorm:"foreignKey:UserId" json:"userId"`
	Username     string             `json:"username"`
	Action       dbTypes.CRUDAction `json:"action"`
	ResourceName string             `json:"resourceName"`
	ResourceId   string             `json:"resourceId"`
	Timestamp    string             `gorm:"type:timestamp" json:"timestamp"`
	Detail       string             `json:"detail"`
	TraceId      string             `json:"traceId"`
}
