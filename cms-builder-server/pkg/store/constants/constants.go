package constants

import (
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
)

const (
	StoreLocal    storeTypes.StoreType = "local"
	StoreS3       storeTypes.StoreType = "s3"
	StoreFilebase storeTypes.StoreType = "filebase"
)
