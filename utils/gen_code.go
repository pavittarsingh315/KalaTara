package utils

import (
	"crypto/rand"
	"io"
	"log"
)

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func GenerateRandomCode(maxLength int) string {
	b := make([]byte, maxLength)
	n, err := io.ReadAtLeast(rand.Reader, b, maxLength)

	if n != maxLength {
		log.Fatal(err)
	}

	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}

	return string(b)
}
