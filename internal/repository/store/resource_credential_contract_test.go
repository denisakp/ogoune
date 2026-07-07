package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/pkg/crypto"
)

const credentialContractTestKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func setupCredentialCryptoForTest(t *testing.T) {
	t.Helper()
	t.Setenv("APP_SECRET_KEY", credentialContractTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	t.Cleanup(func() {
		_ = os.Unsetenv("APP_SECRET_KEY")
		crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	})
}

func TestResourceCredentialRepository_Contract(t *testing.T) {
	setupCredentialCryptoForTest(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResources(t, fx, "res-cred-1", "res-cred-2", "res-cred-3", "res-cred-roundtrip")
		repo := store.NewResourceCredentialRepositorySQLC(fx.Runtime)
		runResourceCredentialContract(t, repo)
	})
}

func runResourceCredentialContract(t *testing.T, repo port.ResourceCredentialRepository) {
	t.Helper()
	ctx := context.Background()
	tag := fmt.Sprintf("%d", time.Now().UnixNano())

	t.Run("Get_NotFound", func(t *testing.T) {
		_, err := repo.Get(ctx, "no-such-resource-"+tag)
		assert.ErrorIs(t, err, repository.ErrCredentialNotFound)
	})

	t.Run("Upsert_and_Get_round_trip_plaintext", func(t *testing.T) {
		cred := &domain.ResourceCredential{
			Base:       domain.Base{ID: "01RC001" + tag[len(tag)-8:]},
			ResourceID: "res-cred-1",
			Username:   "admin",
			Password:   []byte("s3cret-pwd"),
			Options:    []byte(`{"db":"postgres"}`),
		}
		require.NoError(t, repo.Upsert(ctx, cred))

		got, err := repo.Get(ctx, "res-cred-1")
		require.NoError(t, err)
		assert.Equal(t, "admin", got.Username)
		assert.Equal(t, []byte("s3cret-pwd"), got.Password, "Password must round-trip as plaintext via the port")
		assert.Equal(t, []byte(`{"db":"postgres"}`), got.Options)
	})

	t.Run("Upsert_idempotent_overwrite", func(t *testing.T) {
		cred := &domain.ResourceCredential{
			Base:       domain.Base{ID: "01RC002" + tag[len(tag)-8:]},
			ResourceID: "res-cred-2",
			Username:   "u1",
			Password:   []byte("p1"),
		}
		require.NoError(t, repo.Upsert(ctx, cred))

		cred2 := &domain.ResourceCredential{
			Base:       domain.Base{ID: "01RC02B" + tag[len(tag)-8:]},
			ResourceID: "res-cred-2",
			Username:   "u2",
			Password:   []byte("p2"),
		}
		require.NoError(t, repo.Upsert(ctx, cred2))

		got, err := repo.Get(ctx, "res-cred-2")
		require.NoError(t, err)
		assert.Equal(t, "u2", got.Username)
		assert.Equal(t, []byte("p2"), got.Password)
	})

	t.Run("Exists", func(t *testing.T) {
		got, err := repo.Exists(ctx, "res-cred-1")
		require.NoError(t, err)
		assert.True(t, got)

		got, err = repo.Exists(ctx, "nonexistent-"+tag)
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("Delete_and_Delete_NotFound", func(t *testing.T) {
		cred := &domain.ResourceCredential{
			Base:       domain.Base{ID: "01RC003" + tag[len(tag)-8:]},
			ResourceID: "res-cred-3",
			Username:   "u",
			Password:   []byte("p"),
		}
		require.NoError(t, repo.Upsert(ctx, cred))
		require.NoError(t, repo.Delete(ctx, "res-cred-3"))

		_, err := repo.Get(ctx, "res-cred-3")
		assert.ErrorIs(t, err, repository.ErrCredentialNotFound)

		err = repo.Delete(ctx, "res-cred-3")
		assert.ErrorIs(t, err, repository.ErrCredentialNotFound)
	})
}
