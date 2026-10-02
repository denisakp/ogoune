package v1_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	v1 "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/internal/service"
)

// Values seeded so the secret scan has something real to look for.
const (
	privPassword   = "correct-horse-battery"
	privTOTPSecret = "JBSWY3DPEHPK3PXP"
	privRawAPIKey  = "pk_live_ab12SECRETKEYMATERIAL0123456789"
	privSMTPPass   = "smtp-password-should-never-leak"
	privHookToken  = "xoxb-hook-token-should-never-leak"
	privUserID     = "user-privacy"
)

type privacyEnv struct {
	router   *chi.Mux
	erasures *fake.ErasureFake
	keys     *fake.APIKeyRepository
	users   *fake.UserRepository
	hash    string
	keyHash string
	twoFA   bool
}

// newPrivacyEnv wires the real AuthService and PrivacyService over fakes, and
// mounts the routes exactly as the router does (RequireJWTOnly; the export
// behind a per-address limiter).
func newPrivacyEnv(t *testing.T, twoFactor bool, exportLimit int) *privacyEnv {
	t.Helper()
	ctx := context.Background()
	hash, err := bcrypt.GenerateFromPassword([]byte(privPassword), bcrypt.MinCost)
	require.NoError(t, err)

	users := fake.NewUserRepository()
	u := &domain.User{Base: domain.Base{ID: privUserID}, Email: "Jane@Example.com", Name: "Jane", HashedPassword: string(hash), PasswordInitialized: true}
	if twoFactor {
		u.TwoFactorEnabled, u.TwoFactorSecret = true, privTOTPSecret
	}
	_, err = users.Create(ctx, u)
	require.NoError(t, err)

	sessions := fake.NewSessionRepository()
	require.NoError(t, sessions.Create(ctx, &domain.Session{UserID: privUserID, IP: "203.0.113.4", Browser: "Firefox", LastActiveAt: time.Now()}))

	keys := fake.NewAPIKeyRepository()
	keyHash := fmt.Sprintf("%x", []byte(privRawAPIKey))
	require.NoError(t, keys.Create(ctx, &domain.APIKey{UserID: privUserID, Name: "ci", KeyPrefix: privRawAPIKey[:12], KeyHash: keyHash, Scope: domain.APIKeyScopeReadWrite, IsActive: true}))

	updates := fake.NewIncidentUpdateRepository()
	_, err = updates.Create(ctx, &domain.IncidentUpdate{IncidentID: "inc-1", Status: "identified", Message: "on it", PostedBy: privUserID, PostedAt: time.Now()})
	require.NoError(t, err)

	channels := fake.NewNotificationChannelFake()
	require.NoError(t, channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-mail"}, Name: "On-call", Type: "smtp",
		Config: []byte(`{"to":"ops@example.com, jane@example.com","password":"` + privSMTPPass + `"}`)}))
	require.NoError(t, channels.Create(ctx, &domain.NotificationChannel{Base: domain.Base{ID: "ch-hook"}, Name: "Hook", Type: "webhook",
		Config: []byte(`{"url":"https://hooks.example.com/x?to=jane@example.com&token=` + privHookToken + `"}`)}))

	settings := fake.NewReportSettingsFake()
	_, err = settings.Upsert(ctx, &domain.ReportSettings{RecipientEmail: "jane@example.com"})
	require.NoError(t, err)
	history := fake.NewReportHistoryFake()

	auth := service.NewAuthService(users, service.NewJWTManager("test-secret-key-at-least-32-bytes-long", "ogoune", time.Hour))
	privacy := service.NewPrivacyService(users, sessions, keys, updates, channels, settings, history)
	t.Setenv("APP_SECRET_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	_, err = users.Create(ctx, &domain.User{Base: domain.Base{ID: "user-old"}, Email: "old@example.com", Name: "Old admin"})
	require.NoError(t, err)
	erasures := fake.NewErasureFake()
	eraser := service.NewErasureService(users, sessions, keys, updates, channels, settings, history, erasures)
	h := v1.NewPrivacyHandler(privacy, eraser, auth, "1.0.0-test", "")

	r := chi.NewRouter()
	r.Route("/api/v1/me/privacy", func(r chi.Router) {
		r.Use(middleware.RequireJWTOnly)
		r.Get("/", h.Summary)
		r.With(httprate.LimitByIP(exportLimit, time.Minute)).Post("/export", h.Export)
		r.Get("/accounts", h.ErasureAccounts)
		r.Post("/erasure/preview", h.PreviewErasure)
		r.With(httprate.LimitByIP(exportLimit, time.Minute)).Post("/erasure", h.Erase)
	})
	return &privacyEnv{router: r, erasures: erasures, keys: keys, users: users, hash: string(hash), keyHash: keyHash, twoFA: twoFactor}
}

func asUser(req *http.Request, method string) *http.Request {
	ctx := context.WithValue(req.Context(), "user_id", privUserID) //nolint:staticcheck // existing key
	ctx = context.WithValue(ctx, "auth_method", method)            //nolint:staticcheck // existing key
	return req.WithContext(ctx)
}

