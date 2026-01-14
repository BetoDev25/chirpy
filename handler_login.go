package main

import (
	"encoding/json"
	"database/sql"
	"net/http"
	"time"

	"github.com/BetoDev25/chirpy/internal/auth"
	"github.com/BetoDev25/chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	input := params{}
	err := decoder.Decode(&input)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode input")
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	isValid, err := auth.CheckPasswordHash(input.Password, user.HashedPassword)
	if err != nil || !isValid {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	duration := time.Hour

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, duration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate token")
		return
	}

	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate refresh token string")
		return
	}

	_, err = cfg.db.MakeRefreshToken(r.Context(), database.MakeRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
		RevokedAt: sql.NullTime{Valid: false},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token")
		return
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	respondWithJSON(w, http.StatusOK, response {
			User: User{
				ID:           user.ID,
				CreatedAt:    user.CreatedAt,
				UpdatedAt:    user.UpdatedAt,
				Email:        user.Email,
				IsChirpyRed:  user.IsChirpyRed,
			},
			Token:        token,
			RefreshToken: refreshTokenString,
	})
}
