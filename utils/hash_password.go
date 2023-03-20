package utils

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	passwordBytes := []byte(password)

	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		log.Print(err)
	}

	return string(hashedPassword)
}

func VerifyPassword(hashPassword, plainPassword string) bool {
	plainPasswordBytes := []byte(plainPassword)
	hashPasswordBytes := []byte(hashPassword)

	// error == nil if passwords match
	err := bcrypt.CompareHashAndPassword(hashPasswordBytes, plainPasswordBytes)

	if err == nil {
		return true
	} else {
		return false
	}
}
