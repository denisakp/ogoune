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

func TestRemoveAddressFromConfig(t *testing.T) {
	const me = " Jane@Example.com "
	cases := map[string]struct {
		config, want  string
		changed, urls []string
	}{
		"single field emptied":         {`{"to":"jane@example.com"}`, `{"to":""}`, []string{"to"}, nil},
		"list keeps the others":        {`{"to":"ops@example.com,  JANE@example.com ;x@y.io"}`, `{"to":"ops@example.com, x@y.io"}`, []string{"to"}, nil},
		"space separated list":         {`{"to":"ops@example.com jane@example.com"}`, `{"to":"ops@example.com"}`, []string{"to"}, nil},
		"display-name form":            {`{"to":"Ops Team <ops@example.com>, Jane Doe <jane@example.com>"}`, `{"to":"Ops Team <ops@example.com>"}`, []string{"to"}, nil},
		"nested object":                {`{"smtp":{"from":"bot@example.com","reply_to":"jane@example.com"}}`, `{"smtp":{"from":"bot@example.com","reply_to":""}}`, []string{"smtp.reply_to"}, nil},
		"array element removed":        {`{"recipients":["ops@example.com","jane@example.com"]}`, `{"recipients":["ops@example.com"]}`, []string{"recipients[1]"}, nil},
		"array emptied":                {`{"recipients":["jane@example.com"]}`, `{"recipients":[]}`, []string{"recipients[0]"}, nil},
		"array element holding a list": {`{"cc":["a@x.io; jane@example.com"]}`, `{"cc":["a@x.io"]}`, []string{"cc[0]"}, nil},
		"twice in one config":          {`{"recipients":["jane@example.com"],"cc":["JANE@example.com","b@x.io"]}`, `{"recipients":[],"cc":["b@x.io"]}`, []string{"cc[0]", "recipients[0]"}, nil},
		"url left alone and reported":  {`{"url":"https://hooks.example.com/n?to=jane@example.com","to":"jane@example.com"}`, `{"url":"https://hooks.example.com/n?to=jane@example.com","to":""}`, []string{"to"}, []string{"url"}},
		"numbers and bools kept":       {`{"port":587,"big":12345678901234567890,"tls":true,"to":"jane@example.com","x":null}`, `{"port":587,"big":12345678901234567890,"tls":true,"to":"","x":null}`, []string{"to"}, nil},
		"prefix is not a match":        {`{"to":"jane@example.com.evil"}`, `{"to":"jane@example.com.evil"}`, nil, nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out, changed, urls, err := RemoveAddressFromConfig([]byte(tc.config), me)
			require.NoError(t, err)
			assert.JSONEq(t, tc.want, string(out))
			sort.Strings(changed)
			assert.Equal(t, tc.changed, changed)
			assert.Equal(t, tc.urls, urls)

			// What Find sees, Remove removes -- except inside URLs.
			left, err := FindAddressInConfig(out, me)
			require.NoError(t, err)
			assert.Equal(t, tc.urls, left, "only URL matches remain after removal")
		})
	}
}

func TestRemoveAddressFromConfig_NoMatchReturnsInputUnchanged(t *testing.T) {
	in := []byte(`{"b":1,  "a":"ops@example.com"}`)
	out, changed, urls, err := RemoveAddressFromConfig(in, "jane@example.com")
	require.NoError(t, err)
	assert.Equal(t, in, out, "byte-for-byte: a channel without the address is not rewritten")
	assert.Nil(t, changed)
	assert.Nil(t, urls)
}

func TestRemoveAddressFromConfig_EdgeInputs(t *testing.T) {
	out, changed, _, err := RemoveAddressFromConfig([]byte(`{"to":"jane@example.com"}`), "  ")
	require.NoError(t, err)
	assert.JSONEq(t, `{"to":"jane@example.com"}`, string(out), "empty address removes nothing")
	assert.Nil(t, changed)

	_, _, _, err = RemoveAddressFromConfig([]byte(`not json`), "jane@example.com")
	assert.Error(t, err)
}
