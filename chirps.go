package main

import (
	"encoding/json"
	"net/http"
	"sort"
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
	var err error
	var userID uuid.UUID
	var returnChirps []database.Chirp
	sortMethod := r.URL.Query().Get("sort")
	err = nil
	s := r.URL.Query().Get("author_id")
	if s != "" {
		userID, err = uuid.Parse(s)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "unable to get uuid from query", err)
			return
		}
		returnChirps, err = cfg.dbQueries.GetChripsByUserID(r.Context(), userID)
	} else {
		returnChirps, err = cfg.dbQueries.GetChirps(r.Context())
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error retrieving chrips from db", err)
		return
	}
	responseChirps := make([]Chirp, len(returnChirps))
	for i, singleChirp := range returnChirps {
		responseChirps[i] = Chirp{
			ID:        singleChirp.ID,
			CreatedAt: singleChirp.CreatedAt,
			UpdatedAt: singleChirp.UpdatedAt,
			Body:      singleChirp.Body,
			UserID:    singleChirp.UserID,
		}
	}
	if sortMethod == "desc" {
		sort.Slice(responseChirps, func(i, j int) bool { return responseChirps[i].CreatedAt.After(responseChirps[j].CreatedAt) })
	} else {
		sort.Slice(responseChirps, func(i, j int) bool { return responseChirps[i].CreatedAt.Before(responseChirps[j].CreatedAt) })
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

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "token missing", err)
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.secret)

	if err != nil {
		respondWithError(w, 403, "invalid token", err)
		return
	}
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error parsing chirpID from request", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if chirp.UserID != userID {
		respondWithError(w, 403, "user not authorized to delete chirp", err)
		return
	}
	err = cfg.dbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "chirp not found", err)
		return
	}
	respondWithJSON(w, 204, nil)
}
