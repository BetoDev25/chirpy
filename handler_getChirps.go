package main

import (
	"net/http"
	"sort"

	"github.com/google/uuid"

	"github.com/BetoDev25/chirpy/internal/database"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	array := []Chirp{}
	apiChirps := []database.Chirp{}
	var err error

	author_id := r.URL.Query().Get("author_id")
	sortType := r.URL.Query().Get("sort")

	if author_id == "" {
		apiChirps, err = cfg.db.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get Chirps")
			return
		}
	} else {
		userID, err := uuid.Parse(author_id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "invalid ID")
			return
		}
		apiChirps, err = cfg.db.GetChirpsByID(r.Context(), userID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get Chirps")
			return
		}
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

	if sortType == "desc" {
		sort.Slice(array, func(i, j int) bool {
			return array[i].CreatedAt.After(array[j].CreatedAt)
		})
	}

	respondWithJSON(w, 200, array)
}
