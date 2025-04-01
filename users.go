package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raffkelly/chirpy/internal/auth"
	"github.com/raffkelly/chirpy/internal/database"
)

type User struct {
	ID            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Email         string    `json:"email"`
	Token         string    `json:"token"`
	Refresh_Token string    `json:"refresh_token"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	hashedPW, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error hasing password", err)
		return
	}

	userParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPW,
	}
	databaseUserEntry, err := cfg.dbQueries.CreateUser(r.Context(), userParams)

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

	respondWithJSON(w, http.StatusCreated, mainUser)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "login parameter decoding failed", err)
		return
	}

	user, err := cfg.dbQueries.GetUserFromEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to find user", err)
		return
	}

	err = auth.CheckPasswordHash(user.HashedPassword, params.Password)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password", err)
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating token", err)
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating refresh token", err)
	}
	refTokenParams := database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	}
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), refTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating refresh token db entry", err)
	}

	returnedUser := User{
		ID:            user.ID,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		Email:         user.Email,
		Token:         token,
		Refresh_Token: refreshToken,
	}
	respondWithJSON(w, 200, returnedUser)
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "refresh token not found", err)
		return
	}
	token, err := cfg.dbQueries.GetRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 401, "error getting token from db", err)
		return
	}
	if time.Now().After(token.ExpiresAt) || token.RevokedAt.Valid {
		respondWithError(w, 401, "token has been revoked or is expired", err)
		return
	}
	newAccessToken, err := auth.MakeJWT(token.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating access token", err)
		return
	}

	responseToken := make(map[string]string)
	responseToken["token"] = newAccessToken
	respondWithJSON(w, 200, responseToken)
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "refresh token not found", err)
		return
	}
	err = cfg.dbQueries.RevokeToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 401, "error revoking  token", err)
		return
	}
	respondWithJSON(w, 204, nil)
}