func (e *privacyEnv) export(t *testing.T, password, code, method string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"password": password, "code": code})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/privacy/export", bytes.NewReader(body))
	req.RemoteAddr = "198.51.100.20:5000"
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, asUser(req, method))
	return rec
}

func (e *privacyEnv) summary(t *testing.T, method string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/privacy/", nil)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, asUser(req, method))
	return rec
}

func validCode(t *testing.T) string {
	t.Helper()
	c, err := totp.GenerateCode(privTOTPSecret, time.Now())
	require.NoError(t, err)
	return c
}

func TestPrivacyExport_RefusesAPIKeys(t *testing.T) {
	e := newPrivacyEnv(t, false, 100)
	assert.Equal(t, http.StatusForbidden, e.export(t, privPassword, "", "api_key").Code)
	assert.Equal(t, http.StatusForbidden, e.summary(t, "api_key").Code)
}

// FR-009a + analyze I1: every failure is the same 422 -- never 401, which the
// web client would turn into a sign-out -- and the session stays usable.
func TestPrivacyExport_ReauthFailuresAreOne422AndKeepTheSession(t *testing.T) {
	cases := map[string]struct {
		twoFA          bool
		password, code string
	}{
		"missing password":            {false, "", ""},
		"wrong password":              {false, "nope", ""},
		"two-factor on, missing code": {true, privPassword, ""},
		"two-factor on, wrong code":   {true, privPassword, "000000"},
	}
	var bodies []string
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := newPrivacyEnv(t, tc.twoFA, 100)
			rec := e.export(t, tc.password, tc.code, "jwt")
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Empty(t, rec.Header().Get("Content-Disposition"), "no file on failure")
			var pd map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &pd))
			delete(pd, "instance")
			b, _ := json.Marshal(pd)
			bodies = append(bodies, string(b))

			assert.Equal(t, http.StatusOK, e.summary(t, "jwt").Code, "still signed in after a failed attempt")
		})
	}
	sort.Strings(bodies)
	for i := 1; i < len(bodies); i++ {
		assert.Equal(t, bodies[0], bodies[i], "the response never says which factor was wrong")
	}
}

func decodeExport(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var doc map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	return doc
}

func TestPrivacyExport_Success(t *testing.T) {
	e := newPrivacyEnv(t, true, 100)
	rec := e.export(t, privPassword, validCode(t), "jwt")
	doc := decodeExport(t, rec)

	assert.Regexp(t, `^attachment; filename="ogoune-personal-data-\d{4}-\d{2}-\d{2}\.json"$`, rec.Header().Get("Content-Disposition"))
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	assert.Equal(t, "ogoune-personal-data/1", doc["format"])
	assert.Equal(t, []any{"account", "sessions", "api_keys", "incident_updates", "notification_channels", "reports"}, doc["covers"])
	assert.Equal(t, []any{"monitors", "checks", "incidents", "host_metrics", "kernel_events"}, doc["not_personal_data"])
	install := doc["install"].(map[string]any)
	assert.Equal(t, "1.0.0-test", install["version"])
	_, hasBase := install["base_url"]
	assert.False(t, hasBase, "no base_url unless APP_BASE_URL is set")

	assert.Len(t, doc["sessions"], 1)
	assert.Len(t, doc["api_keys"], 1)
	assert.Len(t, doc["incident_updates"], 1)
	channels := doc["notification_channels"].([]any)
	require.Len(t, channels, 2)
	assert.Equal(t, true, doc["reports"].(map[string]any)["is_recipient"])
}

// SC-002: none of the seeded secrets, in any form, anywhere in the document.
func TestPrivacyExport_ContainsNoSecret(t *testing.T) {
	e := newPrivacyEnv(t, true, 100)
	body := e.export(t, privPassword, validCode(t), "jwt").Body.String()
	require.Contains(t, body, `"format"`)

	for name, secret := range map[string]string{
		"password":        privPassword,
		"password hash":   e.hash,
		"totp secret":     privTOTPSecret,
		"raw api key":     privRawAPIKey,
		"api key hash":    e.keyHash,
		"smtp password":   privSMTPPass,
		"webhook token":   privHookToken,
		"other recipient": "ops@example.com",
		"bcrypt prefix":   "$2a$",
		"otpauth uri":     "otpauth://",
		"jwt":             "eyJ",
	} {
		assert.NotContains(t, body, secret, name)
	}
	assert.Contains(t, body, privRawAPIKey[:12], "the key appears only as its prefix")
}

