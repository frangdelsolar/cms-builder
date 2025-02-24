package models

import "time"

type FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified,omitempty"`
	ContentType  string    `json:"content_type,omitempty"`
}
