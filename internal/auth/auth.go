package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	// hash pasword
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("Error creating hash for provided password: %v", err)
	}
	return hash, nil
}

func CheckPasswordHash(password string, hash string) (bool, error) {
	// compare password with hash in db

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("Error comparing provided pasword and hash: %v", err)
	}
	return match, nil
}
