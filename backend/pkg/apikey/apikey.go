package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	prefix           = "pk_live_"
	entropyByteCount = 16
	publicPrefixLen  = 12
)

// Generate creates a new raw API key and returns the key, hash, and public prefix.
func Generate() (string, string, string, error) {
	buf := make([]byte, entropyByteCount)
	if _, err := rand.Read(buf); err != nil {
		return "", "", "", fmt.Errorf("generate api key: %w", err)
	}

	rawKey := prefix + hex.EncodeToString(buf)
	return rawKey, Hash(rawKey), ExtractPrefix(rawKey), nil
}

// Hash returns the SHA-256 hex digest of an API key.
func Hash(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

// ExtractPrefix returns the non-sensitive key prefix used in logs and UI.
func ExtractPrefix(rawKey string) string {
	if len(rawKey) <= publicPrefixLen {
		return rawKey
	}
	return rawKey[:publicPrefixLen]
}

// IsAPIKeyFormat returns true when the token appears to be an API key.
func IsAPIKeyFormat(token string) bool {
	return len(token) > len(prefix) && token[:len(prefix)] == prefix
}
