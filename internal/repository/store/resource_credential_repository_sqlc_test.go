package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/database"
	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestResourceCredentialRepository_SqlcContract(t *testing.T) {
	setupCredentialCryptoForTest(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResources(t, fx, "res-cred-1", "res-cred-2", "res-cred-3", "res-cred-roundtrip")
		repo := store.NewResourceCredentialRepositorySQLC(fx.Runtime)
		runResourceCredentialContract(t, repo)
	})
}

// TestResourceCredentialRepository_SqlcEncryption_RoundTrip is the SC-006
// gate for resource_credential: on-disk Password + Options carry ciphertext;
// port returns plaintext.
func TestResourceCredentialRepository_SqlcEncryption_RoundTrip(t *testing.T) {
	setupCredentialCryptoForTest(t)
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResources(t, fx, "res-cred-roundtrip")
		ctx := context.Background()
		repo := store.NewResourceCredentialRepositorySQLC(fx.Runtime)

		pwd := []byte("plaintext-pwd")
		opts := []byte(`{"opt":"plaintext"}`)
		cred := &domain.ResourceCredential{
			ResourceID: "res-cred-roundtrip",
			Username:   "u",
			Password:   pwd,
			Options:    opts,
		}
		require.NoError(t, repo.Upsert(ctx, cred))

		// (a) raw read — both columns must carry ciphertext.
		var rawPwd, rawOpts []byte
		switch fx.Runtime.Driver {
		case database.DriverPostgres:
			err := fx.Runtime.PgxPool().QueryRow(ctx,
				"SELECT password, options FROM resource_credentials WHERE resource_id=$1",
				"res-cred-roundtrip").Scan(&rawPwd, &rawOpts)
			require.NoError(t, err)
		case database.DriverSQLite:
			err := fx.Runtime.SQLiteDB().QueryRowContext(ctx,
				"SELECT password, options FROM resource_credentials WHERE resource_id=?",
				"res-cred-roundtrip").Scan(&rawPwd, &rawOpts)
			require.NoError(t, err)
		}
		assert.NotEqual(t, pwd, rawPwd, "password column must contain ciphertext")
		assert.NotEqual(t, opts, rawOpts, "options column must contain ciphertext")

		// (b) port read returns plaintext.
		got, err := repo.Get(ctx, "res-cred-roundtrip")
		require.NoError(t, err)
		assert.Equal(t, pwd, got.Password)
		assert.Equal(t, opts, got.Options)
	})
}
