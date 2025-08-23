package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Quak1/chirpy/internal/auth"
	"github.com/Quak1/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to parse request body", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to create user", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to create user", err)
		return
	}

	respondJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to parse request body", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondJSONError(w, http.StatusNotFound, "user not found", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondJSONError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	expiresIn := time.Second * time.Duration(params.ExpiresInSeconds)
	if params.ExpiresInSeconds > 60*60 || params.ExpiresInSeconds <= 0 {
		expiresIn = time.Hour
	}

	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, expiresIn)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "error making JWT", err)
	}

	respondJSON(w, http.StatusOK, struct {
		User
		Token string `json:"token"`
	}{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token: token,
	})
}
