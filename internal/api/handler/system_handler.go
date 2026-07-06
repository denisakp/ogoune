package handler

import (
	"net/http"
	"os"

	"github.com/denisakp/ogoune/internal/api/response"
	"github.com/denisakp/ogoune/internal/config"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/ee/license"
	icmppkg "github.com/denisakp/ogoune/internal/icmp"
)

type SystemHandler struct{}

type editionResponse struct {
	Edition string `json:"edition"`
	Version string `json:"version"`
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (h *SystemHandler) GetEdition(w http.ResponseWriter, r *http.Request) {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0-beta"
	}

	response.JSON(w, http.StatusOK, editionResponse{
		Edition: string(license.Get()),
		Version: version,
	})
}

func (h *SystemHandler) GetCapabilities(w http.ResponseWriter, r *http.Request) {
	cfg := config.Load()
	capability := icmppkg.Detect()

	response.JSON(w, http.StatusOK, dto.SystemCapabilitiesResponse{
		ICMP: dto.ICMPAvailabilityState{
			Enabled:             cfg.EnableICMP,
			CapabilityAvailable: capability.Available,
			Reason:              capability.Reason,
		},
	})
}
