package domain

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEmail(t *testing.T) {
	assert.Equal(t, "jane@example.com", NormalizeEmail("  Jane@Example.COM \n"))
	assert.Equal(t, "", NormalizeEmail("   "))
}

func TestFindAddressInConfig(t *testing.T) {
	const me = "Jane@Example.com"
	cases := map[string]struct {
		config string
		want   []string
	}{
		"plain field":                    {`{"to":"jane@example.com"}`, []string{"to"}},
		"comma list, spaces, other case": {`{"to":"ops@example.com,  JANE@example.com ;x@y.io"}`, []string{"to"}},
		"display-name form":              {`{"to":"Jane Doe <jane@example.com>"}`, []string{"to"}},
		"nested object":                  {`{"smtp":{"from":"bot@example.com","reply_to":"jane@example.com"}}`, []string{"smtp.reply_to"}},
		"array element":                  {`{"recipients":["ops@example.com","jane@example.com"]}`, []string{"recipients[1]"}},
		"url containing the address":     {`{"url":"https://hooks.example.com/notify?to=jane@example.com&t=secret"}`, []string{"url"}},
		"url not containing it":          {`{"url":"https://hooks.slack.com/services/T0/B0/xyz"}`, nil},
		"prefix is not a match":          {`{"to":"jane@example.com.evil"}`, nil},
		"substring outside a url is not": {`{"to":"notjane@example.com"}`, nil},
		"several fields":                 {`{"to":"jane@example.com","cc":["jane@example.com"]}`, []string{"cc[0]", "to"}},
		"non-string values ignored":      {`{"port":587,"tls":true,"to":null}`, nil},
		"empty config":                   {``, nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := FindAddressInConfig([]byte(tc.config), me)
			require.NoError(t, err)
			sort.Strings(got)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFindAddressInConfig_ReturnsPathsNeverValues(t *testing.T) {
	got, err := FindAddressInConfig([]byte(`{"to":"ops@example.com, jane@example.com","password":"hunter2"}`), "jane@example.com")
	require.NoError(t, err)
	assert.Equal(t, []string{"to"}, got)
	for _, p := range got {
		assert.NotContains(t, p, "@")
		assert.NotContains(t, p, "hunter2")
	}
}

func TestFindAddressInConfig_InvalidJSONIsAnError(t *testing.T) {
	_, err := FindAddressInConfig([]byte(`not json`), "jane@example.com")
	assert.Error(t, err)
}

func TestFindAddressInConfig_EmptyAddressMatchesNothing(t *testing.T) {
	got, err := FindAddressInConfig([]byte(`{"to":""}`), "  ")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestInventoryCounts(t *testing.T) {
	inv := &PersonalDataInventory{
		Sessions: make([]SessionData, 3),
		APIKeys:  make([]APIKeyData, 2),
		Updates:  make([]AuthoredUpdate, 1),
		Channels: make([]AddressInChannel, 2),
		Reports:  ReportData{IsRecipient: true, Sent: make([]ReportSent, 4)},
	}
	assert.Equal(t, map[string]int{
		"account": 1, "sessions": 3, "api_keys": 2, "incident_updates": 1,
		"notification_channels": 2, "reports": 5,
	}, inv.Counts(), "reports = sent + 1 when the person is the configured recipient")

	empty := &PersonalDataInventory{}
	assert.Equal(t, 0, empty.Counts()["reports"])
	assert.Len(t, empty.Counts(), len(PrivacyCategories), "every category present, zeros included")
}
