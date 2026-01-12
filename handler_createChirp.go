package main

import (
	"encoding/json"
	"net/http"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/BetoDev25/chirpy/internal/database"
	"github.com/BetoDev25/chirpy/internal/auth"
)

type Chirp struct {
	ID       uuid.UUID  `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func validateChirp(body string) (string, error){
        if len(body) > 140 {
                return "", fmt.Errorf("Chirp longer than 140 characters")
        }

        words := strings.Split(body, " ")
        for i, word := range words {
                lower := strings.ToLower(word)
                if lower == "kerfuffle" || lower == "sharbert" || lower == "fornax" {
                        words[i] = "****"
                }
        }
        return strings.Join(words, " "), nil
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't get token")
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't validate jwt")
		return
	}

	decoder := json.NewDecoder(r.Body)
	input := params{}
	err = decoder.Decode(&input)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode Chirp")
		return
	}

	cleaned, err := validateChirp(input.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create Chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
