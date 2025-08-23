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
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
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
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
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

	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "error making JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "error generating refresh token", err)
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60), // 60 days
	})
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "error generating refresh token", err)
		return
	}

	respondJSON(w, http.StatusOK, struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "failed to get bearer token", err)
		return
	}

	dbRefreshToken, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondJSONError(w, http.StatusUnauthorized, "failed to get refresh token", err)
		return
	}

	if dbRefreshToken.RevokedAt.Valid {
		respondJSONError(w, http.StatusUnauthorized, "revoked token", err)
		return
	}

	isExpired := dbRefreshToken.ExpiresAt.Before(time.Now())
	if isExpired {
		respondJSONError(w, http.StatusUnauthorized, "expired refresh token", err)
		return
	}

	jwtToken, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "error making JWT", err)
		return
	}

	respondJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: jwtToken,
	})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "failed to get bearer token", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to find refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
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

	params := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "failed to parse request body", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, "failed to genereate hashed password", err)
		return
	}

	updatedUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		ID:             userID,
		HashedPassword: hashedPassword,
		UpdatedAt:      time.Now(),
	})

	respondJSON(w, http.StatusOK, User{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	})
}

func (cfg *apiConfig) handlerUpgradeToChirpyRed(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		respondJSONError(w, http.StatusUnauthorized, "missing API key", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondJSONError(w, http.StatusUnauthorized, "wrong API key", err)
		return
	}

	params := struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}{}
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "failed to parse request body", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = cfg.db.UpgradeChirpyRed(r.Context(), params.Data.UserID)
	if err != nil {
		respondJSONError(w, http.StatusNotFound, "user not found", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
