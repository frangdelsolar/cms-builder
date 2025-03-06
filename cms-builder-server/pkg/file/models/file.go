package models

import (
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
)

type File struct {
	*authModels.SystemData
	Name          string `json:"name"`
	Path          string `json:"path"` // relative path
	Url           string `json:"url"`  // absolute path
	Size          int64  `json:"size"`
	MimeType      string `json:"mimeType"`
	DownloadCount int64  `json:"downloadCount"`
}
