package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password string, hashString string) (bool, error) {
	match, _, err := argon2id.CheckHash(password, hashString)
	return match, err
}

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()

	claims := jwt.RegisteredClaims{}
	claims.Issuer = "chirpy-access"
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(expiresIn))
	claims.Subject = userId.String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	keyFunc := func(t *jwt.Token) (any, error) { return []byte(tokenSecret), nil }
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return uuid.Nil, err
	}

	idStr, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(idStr)
}

func GetBearerToken(headers http.Header) (string, error) {
	authStr := headers.Get("Authorization")
	if len(authStr) == 0 {
		return "", fmt.Errorf("no Authorization header")
	}

	tokStr, found := strings.CutPrefix(authStr, "Bearer ")
	if !found {
		return "", fmt.Errorf("no bearer prefix")
	}

	return strings.TrimSpace(tokStr), nil
}

func AuthenticateByJWT(headers http.Header, tokenSecret string) (uuid.UUID, error) {
	token, err := GetBearerToken(headers)
	if err != nil {
		return uuid.Nil, err
	}

	authId, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		return uuid.Nil, err
	}

	return authId, nil
}
