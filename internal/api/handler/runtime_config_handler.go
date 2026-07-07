package handler

import (
	"encoding/json"
	"net/http"

	"github.com/denisakp/ogoune/internal/config"
	"github.com/denisakp/ogoune/internal/ee/license"
)

// RuntimeConfigHandler exposes read-only runtime config used by the SPA
// to render edition-aware + SSL-aware wording on first paint (spec 059
// FR-030 + FR-040). Unauthenticated by design — non-sensitive flags only.
type RuntimeConfigHandler struct {
	cfg     *config.Config
	version string
}

func NewRuntimeConfigHandler(cfg *config.Config, version string) *RuntimeConfigHandler {
	return &RuntimeConfigHandler{cfg: cfg, version: version}
}

type runtimeConfigResponse struct {
	SSLProvider       string `json:"ssl_provider"`
	Edition           string `json:"edition"`
	Version           string `json:"version"`
	PoweredByRequired bool   `json:"powered_by_required"`
}

// Get handles GET /api/config/runtime.
func (h *RuntimeConfigHandler) Get(w http.ResponseWriter, _ *http.Request) {
	resp := runtimeConfigResponse{
		SSLProvider:       h.cfg.SSLProvider,
		Edition:           string(license.Get()),
		Version:           h.version,
		PoweredByRequired: license.PoweredByRequired(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
