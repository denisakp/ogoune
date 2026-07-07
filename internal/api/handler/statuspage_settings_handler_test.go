package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
)

func newSettingsHandler(t *testing.T) (*StatusPageSettingsHandler, *fake.StatusPageSettingsFake, func()) {
	t.Helper()
	repo := fake.NewStatusPageSettingsFake()
	svc := service.NewStatusPageSettingsService(repo)
	dir := t.TempDir()
	t.Setenv("STATIC_DIR", dir)
	return NewStatusPageSettingsHandler(svc), repo, func() { _ = os.RemoveAll(dir) }
}

func multipartBody(t *testing.T, contentType string, payload []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="logo.png"`)
	header.Set("Content-Type", contentType)
	part, err := w.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(payload)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return body, w.FormDataContentType()
}

func TestUploadLogo_Happy(t *testing.T) {
	h, _, cleanup := newSettingsHandler(t)
	defer cleanup()

	body, ct := multipartBody(t, "image/png", []byte("fake-png-bytes"))
	req := httptest.NewRequest(http.MethodPost, "/?slot=light", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	h.UploadLogo(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "/static/uploads/statuspage/light-")
}

func TestUploadLogo_UnknownSlot422(t *testing.T) {
	h, _, cleanup := newSettingsHandler(t)
	defer cleanup()

	body, ct := multipartBody(t, "image/png", []byte("bytes"))
	req := httptest.NewRequest(http.MethodPost, "/?slot=cover", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	h.UploadLogo(rec, req)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "UNKNOWN_SLOT")
}

func TestUploadLogo_InvalidMime422(t *testing.T) {
	h, _, cleanup := newSettingsHandler(t)
	defer cleanup()

	body, ct := multipartBody(t, "application/pdf", []byte("PDF"))
	req := httptest.NewRequest(http.MethodPost, "/?slot=light", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	h.UploadLogo(rec, req)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "LOGO_INVALID_MIME")
}

func TestUploadLogo_TooLarge422(t *testing.T) {
	h, _, cleanup := newSettingsHandler(t)
	defer cleanup()

	big := bytes.Repeat([]byte("A"), maxLogoBytes+10)
	body, ct := multipartBody(t, "image/png", big)
	req := httptest.NewRequest(http.MethodPost, "/?slot=light", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	h.UploadLogo(rec, req)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "LOGO_TOO_LARGE")
}

func TestDeleteLogo_ClearsColumnAnd204(t *testing.T) {
	h, _, cleanup := newSettingsHandler(t)
	defer cleanup()

	// First upload one.
	body, ct := multipartBody(t, "image/png", []byte("png"))
	upReq := httptest.NewRequest(http.MethodPost, "/?slot=light", body)
	upReq.Header.Set("Content-Type", ct)
	upRec := httptest.NewRecorder()
	h.UploadLogo(upRec, upReq)
	require.Equal(t, http.StatusOK, upRec.Code)

	// Delete it.
	delReq := httptest.NewRequest(http.MethodDelete, "/?slot=light", nil)
	delRec := httptest.NewRecorder()
	h.DeleteLogo(delRec, delReq)
	require.Equal(t, http.StatusNoContent, delRec.Code)

	// Second delete on the same empty slot → 404.
	delReq2 := httptest.NewRequest(http.MethodDelete, "/?slot=light", nil)
	delRec2 := httptest.NewRecorder()
	h.DeleteLogo(delRec2, delReq2)
	require.Equal(t, http.StatusNotFound, delRec2.Code)

	// The directory should be empty (or only contain other slots).
	uploadDirPath := filepath.Join(os.Getenv("STATIC_DIR"), "uploads", "statuspage")
	entries, _ := os.ReadDir(uploadDirPath)
	for _, e := range entries {
		assert.False(t, strings.HasPrefix(e.Name(), "light-"), "stray file %s", e.Name())
	}
}
