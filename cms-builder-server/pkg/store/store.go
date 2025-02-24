package store

import (
	"mime/multipart"
	"time"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
)

type StoreType string

const (
	StoreLocal StoreType = "local"
	StoreS3    StoreType = "s3"
)

type StoreConfig struct {
	MaxSize            int64
	SupportedMimeTypes []string
	Folder             string
}

type FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified,omitempty"`
	ContentType  string    `json:"content_type,omitempty"`
}

type Store interface {
	GetPath() string
	StoreFile(cfg *StoreConfig, fileName string, file multipart.File) (fileData models.File, err error)
	DeleteFile(file models.File) error
	ListFiles() ([]string, error)
	ReadFile(file *models.File) ([]byte, error)
	GetFileInfo(file *models.File) (*FileInfo, error)
}
