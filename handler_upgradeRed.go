package main

import (
	"errors"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/BetoDev25/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUpgradeRed(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Event string `json:"event"`
		Data struct{
			UserID uuid.UUID `json:"user_id"`
		}`json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not get API key")
		return
	}
	if apiKey != cfg.apiKey {
		respondWithError(w, http.StatusUnauthorized, "incorrect API key")
		return
	}

	decoder := json.NewDecoder(r.Body)
	input := params{}
	err = decoder.Decode(&input)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode input")
		return
	}

	if input.Event != "user.upgraded" {
		respondWithJSON(w, 204, struct{}{})
		return
	}

	err = cfg.db.UpgradeUserRed(r.Context(), input.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "User does not exist")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't upgrade user")
			return
		}
	}

	respondWithJSON(w, 204, struct{}{})
}
