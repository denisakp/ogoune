package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/pkg/crypto"
)

const erasureTestKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func useErasureTestKey(t *testing.T) {
	t.Helper()
	t.Setenv("APP_SECRET_KEY", erasureTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
}

func TestErasureFingerprint(t *testing.T) {
	useErasureTestKey(t)

	a, err := ErasureFingerprint("jane@example.com")
	require.NoError(t, err)
	b, err := ErasureFingerprint("  Jane@Example.COM ")
	require.NoError(t, err)
	assert.Equal(t, a, b, "case and spacing do not matter")
	assert.Len(t, a, 64, "hex HMAC-SHA256")
	assert.NotContains(t, a, "jane")

	c, err := ErasureFingerprint("john@example.com")
	require.NoError(t, err)
	assert.NotEqual(t, a, c)

	t.Setenv("APP_SECRET_KEY", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	d, err := ErasureFingerprint("jane@example.com")
	require.NoError(t, err)
	assert.NotEqual(t, a, d, "keyed: another install cannot match it")
}
