package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/pquerna/otp/totp"
)

// Two-factor backup codes: shown once at setup, accepted in place of a TOTP
// code at sign-in and at re-authentication, each usable exactly once.
//
// Format: 12 characters from the Crockford base32 alphabet, shown as
// "xxxx-xxxx-xxxx" (60 bits of entropy). The alphabet has no I, L, O or U and
// is case-insensitive, so normalisation can fold case, drop dashes and spaces,
// and read the common look-alikes (o -> 0, i/l -> 1) without ambiguity.
//
// Storage: the users.two_factor_backup_codes column holds a JSON array of the
// hex SHA-256 digests of the normalised codes -- never the codes themselves.
// A plain hash rather than bcrypt is deliberate: each code is 60 bits drawn
// from crypto/rand, which already defeats offline guessing, unlike a
// human-chosen password that a slow hash has to protect. Bcrypt would also
// cost a full slow-hash per stored code on every sign-in attempt. And anyone
// able to read this column reads the TOTP secret from the same row in clear,
// so a slower hash here would protect nothing that is not already exposed.
// Matching uses a constant-time comparison over every stored digest.
const (
	backupCodeAlphabet = "0123456789abcdefghjkmnpqrstvwxyz"
	backupCodeLength   = 12
	backupCodeGroup    = 4
)

// generateBackupCodes returns count fresh codes in display form.
func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := range count {
		codes[i] = generateBackupCode()
	}
	return codes
}

func generateBackupCode() string {
	raw := make([]byte, backupCodeLength)
	if _, err := rand.Read(raw); err != nil {
		// crypto/rand never fails on supported platforms; refuse to hand out a
		// predictable code if it somehow does.
		panic(fmt.Sprintf("backup code: crypto/rand failed: %v", err))
	}
	var b strings.Builder
	for i, v := range raw {
		if i > 0 && i%backupCodeGroup == 0 {
			b.WriteByte('-')
		}
		// 256 is a multiple of 32, so the modulo is unbiased.
		b.WriteByte(backupCodeAlphabet[int(v)%len(backupCodeAlphabet)])
	}
	return b.String()
}

// normalizeBackupCode folds a typed code into its canonical form. It returns
// "" when the input cannot be a backup code (wrong length or alphabet), so a
// 6-digit TOTP code or junk is never hashed and compared.
func normalizeBackupCode(in string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(in) {
		switch r {
		case '-', ' ', '\t':
			continue
		case 'o':
			r = '0'
		case 'i', 'l':
			r = '1'
		}
		if !strings.ContainsRune(backupCodeAlphabet, r) {
			return ""
		}
		b.WriteRune(r)
	}
	if b.Len() != backupCodeLength {
		return ""
	}
	return b.String()
}

func hashBackupCode(normalized string) string {
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

// encodeBackupCodeHashes turns display-form codes into the stored bytes.
func encodeBackupCodeHashes(codes []string) ([]byte, error) {
	hashes := make([]string, 0, len(codes))
	for _, c := range codes {
		n := normalizeBackupCode(c)
		if n == "" {
			return nil, fmt.Errorf("backup code: generated code is not well-formed")
		}
		hashes = append(hashes, hashBackupCode(n))
	}
	return json.Marshal(hashes)
}

// storeNewBackupCodes generates a fresh set, persists its hashes (replacing
// any previous set) and returns the codes to show the user once.
func storeNewBackupCodes(ctx context.Context, users port.UserRepository, userID string, count int) ([]string, error) {
	codes := generateBackupCodes(count)
	stored, err := encodeBackupCodeHashes(codes)
	if err != nil {
		return nil, err
	}
	if err := users.UpdateTwoFactorBackupCodes(ctx, userID, stored); err != nil {
		return nil, fmt.Errorf("persist backup codes: %w", err)
	}
	return codes, nil
}

// consumeBackupCode reports whether code is one of user's unused backup codes
// and, when it is, removes it. The removal is a compare-and-swap against the
// set read with the user, so two concurrent attempts with the same code cannot
// both succeed: the loser's swap matches no row and it is rejected. A storage
// error is returned as such; a non-matching code is (false, nil).
func (s *AuthService) consumeBackupCode(ctx context.Context, user *domain.User, code string) (bool, error) {
	normalized := normalizeBackupCode(code)
	if normalized == "" || len(user.TwoFactorBackupCodes) == 0 {
		return false, nil
	}
	var hashes []string
	if err := json.Unmarshal(user.TwoFactorBackupCodes, &hashes); err != nil {
		// Unreadable legacy content: treat as no codes rather than failing sign-in.
		return false, nil
	}
	want := []byte(hashBackupCode(normalized))
	match := -1
	for i, h := range hashes {
		if subtle.ConstantTimeCompare([]byte(h), want) == 1 && match < 0 {
			match = i
		}
	}
	if match < 0 {
		return false, nil
	}
	remaining := make([]string, 0, len(hashes)-1)
	remaining = append(remaining, hashes[:match]...)
	remaining = append(remaining, hashes[match+1:]...)
	var next []byte
	if len(remaining) > 0 {
		var err error
		if next, err = json.Marshal(remaining); err != nil {
			return false, err
		}
	}
	swapped, err := s.userRepo.SwapTwoFactorBackupCodes(ctx, user.ID, user.TwoFactorBackupCodes, next)
	if err != nil {
		return false, fmt.Errorf("consume backup code: %w", err)
	}
	return swapped, nil
}

// checkSecondFactor accepts a current TOTP code or, failing that, an unused
// backup code (which it consumes).
func (s *AuthService) checkSecondFactor(ctx context.Context, user *domain.User, code string) (bool, error) {
	if totp.Validate(strings.TrimSpace(code), user.TwoFactorSecret) {
		return true, nil
	}
	return s.consumeBackupCode(ctx, user, code)
}
