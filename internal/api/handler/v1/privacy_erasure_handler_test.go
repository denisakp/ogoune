package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/repository"
)

// Spec 095: the erasure endpoints. The operator is Jane (privUserID); the
// install also holds a former account, old@example.com.

func (e *privacyEnv) post(t *testing.T, path string, body any, method string) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/privacy"+path, bytes.NewReader(b))
	req.RemoteAddr = "198.51.100.30:5000"
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, asUser(req, method))
	return rec
}

func TestErasureEndpoints_RefuseAPIKeys(t *testing.T) {
	e := newPrivacyEnv(t, false, 100)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/privacy/accounts", nil)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, asUser(req, "api_key"))
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, http.StatusForbidden, e.post(t, "/erasure/preview", map[string]string{"email": "ops@example.com"}, "api_key").Code)
	assert.Equal(t, http.StatusForbidden, e.post(t, "/erasure", map[string]string{"email": "ops@example.com", "confirm_email": "ops@example.com", "password": privPassword}, "api_key").Code)
	assert.Empty(t, e.erasures.Applied)
}

func TestErasureAccounts_ExcludeTheCaller(t *testing.T) {
	e := newPrivacyEnv(t, false, 100)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/privacy/accounts", nil)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, asUser(req, "jwt"))
	require.Equal(t, http.StatusOK, rec.Code)
	var got struct {
		Data []struct{ ID, Email string } `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Len(t, got.Data, 1)
	assert.Equal(t, "user-old", got.Data[0].ID)
	assert.Equal(t, "old@example.com", got.Data[0].Email)
}

func TestPreviewErasure(t *testing.T) {
	e := newPrivacyEnv(t, false, 100)
	rec := e.post(t, "/erasure/preview", map[string]string{"email": " OPS@example.com"}, "jwt")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var got struct {
		Data struct {
			Kind     string `json:"kind"`
			Channels []struct {
				ID          string   `json:"id"`
				Fields      []string `json:"fields"`
				WillDisable bool     `json:"will_disable"`
			} `json:"channels"`
			ManualReview []struct{ Reason string } `json:"manual_review"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "address", got.Data.Kind)
	require.Len(t, got.Data.Channels, 1)
	assert.Equal(t, "ch-mail", got.Data.Channels[0].ID)
	assert.Equal(t, []string{"to"}, got.Data.Channels[0].Fields)
	assert.NotContains(t, rec.Body.String(), "ops@example.com", "paths, never values")
	assert.NotContains(t, rec.Body.String(), privSMTPPass)
	assert.Empty(t, e.erasures.Applied, "a preview changes nothing")

	cases := []struct {
		name string
		body map[string]string
		code int
		want string
	}{
		{"own address", map[string]string{"email": "jane@example.com"}, http.StatusUnprocessableEntity, "/problems/cannot-erase-self"},
		{"own account", map[string]string{"account_id": privUserID}, http.StatusUnprocessableEntity, "/problems/cannot-erase-self"},
		{"unknown account", map[string]string{"account_id": "nope"}, http.StatusNotFound, "account not found"},
		{"neither", map[string]string{}, http.StatusBadRequest, "BAD_REQUEST"},
		{"not an address", map[string]string{"email": "ops"}, http.StatusBadRequest, "BAD_REQUEST"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := e.post(t, "/erasure/preview", tc.body, "jwt")
			assert.Equal(t, tc.code, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.want)
		})
	}
}

func TestErase_FailedConfirmationIsOne422(t *testing.T) {
	e := newPrivacyEnv(t, true, 100)
	cases := map[string]map[string]string{
		"wrong password":        {"email": "ops@example.com", "confirm_email": "ops@example.com", "password": "nope", "code": validCode(t)},
		"missing code":          {"email": "ops@example.com", "confirm_email": "ops@example.com", "password": privPassword},
		"wrong code":            {"email": "ops@example.com", "confirm_email": "ops@example.com", "password": privPassword, "code": "000000"},
		"typed address differs": {"email": "ops@example.com", "confirm_email": "ops@example.org", "password": privPassword, "code": validCode(t)},
	}
	var bodies []string
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			rec := e.post(t, "/erasure", body, "jwt")
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "never 401: the client would sign the operator out")
			bodies = append(bodies, strings.ReplaceAll(rec.Body.String(), "\n", ""))
		})
	}
	for _, b := range bodies {
		assert.Contains(t, b, "/problems/invalid-credentials")
		assert.NotContains(t, b, "ops@example")
	}
	assert.Empty(t, e.erasures.Applied, "nothing changed")
}

func TestErase_Address(t *testing.T) {
	e := newPrivacyEnv(t, true, 100)
	rec := e.post(t, "/erasure", map[string]string{"email": "ops@example.com", "confirm_email": " OPS@example.com ", "password": privPassword, "code": validCode(t)}, "jwt")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var got struct {
		Data struct {
			RecordID string         `json:"record_id"`
			Changes  map[string]int `json:"changes"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.NotEmpty(t, got.Data.RecordID)
	assert.Equal(t, 1, got.Data.Changes["channels"])
	require.Len(t, e.erasures.Applied, 1)
	assert.NotContains(t, rec.Body.String(), "ops@example.com")
}

func TestErase_FormerAccount(t *testing.T) {
	e := newPrivacyEnv(t, false, 100)
	rec := e.post(t, "/erasure", map[string]string{"account_id": "user-old", "confirm_email": "old@example.com", "password": privPassword}, "jwt")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Len(t, e.erasures.Applied, 1)
	assert.Equal(t, "user-old", e.erasures.Applied[0].AccountID)
}

func TestErase_ConflictAndFailureChangeNothing(t *testing.T) {
	e := newPrivacyEnv(t, false, 100)
	body := map[string]string{"email": "ops@example.com", "confirm_email": "ops@example.com", "password": privPassword}

	e.erasures.ApplyErr = &repository.ErasureConflictError{ChannelID: "ch-mail"}
	rec := e.post(t, "/erasure", body, "jwt")
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "/problems/erasure-conflict")

	e.erasures.ApplyErr = assert.AnError
	rec = e.post(t, "/erasure", body, "jwt")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "nothing was changed")
	assert.NotContains(t, rec.Body.String(), "ops@example.com")
}

func TestErase_IsRateLimited(t *testing.T) {
	e := newPrivacyEnv(t, false, 2)
	body := map[string]string{"email": "ops@example.com", "confirm_email": "nope@example.com", "password": privPassword}
	e.post(t, "/erasure", body, "jwt")
	e.post(t, "/erasure", body, "jwt")
	assert.Equal(t, http.StatusTooManyRequests, e.post(t, "/erasure", body, "jwt").Code)
}