// SC-003: every count on the page equals the matching array in an export.
func TestPrivacySummary_MatchesTheExport(t *testing.T) {
	e := newPrivacyEnv(t, false, 100)
	rec := e.summary(t, "jwt")
	require.Equal(t, http.StatusOK, rec.Code)
	var env struct {
		Data struct {
			TwoFactorEnabled bool `json:"two_factor_enabled"`
			Categories       []struct {
				Key        string `json:"key"`
				Count      int    `json:"count"`
				ManagePath string `json:"manage_path"`
			} `json:"categories"`
			NotPersonalData []string `json:"not_personal_data"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	doc := decodeExport(t, e.export(t, privPassword, "", "jwt"))

	reports := doc["reports"].(map[string]any)
	want := map[string]int{
		"account":               1,
		"sessions":              len(doc["sessions"].([]any)),
		"api_keys":              len(doc["api_keys"].([]any)),
		"incident_updates":      len(doc["incident_updates"].([]any)),
		"notification_channels": len(doc["notification_channels"].([]any)),
		"reports":               len(reports["sent"].([]any)) + 1, // is_recipient
	}
	got := map[string]int{}
	for _, c := range env.Data.Categories {
		got[c.Key] = c.Count
		assert.NotEmpty(t, c.ManagePath, c.Key)
	}
	assert.Equal(t, want, got)
	assert.NotEmpty(t, env.Data.NotPersonalData)
}

// FR-009b: the export shares sign-in's per-address limit.
func TestPrivacyExport_RateLimited(t *testing.T) {
	e := newPrivacyEnv(t, false, 3)
	for i := 0; i < 3; i++ {
		assert.Equal(t, http.StatusUnprocessableEntity, e.export(t, "nope", "", "jwt").Code)
	}
	assert.Equal(t, http.StatusTooManyRequests, e.export(t, "nope", "", "jwt").Code)
}

// The export is logged as an event -- who and how it ended -- never a
// credential or the document.
func TestPrivacyExport_LogsOutcomeOnly(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	e := newPrivacyEnv(t, false, 100)
	e.export(t, "wrong-guess", "", "jwt")
	e.export(t, privPassword, "", "jwt")

	logs := buf.String()
	assert.Contains(t, logs, "outcome=refused")
	assert.Contains(t, logs, "outcome=ok")
	assert.Contains(t, logs, "user_id="+privUserID)
	assert.NotContains(t, logs, "wrong-guess")
	assert.NotContains(t, logs, privPassword)
	assert.NotContains(t, logs, "203.0.113.4", "document content never reaches the log")
}

// SC-004 / FR-013: the feature writes nothing. Real SQLite, real repositories:
// every table except the session activity timestamp is identical before and
// after a run of exports and summaries.
func TestPrivacy_ReadsOnly(t *testing.T) {
	fx := internaltest.SetupSQLite(t)
	ctx := context.Background()
	rt := fx.Runtime
	users := store.NewUserRepositorySQLC(rt)
	hash, _ := bcrypt.GenerateFromPassword([]byte(privPassword), bcrypt.MinCost)
	u, err := users.Create(ctx, &domain.User{Base: domain.Base{ID: privUserID}, Email: "jane@example.com", Name: "Jane", HashedPassword: string(hash), PasswordInitialized: true})
	require.NoError(t, err)
	sessions := store.NewSessionRepositorySQLC(rt)
	require.NoError(t, sessions.Create(ctx, &domain.Session{UserID: u.ID, IP: "203.0.113.4", LastActiveAt: time.Now()}))

	auth := service.NewAuthService(users, service.NewJWTManager("test-secret-key-at-least-32-bytes-long", "ogoune", time.Hour))
	privacy := service.NewPrivacyService(users, sessions, store.NewAPIKeyRepositorySQLC(rt),
		store.NewIncidentUpdateRepositorySQLC(rt), store.NewNotificationChannelRepositorySQLC(rt),
		store.NewReportSettingsRepositorySQLC(rt), store.NewReportHistoryRepositorySQLC(rt))
	h := v1.NewPrivacyHandler(privacy, nil, auth, "test", "")
	r := chi.NewRouter()
	r.Get("/s", h.Summary)
	r.Post("/e", h.Export)

	before := dbSnapshot(t, rt.SQLiteDB())
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/e", strings.NewReader(`{"password":"`+privPassword+`"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, asUser(req, "jwt"))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, asUser(httptest.NewRequest(http.MethodGet, "/s", nil), "jwt"))
		require.Equal(t, http.StatusOK, rec.Code)
	}
	assert.Equal(t, before, dbSnapshot(t, rt.SQLiteDB()), "no write of any kind")
}

func dbSnapshot(t *testing.T, db *sql.DB) map[string][]string {
	t.Helper()
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`)
	require.NoError(t, err)
	var tables []string
	for rows.Next() {
		var n string
		require.NoError(t, rows.Scan(&n))
		tables = append(tables, n)
	}
	require.NoError(t, rows.Close())

	out := map[string][]string{}
	for _, table := range tables {
		q := fmt.Sprintf(`SELECT * FROM %q`, table)
		if table == "sessions" {
			// Every authenticated request already bumps this; not the feature's write.
			q = `SELECT id, user_id, browser, os, ip, location, created_at, revoked_at FROM sessions`
		}
		r, err := db.Query(q)
		require.NoError(t, err, table)
		cols, _ := r.Columns()
		var lines []string
		for r.Next() {
			vals := make([]sql.NullString, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			require.NoError(t, r.Scan(ptrs...))
			parts := make([]string, len(cols))
			for i, v := range vals {
				parts[i] = v.String
			}
			lines = append(lines, strings.Join(parts, "|"))
		}
		require.NoError(t, r.Close())
		sort.Strings(lines)
		out[table] = lines
	}
	return out
}
