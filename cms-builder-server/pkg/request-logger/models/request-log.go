package models

import (
	"time"

	"gorm.io/gorm"
)

// RequestIDKey is the key used to store the request ID in the context.
type RequestLog struct {
	gorm.Model
	Timestamp  time.Time `gorm:"type:timestamp" json:"timestamp"`
	Duration   int64     `json:"duration"`
	Ip         string    `json:"ip"`
	Origin     string    `json:"origin"`
	Referer    string    `json:"referrer"`
	UserId     uint      `json:"userId"`
	UserLabel  string    `json:"userLabel"`
	Roles      string    `json:"roles"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Query      string    `json:"query"`
	StatusCode string    `json:"statusCode"`
	Error      string    `json:"error"`
	Header     string    `json:"header"`
	Body       string    `json:"body"`
	Response   string    `json:"response"`
	TraceId    string    `json:"traceId"`
}
