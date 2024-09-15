package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (b *Builder) VerifyUser(userIdToken string) (*User, error) {
	// verify token
	firebase, err := b.GetFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Error getting firebase")
		return nil, err
	}

	accessToken, err := firebase.VerifyIDToken(context.Background(), userIdToken)
	if err != nil {
		log.Error().Err(err).Msg("Error verifying token")
		return nil, err
	}

	log.Debug().Interface("AccessToken", accessToken).Msg("AccessToken")

	var localUser User

	q := "firebase_id = '" + accessToken.UID + "'"

	b.db.Find(&localUser, q)

	log.Debug().Interface("LocalUser", localUser).Msg("LocalUser")

	return &localUser, nil
}

func (b *Builder) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("AuthMiddleware")

		// get authentication header
		header := r.Header.Get("Authorization")
		if header == "" {
			log.Debug().Msg("No authorization header found")
		} else {
			// get token from header
			token := strings.Split(header, " ")[1]
			if token == "" {
				log.Debug().Msg("No token found in authorization header")
			} else {
				localUser, err := b.VerifyUser(token)
				if err != nil {
					log.Error().Err(err).Msg("Error verifying user")
				}
				if localUser != nil {
					r.Header.Set("user_id", fmt.Sprint(localUser.ID))
				}
			}
		}

		log.Debug().Interface("Request", r.Header).Msg("Request")

		next.ServeHTTP(w, r)
	})
}

func (b *Builder) RegisterUserController(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var input RegisterUserInput
	err = json.Unmarshal(body, &input)
	if err != nil {
		http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
		return
	}
	fb, err := b.GetFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Error getting firebase")
		http.Error(w, "Error getting firebase", http.StatusInternalServerError)
		return
	}
	fbUser, err := fb.RegisterUser(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("Error registering user")
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	log.Debug().Interface("Firebase User", fbUser).Msg("LocalUser")

	localUser, err := NewUser(input.Name, input.Email)
	if err != nil {
		log.Error().Err(err).Msg("Error creating local user")
		http.Error(w, "Error creating local user", http.StatusInternalServerError)
		return
	}

	localUser.FirebaseId = fbUser.UID

	b.db.Create(localUser)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(localUser)
}
