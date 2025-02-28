package database

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"gorm.io/gorm"
)

type CRUDAction string

const (
	CreateCRUDAction CRUDAction = "created"
	UpdateCRUDAction CRUDAction = "updated"
	DeleteCRUDAction CRUDAction = "deleted"
)

type DatabaseLog struct {
	gorm.Model
	User         *models.User `json:"user"`
	UserId       string       `gorm:"foreignKey:UserId" json:"userId"`
	Username     string       `json:"username"`
	Action       CRUDAction   `json:"action"`
	ResourceName string       `json:"resourceName"`
	ResourceId   string       `json:"resourceId"`
	Timestamp    string       `gorm:"type:timestamp" json:"timestamp"`
	Detail       string       `json:"detail"`
	TraceId      string       `json:"traceId"`
}
