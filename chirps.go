package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Quak1/chirpy/internal/auth"
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
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "failed to get bearer token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		respondJSONError(w, http.StatusUnauthorized, "failed to validate JWT", err)
		return
	}

	params := parameters{}
	err = json.NewDecoder(r.Body).Decode(&params)
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
		UserID: userID,
	})
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to create chirp", err)
		return
	}

	respondJSON(w, http.StatusCreated, Chirp(chirp))
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to get chirps", err)
		return
	}

	jsonChirps := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		jsonChirps[i] = Chirp(chirp)
	}

	respondJSON(w, http.StatusOK, jsonChirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "failed to parse chirp id", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondJSONError(w, http.StatusNotFound, "failed to get chirp", err)
		return
	}

	respondJSON(w, http.StatusOK, Chirp(chirp))
}

func (cfg *apiConfig) handlerDelteChirp(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondJSONError(w, http.StatusUnauthorized, "failed to get bearer token", err)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.tokenSecret)
	if err != nil {
		respondJSONError(w, http.StatusUnauthorized, "failed to validate JWT", err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "failed to parse chirp id", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondJSONError(w, http.StatusNotFound, "failed to get chirp", err)
		return
	}

	if chirp.UserID != userID {
		respondJSONError(w, http.StatusForbidden, "author error", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "couldn't delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
