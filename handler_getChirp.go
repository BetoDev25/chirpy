package main

import (
	"net/http"
	"errors"
	"database/sql"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
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

	respondWithJSON(w, 200, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
