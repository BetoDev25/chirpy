package main

import (
	"encoding/json"
	"net/http"
	"time"
	"fmt"

	"github.com/google/uuid"

	"github.com/BetoDev25/chirpy/internal/auth"
	"github.com/BetoDev25/chirpy/internal/database"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	//Token          string    `json:"token"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	input := params{}
	err := decoder.Decode(&input)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode input")
		return
	}

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          input.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		fmt.Println("CreateUser error:", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID: 	   user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
