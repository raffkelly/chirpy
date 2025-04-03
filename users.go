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
	Is_Chirpy_Red bool      `json:"is_chirpy_red"`
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
		ID:            databaseUserEntry.ID,
		CreatedAt:     databaseUserEntry.CreatedAt,
		UpdatedAt:     databaseUserEntry.UpdatedAt,
		Email:         databaseUserEntry.Email,
		Is_Chirpy_Red: databaseUserEntry.IsChirpyRed,
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
		Is_Chirpy_Red: user.IsChirpyRed,
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

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "token missing", err)
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.secret)

	if err != nil {
		respondWithError(w, 401, "invalid token", err)
		return
	}
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
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

	userParams := database.UpdateUserEmailPasswordParams{
		Email:          params.Email,
		HashedPassword: hashedPW,
		ID:             userID,
	}
	updatedUser, err := cfg.dbQueries.UpdateUserEmailPassword(r.Context(), userParams)
	returnedUser := User{
		ID:            userID,
		CreatedAt:     updatedUser.CreatedAt,
		UpdatedAt:     updatedUser.UpdatedAt,
		Email:         updatedUser.Email,
		Is_Chirpy_Red: updatedUser.IsChirpyRed,
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error updating email and password in db", err)
		return
	}
	respondWithJSON(w, 200, returnedUser)

}

func (cfg *apiConfig) handleUpgradeUser(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polka_key {
		respondWithError(w, 401, "bad api key", err)
		return
	}
	type parameters struct {
		Event string            `json:"event"`
		Data  map[string]string `json:"data"`
	}
	var params parameters
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error decoding request body", err)
		return
	}
	if params.Event != "user.upgraded" {
		respondWithJSON(w, 204, nil)
		return
	} else if params.Event == "user.upgraded" {
		userID, err := uuid.Parse(string(params.Data["user_id"]))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "unable to parse user id from request", err)
			return
		}
		err = cfg.dbQueries.UpgradeUser(r.Context(), userID)
		if err != nil {
			respondWithError(w, 404, "user not found", err)
			return
		}
		respondWithJSON(w, 204, nil)
	}
}
