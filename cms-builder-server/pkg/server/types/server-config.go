package types

import (
	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

type ServerConfig struct {
	Host           string
	Port           string
	CsrfToken      string
	AllowedOrigins []string
	LoggerConfig   *loggerTypes.LoggerConfig
	GodToken       string
	GodUser        *authModels.User
	SystemUser     *authModels.User
	Firebase       *cliPkg.FirebaseManager
}
