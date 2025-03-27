package main

import (
	"encoding/json"
	"net/http"
)

func handlerValidate_Chirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Error        string `json:"error"`
		Valid        bool   `json:"valid"`
		Cleaned_body string `json:"cleaned_body"`
	}

	params := parameters{}

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

	cleanPost := removeProfanity(params.Body)

	respondWithJSON(w, http.StatusOK, returnVals{Cleaned_body: cleanPost})
}
