package security

import (
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

const randomPasswordAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789!@#$%"
const randomPasswordLength = 12

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
}

func CheckPassword(hash []byte, password string) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
}

// GenerateRandomPassword returns a new random password, long enough to
// satisfy the app's minimum length rule, meant to be shown once (e.g. after
// a staff-triggered reset) and never stored anywhere in plaintext.
func GenerateRandomPassword() (string, error) {
	buf := make([]byte, randomPasswordLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, randomPasswordLength)
	for i, b := range buf {
		out[i] = randomPasswordAlphabet[int(b)%len(randomPasswordAlphabet)]
	}
	return string(out), nil
}
