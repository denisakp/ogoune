package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
)

// POST /auth/initialize-password has no authentication in front of it. It exists
// for exactly one step: an account that has never set a password signs in with
// the well-known default and then chooses a real one.
//
// Without a guard it set ANY known account's password and returned a session
// token for it — an account takeover from an email address alone, against an
// endpoint anyone can reach. These tests pin both halves: the legitimate step
// still works, and everything else is refused indistinguishably.
func newAuthFixture(t *testing.T) (*service.AuthService, *fake.UserRepository) {
	t.Helper()
	users := fake.NewUserRepository()
	svc := service.NewAuthService(users, service.NewJWTManager("test-secret-key-at-least-32-bytes-long", "ogoune", time.Hour))
	return svc, users
}

func seedAuthUser(t *testing.T, users *fake.UserRepository, email string, initialized bool) *domain.User {
	t.Helper()
	u, err := users.Create(context.Background(), &domain.User{
		Email:               email,
		Name:                "Test",
		HashedPassword:      "$2a$12$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ012",
		PasswordInitialized: initialized,
	})
	require.NoError(t, err)
	return u
}

func TestInitializePassword_AllowedOnceForAnUninitializedAccount(t *testing.T) {
	svc, users := newAuthFixture(t)
	seedAuthUser(t, users, "new@ogoune.test", false)

	got, err := svc.InitializePassword(context.Background(), "new@ogoune.test", "a-real-password")
	require.NoError(t, err, "the first-login flow must keep working")
	require.NotNil(t, got)
	assert.True(t, got.PasswordInitialized, "and the flag must flip, closing the door behind it")
}

func TestInitializePassword_RefusedForAnInitializedAccount(t *testing.T) {
	svc, users := newAuthFixture(t)
	seedAuthUser(t, users, "admin@ogoune.test", true)

	_, err := svc.InitializePassword(context.Background(), "admin@ogoune.test", "attacker-chosen")
	require.ErrorIs(t, err, service.ErrPasswordInitializationRefused,
		"an account with a password must not be resettable by anyone who knows its address")
}

// The second call must fail even though the first succeeded: otherwise the
// window stays open for every account that has ever used the flow.
func TestInitializePassword_IsNotRepeatable(t *testing.T) {
	svc, users := newAuthFixture(t)
	seedAuthUser(t, users, "once@ogoune.test", false)

	_, err := svc.InitializePassword(context.Background(), "once@ogoune.test", "first-password")
	require.NoError(t, err)

	_, err = svc.InitializePassword(context.Background(), "once@ogoune.test", "second-password")
	require.ErrorIs(t, err, service.ErrPasswordInitializationRefused)
}

// An unauthenticated endpoint that answers differently for known and unknown
// addresses is an oracle for which emails have accounts.
func TestInitializePassword_UnknownAccountIsIndistinguishable(t *testing.T) {
	svc, users := newAuthFixture(t)
	seedAuthUser(t, users, "known@ogoune.test", true)

	_, errKnown := svc.InitializePassword(context.Background(), "known@ogoune.test", "attacker-chosen")
	_, errUnknown := svc.InitializePassword(context.Background(), "nobody@ogoune.test", "attacker-chosen")

	require.ErrorIs(t, errKnown, service.ErrPasswordInitializationRefused)
	require.ErrorIs(t, errUnknown, service.ErrPasswordInitializationRefused)
	assert.Equal(t, errKnown.Error(), errUnknown.Error(),
		"the two must be the same error, not merely both errors")
}

func TestInitializePassword_StillRejectsAWeakPassword(t *testing.T) {
	svc, users := newAuthFixture(t)
	seedAuthUser(t, users, "weak@ogoune.test", false)

	_, err := svc.InitializePassword(context.Background(), "weak@ogoune.test", "short")
	require.ErrorIs(t, err, service.ErrInvalidPassword)
}
