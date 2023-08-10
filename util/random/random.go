package random

import "crypto/rand"

const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// String [0-9a-zA-Z]*
func String(length int) string {
	if length <= 0 {
		return ""
	}

	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return ""
	}

	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}

	return string(bytes)
}
