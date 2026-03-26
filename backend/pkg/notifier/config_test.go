package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSMTPNotifierFromConfig_AcceptsNumericPort(t *testing.T) {
	cfg := `{"host":"smtp.example.com","port":587,"username":"user","password":"pass","sender":"noreply@example.com","recipients":["ops@example.com"]}`

	n, err := NewSMTPNotifierFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, n)
	assert.Equal(t, "587", n.smtpPort)
	assert.Equal(t, "ops@example.com", n.recipient)
	assert.Equal(t, "user", n.smtpUser)
}

func TestNewSMTPNotifierFromConfig_AcceptsLegacyFields(t *testing.T) {
	cfg := `{"host":"smtp.example.com","port":"587","user":"legacy","password":"pass","sender":"noreply@example.com","recipient":"admin@example.com"}`

	n, err := NewSMTPNotifierFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, n)
	assert.Equal(t, "587", n.smtpPort)
	assert.Equal(t, "admin@example.com", n.recipient)
	assert.Equal(t, "legacy", n.smtpUser)
}
