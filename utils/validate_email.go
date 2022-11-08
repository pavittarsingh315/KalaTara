package utils

import "net/mail"

// Returns true if email is valid, false if not
func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
