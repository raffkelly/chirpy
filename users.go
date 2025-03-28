package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error decoding", err)
		return
	}
	if params.Email == "" || (!strings.Contains(params.Email, "@")) {
		respondWithError(w, http.StatusBadRequest, "provided email improper", nil)
		return
	}

	databaseUserEntry, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating user in database", err)
		return
	}

	mainUser := User{
		ID:        databaseUserEntry.ID,
		CreatedAt: databaseUserEntry.CreatedAt,
		UpdatedAt: databaseUserEntry.UpdatedAt,
		Email:     databaseUserEntry.Email,
	}

	respondWithJSON(w, 201, mainUser)
}
