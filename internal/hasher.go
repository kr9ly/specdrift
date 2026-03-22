package internal

import (
	"crypto/sha256"
	"fmt"
	"os"
)

// HashFile computes the SHA-256 hash of a file and returns the first 8 hex characters.
func HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)[:8], nil
}
