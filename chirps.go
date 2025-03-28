package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/raffkelly/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {

	params := Chirp{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error decoding", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	params.Body = removeProfanity(params.Body)

	postingParams := database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserID,
	}

	interChirp, err := cfg.dbQueries.CreateChirp(r.Context(), postingParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create chirp in database", err)
		return
	}
	postedChirp := Chirp{
		ID:        interChirp.ID,
		CreatedAt: interChirp.CreatedAt,
		UpdatedAt: interChirp.UpdatedAt,
		Body:      interChirp.Body,
		UserID:    interChirp.UserID,
	}
	respondWithJSON(w, http.StatusCreated, postedChirp)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	returnChirps, err := cfg.dbQueries.GetChirps(r.Context())
	responseChirps := make([]Chirp, len(returnChirps))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error retrieving chirps from db", err)
		return
	}
	for i, singleChirp := range returnChirps {
		responseChirps[i] = Chirp{
			ID:        singleChirp.ID,
			CreatedAt: singleChirp.CreatedAt,
			UpdatedAt: singleChirp.UpdatedAt,
			Body:      singleChirp.Body,
			UserID:    singleChirp.UserID,
		}
	}
	respondWithJSON(w, http.StatusOK, responseChirps)
}
