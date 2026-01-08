package main

import (
	"net/http"

	"github.com/BetoDev25/chirpy/internal/database"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	array := []Chirp{}
	apiChirps := []database.Chirp{}

	apiChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get Chirps")
	}

	for _, chirp := range apiChirps {
		newChirp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		array = append(array, newChirp)
	}

	respondWithJSON(w, 200, array)
}
