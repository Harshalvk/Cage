package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateAPIKey() (rawKey string, keyHash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random key: %w", err)
	}

	rawKey = "cage_" + hex.EncodeToString(bytes)
	keyHash = HashKey(rawKey)
	return rawKey, keyHash, nil
}

func HashKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}
