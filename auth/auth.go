package auth

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashAndSalt hash and salt the password
func HashAndSalt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), err
}

// ComparePasswords compare the hashed and plain passwords
func ComparePasswords(hashed, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}

func GeneratePwd() (string, error) {
	size := 10
	buf := make([]byte, size)
	size, err := rand.Read(buf)
	if err != nil {
		return "", fmt.Errorf("generate pwd error: %w", err)
	}

	return fmt.Sprintf("%x", sha1.Sum(buf[:size]))[:6], nil
}
