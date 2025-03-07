package auth

import (
	"context"
	"net/http"
	"time"

	authConstants "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/constants"
	"github.com/google/uuid"
)

// UserCookieMiddleware extracts the user_id cookie and stores it in the request context
func UserCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the user_id cookie
		cookie, err := r.Cookie(authConstants.UserCookieName)
		var userID string

		if err == http.ErrNoCookie {
			// If the cookie is missing, generate a new user_id
			userID = uuid.New().String()
		} else if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		} else {
			// Use the existing user_id
			userID = cookie.Value
		}

		// Set the user_id cookie in the response (if it was missing or regenerated)
		if err == http.ErrNoCookie {
			http.SetCookie(w, &http.Cookie{
				Name:     authConstants.UserCookieName,
				Value:    userID,
				Expires:  time.Now().Add(365 * 24 * time.Hour), // Expires in 1 year
				Path:     "/",                                  // Accessible across the entire domain
				HttpOnly: true,                                 // Prevent client-side JavaScript access
				Secure:   true,                                 // FIXME: Set to true in production (requires HTTPS)
				SameSite: http.SameSiteNoneMode,                // Prevent CSRF attacks
			})
		}

		// Add the user_id to the request context
		ctx := context.WithValue(r.Context(), authConstants.CtxRequestUserCookie, userID)
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
