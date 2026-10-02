package service_test

import (
	"context"
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

const reauthSecret = "JBSWY3DPEHPK3PXP"

func reauthUser(t *testing.T, twoFactor bool) (*service.AuthService, string) {
	t.Helper()
	users := fake.NewUserRepository()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-horse"), bcrypt.MinCost)
	require.NoError(t, err)
	u := &domain.User{Base: domain.Base{ID: "user-reauth"}, Email: "jane@example.com", HashedPassword: string(hash), PasswordInitialized: true}
	if twoFactor {
		u.TwoFactorEnabled = true
		u.TwoFactorSecret = reauthSecret
	}
	created, err := users.Create(context.Background(), u)
	require.NoError(t, err)
	return service.NewAuthService(users, service.NewJWTManager("test-secret-key-at-least-32-bytes-long", "ogoune", time.Hour)), created.ID
}

// Spec 094 FR-009a: the export re-checks exactly what sign-in checks, and
// every failure looks the same.
func TestReauthenticate(t *testing.T) {
	ctx := context.Background()
	code := func() string {
		c, err := totp.GenerateCode(reauthSecret, time.Now())
		require.NoError(t, err)
		return c
	}

	t.Run("two-factor off: password alone, code ignored", func(t *testing.T) {
		svc, id := reauthUser(t, false)
		assert.NoError(t, svc.Reauthenticate(ctx, id, "correct-horse", ""))
		assert.NoError(t, svc.Reauthenticate(ctx, id, "correct-horse", "whatever"))
	})

	t.Run("two-factor on: password and a valid code", func(t *testing.T) {
		svc, id := reauthUser(t, true)
		assert.NoError(t, svc.Reauthenticate(ctx, id, "correct-horse", code()))
		assert.NoError(t, svc.Reauthenticate(ctx, id, "correct-horse", " "+code()+" "), "code is trimmed")
	})

	failures := map[string]struct {
		twoFactor      bool
		password, code string
	}{
		"missing password":                                {false, "", ""},
		"wrong password":                                  {false, "wrong", ""},
		"two-factor on, missing code":                     {true, "correct-horse", ""},
		"two-factor on, wrong code":                       {true, "correct-horse", "000000"},
		"two-factor on, wrong password with a valid code": {true, "wrong", "VALID"},
	}
	for name, tc := range failures {
		t.Run(name, func(t *testing.T) {
			svc, id := reauthUser(t, tc.twoFactor)
			c := tc.code
			if c == "VALID" {
				c = code()
			}
			err := svc.Reauthenticate(ctx, id, tc.password, c)
			assert.ErrorIs(t, err, service.ErrInvalidCredentials, "one error for every failure")
		})
	}

	t.Run("unknown user is the same error", func(t *testing.T) {
		svc, _ := reauthUser(t, false)
		assert.ErrorIs(t, svc.Reauthenticate(ctx, "nobody", "correct-horse", ""), service.ErrInvalidCredentials)
	})
}
