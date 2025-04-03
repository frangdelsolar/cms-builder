package models

import (
	"time"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	"gorm.io/gorm"
)

// RequestIDKey is the key used to store the request ID in the context.
type RequestLog struct {
	gorm.Model
	Timestamp  time.Time        `gorm:"type:timestamp" json:"timestamp"`
	Duration   int64            `json:"duration"`
	Ip         string           `json:"ip"`
	Origin     string           `json:"origin"`
	Referer    string           `json:"referrer"`
	UserId     *string          `gorm:"foreignKey:UserId" json:"userId"`
	UserLabel  string           `json:"userLabel"`
	User       *authModels.User `json:"user,omitempty"`
	Roles      string           `json:"roles"`
	Method     string           `json:"method"`
	Path       string           `json:"path"`
	Query      string           `json:"query"`
	StatusCode string           `json:"statusCode"`
	Error      string           `json:"error"`
	Header     string           `json:"header"`
	Body       string           `json:"body"`
	Response   string           `json:"response"`
	TraceId    string           `json:"traceId"`
}
