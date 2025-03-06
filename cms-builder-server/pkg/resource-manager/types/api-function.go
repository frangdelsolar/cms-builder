package types

import (
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"

	"net/http"
)

// ApiFunction defines the signature for API handler functions.
type ApiFunction func(resource *Resource, db *dbTypes.DatabaseConnection) http.HandlerFunc
