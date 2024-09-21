package builder

import (
	"bytes"
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

	var localUser User

	q := "firebase_id = '" + accessToken.UID + "'"
	b.db.Find(&localUser, q)

	return &localUser, nil
}

func (b *Builder) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := r.Header.Get("Authorization")
		if header != "" {
			// get token from header
			token := strings.Split(header, " ")[1]
			if token != "" {
				localUser, err := b.VerifyUser(token)
				if err != nil {
					log.Error().Err(err).Msg("Error verifying user")
				}
				if localUser != nil {
					r.Header.Set("requested_by", fmt.Sprint(localUser.ID))
				}
			}
		} else {
			r.Header.Set("requested_by", "")
		}

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

	userApp, err := b.admin.GetApp("user")
	if err != nil {
		log.Error().Err(err).Msg("Error getting user app")
		http.Error(w, "Error getting user app", http.StatusInternalServerError)
		return
	}

	// set body
	userRequestBody := map[string]string{
		"name":        input.Name,
		"email":       input.Email,
		"firebase_id": fbUser.UID,
	}

	bodyBytes, err := json.Marshal(userRequestBody)
	if err != nil {
		log.Error().Err(err).Msg("Error marshalling user request body")
		http.Error(w, "Error marshalling user request body", http.StatusInternalServerError)
		return
	}

	userRequest := &http.Request{
		Method: http.MethodPost,
		Header: r.Header,
		Body:   io.NopCloser(bytes.NewBuffer(bodyBytes)),
	}

	userApp.ApiNew(b.db)(w, userRequest)

	// TODO: Should rollback firebase if unsuccessful

}
