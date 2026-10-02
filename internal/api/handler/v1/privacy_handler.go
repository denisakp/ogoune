package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/ee/license"
	"github.com/denisakp/ogoune/internal/service"
)

// Personal data (spec 094): the privacy page and the export. Both are for the
// signed-in person only; the router mounts them behind RequireJWTOnly.

type privacyInventory interface {
	Inventory(ctx context.Context, userID string) (*domain.PersonalDataInventory, error)
}

type reauthenticator interface {
	Reauthenticate(ctx context.Context, userID, password, code string) error
}

type PrivacyHandler struct {
	privacy privacyInventory
	auth    reauthenticator
	version string
	baseURL string
}

// NewPrivacyHandler takes the install's version and, only when explicitly
// configured, its public base URL -- the two facts the export states about
// what produced it.
func NewPrivacyHandler(privacy privacyInventory, auth reauthenticator, version, baseURL string) *PrivacyHandler {
	return &PrivacyHandler{privacy: privacy, auth: auth, version: version, baseURL: baseURL}
}

// Where each category is managed in the interface (FR-011).
var privacyManagePaths = map[string]string{
	domain.PrivacyCategoryAccount:  "/settings/account",
	domain.PrivacyCategorySessions: "/settings/sessions",
	domain.PrivacyCategoryAPIKeys:  "/api-keys",
	domain.PrivacyCategoryUpdates:  "/incidents",
	domain.PrivacyCategoryChannels: "/notifications",
	domain.PrivacyCategoryReports:  "/reports",
}

