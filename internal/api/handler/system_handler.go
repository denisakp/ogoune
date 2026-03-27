package handler

import (
	"net/http"
	"os"

	"github.com/denisakp/ogoune/internal/api/response"
	"github.com/denisakp/ogoune/internal/ee/license"
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
		version = "1.0.0"
	}

	response.JSON(w, http.StatusOK, editionResponse{
		Edition: string(license.Get()),
		Version: version,
	})
}
