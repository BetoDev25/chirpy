package auth

import (
	"strings"
	"net/http"
	"errors"
)

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func GetBearerToken(headers http.Header) (string, error) {
	tokenString := headers.Get("Authorization")
	if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
		return "", ErrNoAuthHeaderIncluded
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer")
	return strings.TrimSpace(tokenString), nil
}
