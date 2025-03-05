package store

import (
	"mime/multipart"

	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
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
	MediaFolder        string // the folder where the files are stored i. e. media/easy-files/
}

type Store interface {
	GetPath(file *models.File) string
	StoreFile(fileName string, file multipart.File, header *multipart.FileHeader, log *loggerTypes.Logger) (fileData *models.File, err error)
	DeleteFile(file *models.File, log *loggerTypes.Logger) error
	ListFiles(log *loggerTypes.Logger) ([]string, error)
	ReadFile(file *models.File, log *loggerTypes.Logger) ([]byte, error)
	GetFileInfo(file *models.File, log *loggerTypes.Logger) (*models.FileInfo, error)
	GetConfig() *StoreConfig
}
