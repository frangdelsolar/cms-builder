package auth

import (
	"context"
	"net/http"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	authUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/utils"
	cliPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/clients"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	svrConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/constants"
	svrUtils "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/utils"
)

const GodTokenHeader = "X-God-Token"

// AuthMiddleware is a middleware function that verifies the user based on the
// access token provided in the Authorization header of the request. If the
// verification fails, it will return a 401 error. If the verification is
// successful, it will continue to the next handler in the chain, setting a
// "requested_by" header in the request with the ID of the verified user.
func AuthMiddleware(envGodToken string, godUser *authModels.User, firebase *cliPkg.FirebaseManager, db *dbTypes.DatabaseConnection, systemUser *authModels.User) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Check if the request has a god token
			headerGodToken := r.Header.Get(GodTokenHeader)

			accessToken := svrUtils.GetRequestAccessToken(r)
			requestId := svrUtils.GetRequestId(r)
			log := svrUtils.GetRequestLogger(r)

			localUser := &authModels.User{}
			if headerGodToken != "" && authUtils.VerifyGodUser(envGodToken, headerGodToken) {
				localUser = godUser
			}

			var err error
			if localUser.ID == 0 && accessToken != "" {
				localUser, err = authUtils.VerifyUser(accessToken, firebase, db, systemUser, requestId, log)
				if err != nil {
					log.Error().Err(err).Msg("Error verifying user. User may not be authenticated")
				}
			}

			if localUser != nil {

				// Create a new context with both values
				ctx := r.Context()
				ctx = context.WithValue(ctx, svrConstants.CtxRequestIsAuth, true)
				ctx = context.WithValue(ctx, svrConstants.CtxRequestUser, localUser)

				// Update the request with the new context
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
