package service

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAPIKeyServiceForTest() (*APIKeyService, *fake.APIKeyRepository, *fake.UserRepository) {
	apiKeyRepo := fake.NewAPIKeyRepository()
	userRepo := fake.NewUserRepository()
	svc := NewAPIKeyService(apiKeyRepo, userRepo)
	return svc, apiKeyRepo, userRepo
}

// seedUser creates and stores a test user, returning its ID.
func seedTestUser(t *testing.T, userRepo *fake.UserRepository) string {
	t.Helper()
	user := &domain.User{
		Email:               "test@example.com",
		HashedPassword:      "hash",
		PasswordInitialized: true,
	}
	created, err := userRepo.Create(context.Background(), user)
	require.NoError(t, err)
	return created.ID
}

// T013 – create validation: name constraints, scope enum, 10-key hard limit.
func TestAPIKeyService_CreateAPIKey_Success(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	resp, err := svc.CreateAPIKey(context.Background(), userID, "CI Pipeline", domain.APIKeyScopeReadWrite, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "CI Pipeline", resp.Name)
	assert.NotEmpty(t, resp.Key)
	assert.True(t, len(resp.Key) > 0)
	assert.Equal(t, domain.APIKeyScopeReadWrite, resp.Scope)
	assert.Nil(t, resp.ExpiresAt)
}

func TestAPIKeyService_CreateAPIKey_NameTrimmed(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	resp, err := svc.CreateAPIKey(context.Background(), userID, "  My Key  ", domain.APIKeyScopeRead, nil)
	require.NoError(t, err)
	assert.Equal(t, "My Key", resp.Name)
}

func TestAPIKeyService_CreateAPIKey_EmptyNameRejected(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	_, err := svc.CreateAPIKey(context.Background(), userID, "", domain.APIKeyScopeRead, nil)
	assert.ErrorIs(t, err, ErrValidationFailed)
}

func TestAPIKeyService_CreateAPIKey_WhitespaceOnlyNameRejected(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	_, err := svc.CreateAPIKey(context.Background(), userID, "   ", domain.APIKeyScopeRead, nil)
	assert.ErrorIs(t, err, ErrValidationFailed)
}

func TestAPIKeyService_CreateAPIKey_NameTooLong(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := svc.CreateAPIKey(context.Background(), userID, string(longName), domain.APIKeyScopeRead, nil)
	assert.ErrorIs(t, err, ErrValidationFailed)
}

func TestAPIKeyService_CreateAPIKey_InvalidScope(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	_, err := svc.CreateAPIKey(context.Background(), userID, "Bad Scope Key", domain.APIKeyScope("admin"), nil)
	assert.ErrorIs(t, err, ErrValidationFailed)
}

func TestAPIKeyService_CreateAPIKey_10KeyHardLimit(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	for i := 0; i < 10; i++ {
		_, err := svc.CreateAPIKey(context.Background(), userID, "Key", domain.APIKeyScopeRead, nil)
		require.NoError(t, err)
	}

	_, err := svc.CreateAPIKey(context.Background(), userID, "One Too Many", domain.APIKeyScopeRead, nil)
	assert.ErrorIs(t, err, ErrAPIKeyLimitReached)
}

// T032 – expiry validation: past date rejected; future date accepted.
func TestAPIKeyService_CreateAPIKey_PastExpiryRejected(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	past := time.Now().Add(-1 * time.Hour)
	_, err := svc.CreateAPIKey(context.Background(), userID, "Expired Key", domain.APIKeyScopeRead, &past)
	assert.ErrorIs(t, err, ErrValidationFailed)
}

func TestAPIKeyService_CreateAPIKey_FutureExpiryAccepted(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	future := time.Now().Add(24 * time.Hour)
	resp, err := svc.CreateAPIKey(context.Background(), userID, "Temporary Key", domain.APIKeyScopeRead, &future)
	require.NoError(t, err)
	require.NotNil(t, resp.ExpiresAt)
	assert.True(t, resp.ExpiresAt.After(time.Now()))
}

// ListAPIKeys covers T024.
func TestAPIKeyService_ListAPIKeys(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	_, err := svc.CreateAPIKey(context.Background(), userID, "Key A", domain.APIKeyScopeRead, nil)
	require.NoError(t, err)
	_, err = svc.CreateAPIKey(context.Background(), userID, "Key B", domain.APIKeyScopeReadWrite, nil)
	require.NoError(t, err)

	items, err := svc.ListAPIKeys(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestAPIKeyService_ListAPIKeys_EmptyForNewUser(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	items, err := svc.ListAPIKeys(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, items)
}

// RevokeAPIKey covers T024.
func TestAPIKeyService_RevokeAPIKey_Success(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	created, err := svc.CreateAPIKey(context.Background(), userID, "To Revoke", domain.APIKeyScopeRead, nil)
	require.NoError(t, err)

	err = svc.RevokeAPIKey(context.Background(), userID, created.ID)
	require.NoError(t, err)
}

func TestAPIKeyService_RevokeAPIKey_NotFound(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	err := svc.RevokeAPIKey(context.Background(), userID, "nonexistent-id")
	assert.ErrorIs(t, err, ErrAPIKeyNotFound)
}

func TestAPIKeyService_RevokeAPIKey_EmptyID(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	err := svc.RevokeAPIKey(context.Background(), userID, "")
	assert.ErrorIs(t, err, ErrValidationFailed)
}

// AuthenticateAPIKey covers T017.
func TestAPIKeyService_AuthenticateAPIKey_Success(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	created, err := svc.CreateAPIKey(context.Background(), userID, "Auth Key", domain.APIKeyScopeReadWrite, nil)
	require.NoError(t, err)

	authenticated, err := svc.AuthenticateAPIKey(context.Background(), created.Key)
	require.NoError(t, err)
	assert.Equal(t, userID, authenticated.User.ID)
	assert.Equal(t, domain.APIKeyScopeReadWrite, authenticated.Key.Scope)
}

func TestAPIKeyService_AuthenticateAPIKey_RevokedKeyRejected(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	created, err := svc.CreateAPIKey(context.Background(), userID, "Key To Revoke", domain.APIKeyScopeRead, nil)
	require.NoError(t, err)
	rawKey := created.Key

	err = svc.RevokeAPIKey(context.Background(), userID, created.ID)
	require.NoError(t, err)

	_, err = svc.AuthenticateAPIKey(context.Background(), rawKey)
	assert.ErrorIs(t, err, ErrAPIKeyRevoked)
}

// T032 – expired keys are rejected at auth time.
func TestAPIKeyService_AuthenticateAPIKey_ExpiredKeyRejected(t *testing.T) {
	svc, _, userRepo := newAPIKeyServiceForTest()
	userID := seedTestUser(t, userRepo)

	future := time.Now().Add(24 * time.Hour)
	created, err := svc.CreateAPIKey(context.Background(), userID, "Will Expire", domain.APIKeyScopeRead, &future)
	require.NoError(t, err)

	// Override the service clock to simulate time passing past expiry.
	svc.now = func() time.Time { return time.Now().Add(48 * time.Hour) }

	_, err = svc.AuthenticateAPIKey(context.Background(), created.Key)
	assert.ErrorIs(t, err, ErrAPIKeyExpired)
}

func TestAPIKeyService_AuthenticateAPIKey_InvalidFormatRejected(t *testing.T) {
	svc, _, _ := newAPIKeyServiceForTest()

	_, err := svc.AuthenticateAPIKey(context.Background(), "Bearer not-an-api-key")
	assert.ErrorIs(t, err, ErrUnauthorized)
}
