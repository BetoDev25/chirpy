package main

import (
	"strings"
	"net/http"
	"database/sql"
	"errors"
	"time"

	"github.com/BetoDev25/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusBadRequest, "could not get header")
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), tokenString)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 401, "Refresh token does not exist or expired")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get refresh token")
			return
		}
	}

	jwtToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not generate jwt token")
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, 200, response{
		Token: jwtToken,
	})
}
