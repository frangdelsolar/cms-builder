package types

import (
	"mime/multipart"

	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	fileTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/types"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

type Store interface {
	GetPath(file *fileModels.File) string
	StoreFile(fileName string, file multipart.File, header *multipart.FileHeader, log *loggerTypes.Logger) (fileData *fileModels.File, err error)
	DeleteFile(file *fileModels.File, log *loggerTypes.Logger) error
	ListFiles(log *loggerTypes.Logger) ([]string, error)
	ReadFile(file *fileModels.File, log *loggerTypes.Logger) ([]byte, error)
	GetFileInfo(file *fileModels.File, log *loggerTypes.Logger) (*fileTypes.FileInfo, error)
	GetConfig() *StoreConfig
}
