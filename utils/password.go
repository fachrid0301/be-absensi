package utils

import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

// HashPassword enkripsi password dengan bcrypt.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ComparePassword cocokkan password plain dengan hash di DB.
func ComparePassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
