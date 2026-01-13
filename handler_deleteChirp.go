package main

import (
	"strings"
	"net/http"
	"errors"
	"database/sql"

	"github.com/google/uuid"

	"github.com/BetoDev25/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusUnauthorized, "could not get header")
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 403, "Couldn't validate jwt token")
		return
	}

	chirpIDStr := r.PathValue("chirpID")
	id, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, 400, "Couldn't parse Chirp ID")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 404, "Chirp does not exist")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get Chirp")
			return
		}
	}

	if userID != chirp.UserID {
		respondWithError(w, 403, "Not owner of this chirp")
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete Chirp")
		return
	}

	respondWithJSON(w, 204, struct{}{})
}
