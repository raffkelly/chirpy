package auth

import (
	"net/http"
)

func GetAPIKey(headers http.Header) (string, error) {
	token, err := GetBearerToken(headers)
	if err != nil {
		return "", err
	}
	return token, nil
}
