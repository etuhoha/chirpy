package auth

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password string, hashString string) (bool, error) {
	match, _, err := argon2id.CheckHash(password, hashString)
	return match, err
}
