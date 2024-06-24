package auth

import (
	"errors"
	"net/http"
	"strings"
)

// GetApiKey extracts an API key from the headers
// Example:
// Authorization: ApiKey {hex string}
func GetApiKey(headers http.Header) (string, error) {
	authorization := headers.Get("Authorization")
	if authorization == "" {
		return "", errors.New("no authorization header")
	}
	parts := strings.Split(authorization, " ")
	if len(parts) != 2 {
		return "", errors.New("invalid authorization header")
	}
	if parts[0] != "ApiKey" {
		return "", errors.New("malformed first part of auth header")
	}
	return parts[1], nil
}
