package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestUserRepository_Contract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewUserRepositorySQLC(fx.Runtime)
		runUserContract(t, repo)
	})
}

func runUserContract(t *testing.T, repo port.UserRepository) {
	t.Helper()
	ctx := context.Background()
	tag := fmt.Sprintf("%d", time.Now().UnixNano())

	uid := func(suffix string) string { return "01USR" + suffix + tag }

	t.Run("Create_and_FindByID", func(t *testing.T) {
		u := &domain.User{
			Base:           domain.Base{ID: uid("CREATE")},
			Email:          "create-" + tag + "@example.invalid",
			Name:           "Create Test",
			HashedPassword: "hash",
		}
		created, err := repo.Create(ctx, u)
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, u.ID, created.ID)

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, u.Email, got.Email)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "non-existent-"+tag)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("FindByEmail", func(t *testing.T) {
		email := "by-email-" + tag + "@example.invalid"
		u := &domain.User{
			Base:           domain.Base{ID: uid("EMAIL")},
			Email:          email,
			HashedPassword: "h",
		}
		_, err := repo.Create(ctx, u)
		require.NoError(t, err)

		got, err := repo.FindByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, u.ID, got.ID)

		_, err = repo.FindByEmail(ctx, "nope-"+tag+"@example.invalid")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		u := &domain.User{
			Base:           domain.Base{ID: uid("UPD")},
			Email:          "upd-" + tag + "@example.invalid",
			Name:           "Initial",
			HashedPassword: "h",
		}
		_, err := repo.Create(ctx, u)
		require.NoError(t, err)

		u.Name = "Updated"
		require.NoError(t, repo.Update(ctx, u))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", got.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		u := &domain.User{
			Base:           domain.Base{ID: uid("DEL")},
			Email:          "del-" + tag + "@example.invalid",
			HashedPassword: "h",
		}
		_, err := repo.Create(ctx, u)
		require.NoError(t, err)
		require.NoError(t, repo.Delete(ctx, u.ID))

		_, err = repo.FindByID(ctx, u.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("UpdatePassword", func(t *testing.T) {
		u := &domain.User{
			Base:                domain.Base{ID: uid("PWD")},
			Email:               "pwd-" + tag + "@example.invalid",
			HashedPassword:      "old-hash",
			PasswordInitialized: false,
			ForcePasswordChange: true,
		}
		_, err := repo.Create(ctx, u)
		require.NoError(t, err)

		require.NoError(t, repo.UpdatePassword(ctx, u.ID, "new-hash"))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, "new-hash", got.HashedPassword)
		assert.True(t, got.PasswordInitialized)
		assert.False(t, got.ForcePasswordChange)
	})

	t.Run("UpdateLastLogin_SQL_native_timestamp", func(t *testing.T) {
		u := &domain.User{
			Base:           domain.Base{ID: uid("LL")},
			Email:          "ll-" + tag + "@example.invalid",
			HashedPassword: "h",
		}
		_, err := repo.Create(ctx, u)
		require.NoError(t, err)

		// Pre-state: LastLoginAt is nil.
		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Nil(t, got.LastLoginAt)

		require.NoError(t, repo.UpdateLastLogin(ctx, u.ID))

		got, err = repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.NotNil(t, got.LastLoginAt, "last_login_at must be set by CURRENT_TIMESTAMP")
		assert.WithinDuration(t, time.Now(), *got.LastLoginAt, 5*time.Second)
	})

	t.Run("UpdateTwoFactorSecret", func(t *testing.T) {
		u := &domain.User{
			Base:           domain.Base{ID: uid("2FA")},
			Email:          "2fa-" + tag + "@example.invalid",
			HashedPassword: "h",
		}
		_, err := repo.Create(ctx, u)
		require.NoError(t, err)

		require.NoError(t, repo.UpdateTwoFactorSecret(ctx, u.ID, "TOTP-SECRET", true))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, "TOTP-SECRET", got.TwoFactorSecret)
		assert.True(t, got.TwoFactorEnabled)
	})
}
