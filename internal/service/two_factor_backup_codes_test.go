package service_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
)

// enrolTwoFactor runs the v1 setup + verify flow for a password-protected user
// and returns the backup codes shown to them.
func enrolTwoFactor(t *testing.T) (*service.AuthService, *service.TwoFactorService, *fake.UserRepository, *domain.User, []string) {
	t.Helper()
	ctx := context.Background()
	authSvc, tfSvc, users, _ := newTwoFactorTestServices(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-horse"), bcrypt.MinCost)
	require.NoError(t, err)
	u, err := users.Create(ctx, &domain.User{
		Base:                domain.Base{ID: "user-backup"},
		Email:               "backup@example.com",
		HashedPassword:      string(hash),
		PasswordInitialized: true,
	})
	require.NoError(t, err)

	setup, err := tfSvc.Setup(ctx, u.ID)
	require.NoError(t, err)
	code, err := totp.GenerateCode(setup.Secret, time.Now())
	require.NoError(t, err)
	codes, err := tfSvc.Verify(ctx, u.ID, code)
	require.NoError(t, err)
	require.Len(t, codes, 10)

	u, err = users.FindByID(ctx, u.ID)
	require.NoError(t, err)
	return authSvc, tfSvc, users, u, codes
}

func storedHashes(t *testing.T, users *fake.UserRepository, id string) []string {
	t.Helper()
	u, err := users.FindByID(context.Background(), id)
	require.NoError(t, err)
	if len(u.TwoFactorBackupCodes) == 0 {
		return nil
	}
	var hashes []string
	require.NoError(t, json.Unmarshal(u.TwoFactorBackupCodes, &hashes))
	return hashes
}

func TestBackupCodes_VerifyPersistsHashesOnly(t *testing.T) {
	_, _, users, u, codes := enrolTwoFactor(t)

	hashes := storedHashes(t, users, u.ID)
	require.Len(t, hashes, len(codes))
	seen := map[string]bool{}
	for _, c := range codes {
		assert.Regexp(t, `^[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}$`, c)
		assert.False(t, seen[c], "codes are distinct")
		seen[c] = true
		bare := strings.ReplaceAll(c, "-", "")
		assert.NotContains(t, string(u.TwoFactorBackupCodes), bare, "no cleartext code in the stored bytes")
		assert.NotContains(t, string(u.TwoFactorBackupCodes), c, "no cleartext code in the stored bytes")
	}
	for _, h := range hashes {
		assert.Len(t, h, 64, "hex SHA-256 digest")
	}
}

func TestBackupCodes_SignInConsumesCodeOnce(t *testing.T) {
	ctx := context.Background()
	authSvc, _, users, u, codes := enrolTwoFactor(t)

	token, err := authSvc.Verify2FA(ctx, u.Email, codes[3])
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, storedHashes(t, users, u.ID), len(codes)-1, "used code removed")

	_, err = authSvc.Verify2FA(ctx, u.Email, codes[3])
	assert.Error(t, err, "a used backup code must not work twice")

	// Normalisation: case, spaces and missing dashes are tolerated.
	typed := " " + strings.ToUpper(strings.ReplaceAll(codes[4], "-", " ")) + " "
	_, err = authSvc.Verify2FA(ctx, u.Email, typed)
	require.NoError(t, err)
	assert.Len(t, storedHashes(t, users, u.ID), len(codes)-2)

	// TOTP still works and consumes nothing.
	reloaded, _ := users.FindByID(ctx, u.ID)
	otp, err := totp.GenerateCode(reloaded.TwoFactorSecret, time.Now())
	require.NoError(t, err)
	_, err = authSvc.Verify2FA(ctx, u.Email, otp)
	require.NoError(t, err)
	assert.Len(t, storedHashes(t, users, u.ID), len(codes)-2)

	// Wrong codes are rejected without touching the set.
	for _, bad := range []string{"", "000000", "abcd-efgh-jkmn", "not a code at all"} {
		_, err = authSvc.Verify2FA(ctx, u.Email, bad)
		assert.Error(t, err, "input %q", bad)
	}
	assert.Len(t, storedHashes(t, users, u.ID), len(codes)-2)
}

func TestBackupCodes_ConcurrentUseSucceedsOnce(t *testing.T) {
	ctx := context.Background()
	authSvc, _, _, u, codes := enrolTwoFactor(t)

	var ok atomic.Int32
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := authSvc.Verify2FA(ctx, u.Email, codes[0]); err == nil {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(1), ok.Load())
}

func TestBackupCodes_RegenerateReplacesPreviousSet(t *testing.T) {
	ctx := context.Background()
	authSvc, _, users, u, oldCodes := enrolTwoFactor(t)

	// The legacy root setup path also issues (and now stores) a set.
	resp, err := authSvc.GenerateTOTPSecret(ctx, u.ID, u.Email)
	require.NoError(t, err)
	require.Len(t, resp.BackupCodes, 10)
	assert.Len(t, storedHashes(t, users, u.ID), 10)

	_, err = authSvc.Verify2FA(ctx, u.Email, oldCodes[0])
	assert.Error(t, err, "codes from the replaced set no longer work")
	_, err = authSvc.Verify2FA(ctx, u.Email, resp.BackupCodes[0])
	assert.NoError(t, err)
}

func TestBackupCodes_DisableClearsCodes(t *testing.T) {
	ctx := context.Background()

	t.Run("v1 disable", func(t *testing.T) {
		_, tfSvc, users, u, _ := enrolTwoFactor(t)
		otp, err := totp.GenerateCode(u.TwoFactorSecret, time.Now())
		require.NoError(t, err)
		require.NoError(t, tfSvc.Disable(ctx, u.ID, otp))
		assert.Empty(t, storedHashes(t, users, u.ID))
	})

	t.Run("legacy disable", func(t *testing.T) {
		authSvc, _, users, u, _ := enrolTwoFactor(t)
		require.NoError(t, authSvc.Disable2FA(ctx, u.ID, "correct-horse"))
		assert.Empty(t, storedHashes(t, users, u.ID))
	})
}

func TestBackupCodes_Reauthenticate(t *testing.T) {
	ctx := context.Background()
	authSvc, _, users, u, codes := enrolTwoFactor(t)

	require.NoError(t, authSvc.Reauthenticate(ctx, u.ID, "correct-horse", codes[0]))
	assert.Len(t, storedHashes(t, users, u.ID), len(codes)-1, "the code is consumed")
	assert.ErrorIs(t, authSvc.Reauthenticate(ctx, u.ID, "correct-horse", codes[0]), service.ErrInvalidCredentials, "second use fails")

	// A valid backup code does not excuse a wrong password, and is not spent by it.
	assert.ErrorIs(t, authSvc.Reauthenticate(ctx, u.ID, "wrong", codes[1]), service.ErrInvalidCredentials)
	assert.Len(t, storedHashes(t, users, u.ID), len(codes)-1)

	otp, err := totp.GenerateCode(u.TwoFactorSecret, time.Now())
	require.NoError(t, err)
	assert.NoError(t, authSvc.Reauthenticate(ctx, u.ID, "correct-horse", otp), "TOTP still works")

	// A code used to re-authenticate is gone for sign-in too.
	_, err = authSvc.Verify2FA(ctx, u.Email, codes[0])
	assert.Error(t, err)
}
