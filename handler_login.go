package main

import (
	"encoding/json"
	"net/http"

	"github.com/BetoDev25/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Password string `json:"password"`
		Email     string `json:"email"`
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

	respondWithJSON(w, http.StatusOK, User {
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
	})
}
