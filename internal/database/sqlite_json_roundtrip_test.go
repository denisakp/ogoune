package database

import (
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestSQLiteJSONAndBinaryFieldsRoundTrip(t *testing.T) {
	runtime := openSQLiteTestRuntime(t)
	now := time.Now().UTC().Round(time.Second)

	resource := domain.Resource{
		Name:     "API",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com/health",
		Interval: 60,
		Timeout:  5,
	}
	require.NoError(t, runtime.DB.Create(&resource).Error)

	incident := domain.Incident{
		ResourceID: resource.ID,
		StartedAt:  now,
		Details:    []byte("timeout"),
	}
	require.NoError(t, runtime.DB.Create(&incident).Error)

	diagnostics := domain.IncidentDiagnostics{
		IncidentID:      incident.ID,
		RequestMethod:   "GET",
		RequestURL:      resource.Target,
		RequestHeaders:  map[string]string{"authorization": "redacted"},
		ResponseHeaders: map[string]string{"content-type": "application/json"},
		FailureType:     "timeout",
		ErrorSummary:    "Timed out",
	}
	require.NoError(t, runtime.DB.Create(&diagnostics).Error)

	channel := domain.NotificationChannel{
		Name:   "Primary SMTP",
		Type:   domain.NotificationChannelTypeSMTP,
		Config: []byte(`{"recipient":"ops@example.com"}`),
	}
	require.NoError(t, runtime.DB.Create(&channel).Error)

	user := domain.User{
		Email:                "ops@example.com",
		HashedPassword:       "hashed-password",
		TwoFactorBackupCodes: []byte(`["code-1","code-2"]`),
	}
	require.NoError(t, runtime.DB.Create(&user).Error)

	var storedDiagnostics domain.IncidentDiagnostics
	require.NoError(t, runtime.DB.First(&storedDiagnostics, "id = ?", diagnostics.ID).Error)
	require.Equal(t, diagnostics.RequestHeaders, storedDiagnostics.RequestHeaders)
	require.Equal(t, diagnostics.ResponseHeaders, storedDiagnostics.ResponseHeaders)

	var storedChannel domain.NotificationChannel
	require.NoError(t, runtime.DB.First(&storedChannel, "id = ?", channel.ID).Error)
	require.Equal(t, channel.Config, storedChannel.Config)

	var storedUser domain.User
	require.NoError(t, runtime.DB.First(&storedUser, "id = ?", user.ID).Error)
	require.Equal(t, user.TwoFactorBackupCodes, storedUser.TwoFactorBackupCodes)
}
