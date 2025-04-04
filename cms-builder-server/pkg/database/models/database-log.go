package models

import (
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	"gorm.io/gorm"
)

type DatabaseLog struct {
	gorm.Model
	UserId       uint               `gorm:"type:bigint,default:0" json:"userId"`
	Username     string             `json:"username"`
	Action       dbTypes.CRUDAction `json:"action"`
	ResourceName string             `json:"resourceName"`
	ResourceId   string             `json:"resourceId"`
	Timestamp    string             `gorm:"type:timestamp" json:"timestamp"`
	Detail       string             `json:"detail"`
	TraceId      string             `json:"traceId"`
}
