package utils

import (
	"crypto/rand"
	"io"
	"log"
	"strconv"
)

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func EncodeToInt(maxLength int) int {
	b := make([]byte, maxLength)
	n, err := io.ReadAtLeast(rand.Reader, b, maxLength)

	if n != maxLength {
		log.Fatal(err)
	}

	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}

	code, _ := strconv.Atoi(string(b))

	// code might potentially be say 058798 ie 58798 ie 5 digits. So add 111,222 to make it 6 digits if thats the case
	if code < 100000 {
		code += 111222
	}

	return code
}