// Summary godoc
// @Summary      What this install holds about me
// @Description  Counts per category of personal data held about the signed-in user, the categories that are not personal data, and configurations that could not be checked. Interactive sessions only.
// @Tags         privacy
// @Produce      json
// @Success      200 {object} dtoV1.SingleResponse[dtoV1.PrivacySummaryResponse]
// @Failure      401 {object} dtoV1.ErrorResponse
// @Failure      403 {object} dtoV1.ErrorResponse
// @Security     BearerAuth
// @Router       /v1/me/privacy [get]
func (h *PrivacyHandler) Summary(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	if userID == "" {
		respondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	inv, err := h.privacy.Inventory(r.Context(), userID)
	if err != nil {
		respondPrivacyError(w, r, err)
		return
	}
	counts := inv.Counts()
	cats := make([]dtoV1.PrivacyCategoryResponse, 0, len(domain.PrivacyCategories))
	for _, key := range domain.PrivacyCategories {
		cats = append(cats, dtoV1.PrivacyCategoryResponse{Key: key, Count: counts[key], ManagePath: privacyManagePaths[key]})
	}
	respond(w, http.StatusOK, dtoV1.PrivacySummaryResponse{
		GeneratedAt:       inv.GeneratedAt.UTC().Format(time.RFC3339),
		TwoFactorEnabled:  inv.Account.TwoFactorEnabled,
		Categories:        cats,
		NotPersonalData:   domain.NotPersonalData,
		UncheckedChannels: mapUnchecked(inv.Unchecked),
	})
}

// Export godoc
// @Summary      Download my personal data
// @Description  Re-authenticates (password, plus the two-factor code when enabled) and returns every item of personal data held about the signed-in user as one JSON document. Contains no secret. A failed re-authentication is 422 -- never 401, which would end the session. Rate-limited like sign-in. Interactive sessions only.
// @Tags         privacy
// @Accept       json
// @Produce      json
// @Param        body body dtoV1.PrivacyExportRequest true "Current password, and the two-factor code when enabled"
// @Success      200 {object} dtoV1.PersonalDataExport
// @Failure      401 {object} dtoV1.ErrorResponse
// @Failure      403 {object} dtoV1.ErrorResponse
// @Failure      422 {object} dtoV1.ErrorResponse
// @Failure      429 {object} dtoV1.ErrorResponse
// @Security     BearerAuth
// @Router       /v1/me/privacy/export [post]
func (h *PrivacyHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	if userID == "" {
		respondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	var req dtoV1.PrivacyExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	// 422, not 401: the web client treats any 401 as an expired session and
	// signs the person out. A mistyped password must leave them signed in.
	if err := h.auth.Reauthenticate(r.Context(), userID, req.Password, req.Code); err != nil {
		service.LogExport(userID, false)
		respondError(w, r, http.StatusUnprocessableEntity, "/problems/invalid-credentials", "the password or the two-factor code is not correct")
		return
	}

	inv, err := h.privacy.Inventory(r.Context(), userID)
	if err != nil {
		service.LogExport(userID, false)
		respondPrivacyError(w, r, err)
		return
	}
	doc := h.render(inv)
	service.LogExport(userID, true)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="ogoune-personal-data-%s.json"`, inv.GeneratedAt.UTC().Format("2006-01-02")))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(doc)
}

func respondPrivacyError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, service.ErrResourceNotFound) {
		respondError(w, r, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	respondError(w, r, http.StatusInternalServerError, "INTERNAL", "could not read personal data")
}

func (h *PrivacyHandler) render(inv *domain.PersonalDataInventory) dtoV1.PersonalDataExport {
	doc := dtoV1.PersonalDataExport{
		Format:          "ogoune-personal-data/1",
		GeneratedAt:     inv.GeneratedAt.UTC().Format(time.RFC3339),
		Install:         dtoV1.PersonalDataInstall{Version: h.version, Edition: string(license.Get()), BaseURL: h.baseURL},
		Covers:          domain.PrivacyCategories,
		NotPersonalData: domain.NotPersonalData,
		Account: dtoV1.PersonalDataAccount{
			ID: inv.Account.ID, Email: inv.Account.Email, Name: inv.Account.Name,
			CreatedAt: rfc3339(inv.Account.CreatedAt), LastLoginAt: rfc3339Ptr(inv.Account.LastLoginAt),
			TwoFactorEnabled: inv.Account.TwoFactorEnabled,
		},
		Sessions:             make([]dtoV1.PersonalDataSession, 0, len(inv.Sessions)),
		APIKeys:              make([]dtoV1.PersonalDataAPIKey, 0, len(inv.APIKeys)),
		IncidentUpdates:      make([]dtoV1.PersonalDataIncidentUpdate, 0, len(inv.Updates)),
		NotificationChannels: make([]dtoV1.PersonalDataChannelMatch, 0, len(inv.Channels)),
		Reports:              dtoV1.PersonalDataReports{IsRecipient: inv.Reports.IsRecipient, Sent: make([]dtoV1.PersonalDataReportSent, 0, len(inv.Reports.Sent))},
		UncheckedChannels:    mapUnchecked(inv.Unchecked),
	}
	for _, s := range inv.Sessions {
		doc.Sessions = append(doc.Sessions, dtoV1.PersonalDataSession{
			ID: s.ID, IP: s.IP, Browser: s.Browser, OS: s.OS, Location: s.Location,
			CreatedAt: rfc3339(s.CreatedAt), LastActiveAt: rfc3339(s.LastActiveAt), RevokedAt: rfc3339Ptr(s.RevokedAt),
		})
	}
	for _, k := range inv.APIKeys {
		doc.APIKeys = append(doc.APIKeys, dtoV1.PersonalDataAPIKey{
			ID: k.ID, Name: k.Name, KeyPrefix: k.KeyPrefix, Scope: k.Scope,
			CreatedAt: rfc3339(k.CreatedAt), ExpiresAt: rfc3339Ptr(k.ExpiresAt), LastUsedAt: rfc3339Ptr(k.LastUsedAt),
			LastUsedIP: k.LastUsedIP, Active: k.Active,
		})
	}
	for _, u := range inv.Updates {
		doc.IncidentUpdates = append(doc.IncidentUpdates, dtoV1.PersonalDataIncidentUpdate{
			ID: u.ID, IncidentID: u.IncidentID, Status: u.Status, Message: u.Message, PostedAt: rfc3339(u.PostedAt), Public: u.Public,
		})
	}
	for _, c := range inv.Channels {
		doc.NotificationChannels = append(doc.NotificationChannels, dtoV1.PersonalDataChannelMatch{ID: c.ChannelID, Name: c.ChannelName, Type: c.ChannelType, Fields: c.Fields})
	}
	for _, rs := range inv.Reports.Sent {
		doc.Reports.Sent = append(doc.Reports.Sent, dtoV1.PersonalDataReportSent{Period: rs.Period, Status: rs.Status, SentAt: rfc3339Ptr(rs.SentAt)})
	}
	return doc
}

func mapUnchecked(in []domain.UncheckedChannel) []dtoV1.PrivacyUncheckedChannel {
	out := make([]dtoV1.PrivacyUncheckedChannel, 0, len(in))
	for _, c := range in {
		out = append(out, dtoV1.PrivacyUncheckedChannel{ID: c.ChannelID, Name: c.ChannelName, Type: c.ChannelType})
	}
	return out
}

func rfc3339(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func rfc3339Ptr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := rfc3339(*t)
	return &s
}
