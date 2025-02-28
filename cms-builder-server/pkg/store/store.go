package store

import (
	"mime/multipart"

	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger"
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

type Store interface {
	GetPath() string
	StoreFile(fileName string, file multipart.File, header *multipart.FileHeader, log *logger.Logger) (fileData *models.File, err error)
	DeleteFile(file *models.File, log *logger.Logger) error
	ListFiles(log *logger.Logger) ([]string, error)
	ReadFile(file *models.File, log *logger.Logger) ([]byte, error)
	GetFileInfo(file *models.File, log *logger.Logger) (*models.FileInfo, error)
	GetConfig() *StoreConfig
}
