package main

import (
	"strings"
)

func removeProfanity(post string) string {
	words := strings.Split(post, " ")
	for i, word := range words {
		checkWord := strings.ToLower(word)
		if checkWord == "kerfuffle" || checkWord == "sharbert" || checkWord == "fornax" {
			words[i] = "****"
		}
	}
	cleanPost := strings.Join(words, " ")
	return cleanPost
}
