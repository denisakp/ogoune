package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

// secHeadersOKHandler is a simple handler that returns 200 OK.
var secHeadersOKHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestSecurityHeaders_StaticHeaders(t *testing.T) {
	mw := SecurityHeaders(SecurityHeadersConfig{
		AppEnv: "development",
	})
	handler := mw(secHeadersOKHandler)

	tests := []struct {
		header string
		want   string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Permissions-Policy", "camera=(), microphone=(), geolocation=()"},
		{"Content-Security-Policy", defaultCSP},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/anything", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			got := rec.Header().Get(tt.header)
			if got != tt.want {
				t.Errorf("%s = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestSecurityHeaders_ErrorResponses(t *testing.T) {
	mw := SecurityHeaders(SecurityHeadersConfig{
		AppEnv: "development",
	})

	tests := []struct {
		name   string
		status int
	}{
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			})
			handler := mw(errorHandler)

			req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}

			// All static headers must be present even on error responses.
			headers := []string{
				"X-Content-Type-Options",
				"X-Frame-Options",
				"Referrer-Policy",
				"Permissions-Policy",
				"Content-Security-Policy",
			}
			for _, h := range headers {
				if rec.Header().Get(h) == "" {
					t.Errorf("header %s missing on %d response", h, tt.status)
				}
			}
		})
	}
}

func TestSecurityHeaders_HSTS(t *testing.T) {
	const hstsValue = "max-age=31536000; includeSubDomains"

	tests := []struct {
		name     string
		appEnv   string
		tls      bool
		xfp      string // X-Forwarded-Proto
		fwd      string // Forwarded header
		wantHSTS bool
	}{
		{
			name:     "development over HTTP — no HSTS",
			appEnv:   "development",
			wantHSTS: false,
		},
		{
			name:     "production over HTTP — HSTS present",
			appEnv:   "production",
			wantHSTS: true,
		},
		{
			name:     "development with direct TLS — HSTS present",
			appEnv:   "development",
			tls:      true,
			wantHSTS: true,
		},
		{
			name:     "development with X-Forwarded-Proto https — HSTS present",
			appEnv:   "development",
			xfp:      "https",
			wantHSTS: true,
		},
		{
			name:     "development with Forwarded proto=https — HSTS present",
			appEnv:   "development",
			fwd:      "for=192.0.2.1;proto=https",
			wantHSTS: true,
		},
		{
			name:     "development with X-Forwarded-Proto http — no HSTS",
			appEnv:   "development",
			xfp:      "http",
			wantHSTS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := SecurityHeaders(SecurityHeadersConfig{
				AppEnv: tt.appEnv,
			})
			handler := mw(secHeadersOKHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.tls {
				req.TLS = &tls.ConnectionState{}
			}
			if tt.xfp != "" {
				req.Header.Set("X-Forwarded-Proto", tt.xfp)
			}
			if tt.fwd != "" {
				req.Header.Set("Forwarded", tt.fwd)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			got := rec.Header().Get("Strict-Transport-Security")
			if tt.wantHSTS && got != hstsValue {
				t.Errorf("Strict-Transport-Security = %q, want %q", got, hstsValue)
			}
			if !tt.wantHSTS && got != "" {
				t.Errorf("Strict-Transport-Security = %q, want empty", got)
			}
		})
	}
}

func TestSecurityHeaders_SwaggerCSP(t *testing.T) {
	tests := []struct {
		name          string
		enableSwagger bool
		path          string
		wantCSP       string
	}{
		{
			name:          "swagger enabled — docs path gets relaxed CSP",
			enableSwagger: true,
			path:          "/v1/docs/index.html",
			wantCSP:       swaggerCSP,
		},
		{
			name:          "swagger enabled — non-docs path gets strict CSP",
			enableSwagger: true,
			path:          "/v1/monitors",
			wantCSP:       defaultCSP,
		},
		{
			name:          "swagger disabled — docs path gets strict CSP",
			enableSwagger: false,
			path:          "/v1/docs/index.html",
			wantCSP:       defaultCSP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := SecurityHeaders(SecurityHeadersConfig{
				AppEnv:            "development",
				EnableSwagger:     tt.enableSwagger,
				SwaggerPathPrefix: "/v1/docs/",
			})
			handler := mw(secHeadersOKHandler)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			got := rec.Header().Get("Content-Security-Policy")
			if got != tt.wantCSP {
				t.Errorf("Content-Security-Policy = %q, want %q", got, tt.wantCSP)
			}
		})
	}
}
