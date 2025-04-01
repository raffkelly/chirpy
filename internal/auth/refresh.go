package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

func MakeRefreshToken() (string, error) {
	byteSlice := make([]byte, 32)
	_, err := rand.Read(byteSlice)
	if err != nil {
		return "", err
	}
	tokenString := hex.EncodeToString(byteSlice)
	return tokenString, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization key")
	}
	token := strings.Fields(authHeader)[1]
	return token, nil
}
