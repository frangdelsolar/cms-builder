package builder

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

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

// TODO: Do not use in prod!!!!!
type Response struct {
	Success bool          `json:"success"`
	Data    LoginResponse `json:"data"`
}
type LoginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func (b *Builder) LoginController(w http.ResponseWriter, r *http.Request) {
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

	query := "email = '" + input.Email + "'"
	var user User

	b.db.Find(&user, query)

	fb, err := b.GetFirebase()
	if err != nil {
		log.Error().Err(err).Msg("Error getting firebase")
		http.Error(w, "Error getting firebase", http.StatusInternalServerError)
		return
	}

	u, err := fb.CustomToken(context.Background(), user.FirebaseId)
	if err != nil {
		log.Error().Err(err).Msg("Error creating custom token")
		http.Error(w, "Error creating custom token", http.StatusInternalServerError)
		return
	}

	output := Response{
		Success: true,
		Data: LoginResponse{
			Token: u,
			Email: input.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
