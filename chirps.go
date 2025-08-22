package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Quak1/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to parse request body", err)
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: params.UserID,
	})
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to create chirp", err)
		return
	}

	respondJSON(w, http.StatusCreated, Chirp(chirp))
}
