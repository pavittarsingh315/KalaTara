package utils

import (
	"net/mail"

	"github.com/nyaruka/phonenumbers"
)

// Returns true if email is valid, false if not
func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Returns true if phone number is valid, false if not
func ValidatePhone(number string) bool {
	parsedNumber, err := phonenumbers.Parse(number, "")
	if err != nil {
		return false
	}
	return phonenumbers.IsValidNumber(parsedNumber)
}
