package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/raffkelly/chirpy/internal/auth"
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
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no token found for user", err)
	}
	userIDfromJWT, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token", err)
		return
	}
	params := Chirp{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
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
		UserID: userIDfromJWT,
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

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, "error parsing chirp id", err)
		return
	}

	intermedChirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "unable to find chirp", err)
		return
	}
	returnedChirp := Chirp{
		ID:        intermedChirp.ID,
		CreatedAt: intermedChirp.CreatedAt,
		UpdatedAt: intermedChirp.UpdatedAt,
		Body:      intermedChirp.Body,
		UserID:    intermedChirp.UserID,
	}
	respondWithJSON(w, 200, returnedChirp)
}
