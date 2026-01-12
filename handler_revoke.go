package main

import (
	"strings"
	"net/http"
	"errors"
	"database/sql"
)

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusBadRequest, "could not get header")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	err := cfg.db.RevokeRefreshToken(r.Context(), tokenString)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 401, "Refresh token does not exist or expired")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get refresh token")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)
}
