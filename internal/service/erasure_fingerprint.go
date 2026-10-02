package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/pkg/crypto"
)

// erasureFingerprintDomain separates this use of the install key from every
// other one: the same key never yields the same value for two purposes.
const erasureFingerprintDomain = "ogoune-erasure-v1:"

// ErasureFingerprint is the keyed one-way fingerprint of an address kept in
// erasure records (spec 095, FR-012): it lets a later preview recognise an
// address already erased without storing it. Whoever holds the install's
// secret key can test a guess against it -- an accepted, documented limit.
func ErasureFingerprint(email string) (string, error) {
	key, err := crypto.GlobalProvider().GetEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("erasure fingerprint: %w", err)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(erasureFingerprintDomain + domain.NormalizeEmail(email)))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
