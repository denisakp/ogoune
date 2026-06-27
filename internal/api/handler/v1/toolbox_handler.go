package v1

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/service"
)

// ToolboxV1ServiceInterface is the slice of *service.ToolboxService used by the handler.
type ToolboxV1ServiceInterface interface {
	DNS(ctx context.Context, q service.ToolboxDNSQuery) (service.ToolboxDNSResult, error)
	PortScan(ctx context.Context, q service.ToolboxPortScanQuery) (service.ToolboxPortScanResult, error)
	SSL(ctx context.Context, q service.ToolboxSSLQuery) (service.ToolboxSSLResult, error)
	WHOIS(ctx context.Context, domain string) (service.ToolboxWhoisResult, error)
}

// ToolboxHandler exposes the one-shot network tools under /api/v1/toolbox.
type ToolboxHandler struct {
	service ToolboxV1ServiceInterface
}

// NewToolboxHandler creates a new ToolboxHandler.
func NewToolboxHandler(svc ToolboxV1ServiceInterface) *ToolboxHandler {
	return &ToolboxHandler{service: svc}
}

// DNS handles POST /api/v1/toolbox/dns.
//
// @Summary  Run a one-off DNS lookup
// @Tags     toolbox
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dtoV1.DNSLookupRequest true "DNS lookup request"
// @Success  200 {object} dtoV1.SingleResponse[dtoV1.DNSLookupResponse]
// @Failure  400 {object} dtoV1.ErrorResponse
// @Failure  403 {object} dtoV1.ErrorResponse
// @Failure  429 {object} dtoV1.ErrorResponse
// @Router   /toolbox/dns [post]
func (h *ToolboxHandler) DNS(w http.ResponseWriter, r *http.Request) {
	var req dtoV1.DNSLookupRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	res, err := h.service.DNS(r.Context(), service.ToolboxDNSQuery{
		Domain:         req.Domain,
		RecordTypes:    req.RecordTypes,
		Resolver:       req.Resolver,
		CustomResolver: req.CustomResolver,
	})
	if err != nil {
		h.audit(r.Context(), "toolbox.dns", req.Domain, outcomeFromErr(err))
		h.mapToolboxError(w, r, err)
		return
	}
	records := make([]dtoV1.DNSRecord, 0, len(res.Records))
	for _, rec := range res.Records {
		records = append(records, dtoV1.DNSRecord{Type: rec.Type, Value: rec.Value, TTL: rec.TTL})
	}
	h.audit(r.Context(), "toolbox.dns", req.Domain, "ok")
	respond(w, http.StatusOK, dtoV1.DNSLookupResponse{
		Records:      records,
		QueryMs:      res.QueryMs,
		ResolverUsed: res.ResolverUsed,
	})
}

// PortScan handles POST /api/v1/toolbox/port-scan.
//
// @Summary  Scan ports on a registered monitor host (rate-limited 5/min)
// @Tags     toolbox
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dtoV1.PortScanRequest true "Port scan request"
// @Success  200 {object} dtoV1.SingleResponse[dtoV1.PortScanResponse]
// @Failure  400 {object} dtoV1.ErrorResponse
// @Failure  403 {object} dtoV1.ErrorResponse
// @Failure  429 {object} dtoV1.ErrorResponse
// @Router   /toolbox/port-scan [post]
func (h *ToolboxHandler) PortScan(w http.ResponseWriter, r *http.Request) {
	var req dtoV1.PortScanRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	res, err := h.service.PortScan(r.Context(), service.ToolboxPortScanQuery{
		Target:    req.Target,
		Ports:     req.Ports,
		TimeoutMs: req.TimeoutMs,
	})
	if err != nil {
		h.audit(r.Context(), "toolbox.port-scan", req.Target, outcomeFromErr(err))
		h.mapToolboxError(w, r, err)
		return
	}
	results := make([]dtoV1.PortResult, 0, len(res.Results))
	for _, pr := range res.Results {
		results = append(results, dtoV1.PortResult{Port: pr.Port, Service: pr.Service, Status: pr.Status, Banner: pr.Banner})
	}
	h.audit(r.Context(), "toolbox.port-scan", req.Target, "ok")
	respond(w, http.StatusOK, dtoV1.PortScanResponse{
		Results:      results,
		OpenCount:    res.OpenCount,
		ScannedCount: res.ScannedCount,
	})
}

// SSL handles POST /api/v1/toolbox/ssl-check.
//
// @Summary  Inspect a TLS certificate
// @Tags     toolbox
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dtoV1.SSLCheckRequest true "SSL check request"
// @Success  200 {object} dtoV1.SingleResponse[dtoV1.SSLCheckResponse]
// @Failure  400 {object} dtoV1.ErrorResponse
// @Failure  422 {object} dtoV1.ErrorResponse
// @Router   /toolbox/ssl-check [post]
func (h *ToolboxHandler) SSL(w http.ResponseWriter, r *http.Request) {
	var req dtoV1.SSLCheckRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	res, err := h.service.SSL(r.Context(), service.ToolboxSSLQuery{Domain: req.Domain, Port: req.Port})
	if err != nil {
		h.audit(r.Context(), "toolbox.ssl-check", req.Domain, outcomeFromErr(err))
		h.mapToolboxError(w, r, err)
		return
	}
	vulns := make([]dtoV1.SSLVulnCheck, 0, len(res.Vulnerabilities))
	for _, v := range res.Vulnerabilities {
		vulns = append(vulns, dtoV1.SSLVulnCheck{Name: v.Name, Status: v.Status})
	}
	h.audit(r.Context(), "toolbox.ssl-check", req.Domain, "ok")
	respond(w, http.StatusOK, dtoV1.SSLCheckResponse{
		Certificate: dtoV1.SSLCertificate{
			Subject:   res.Certificate.Subject,
			Issuer:    res.Certificate.Issuer,
			ValidFrom: res.Certificate.ValidFrom.UTC().Format(time.RFC3339),
			ValidTo:   res.Certificate.ValidTo.UTC().Format(time.RFC3339),
			Cipher:    res.Certificate.Cipher,
			SANs:      res.Certificate.SANs,
			Chain:     res.Certificate.Chain,
		},
		DaysToExpiry:    res.DaysToExpiry,
		ExpiringSoon:    res.ExpiringSoon,
		Vulnerabilities: vulns,
	})
}

// WHOIS handles POST /api/v1/toolbox/whois.
//
// @Summary  Look up domain registration data
// @Tags     toolbox
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dtoV1.WhoisRequest true "WHOIS request"
// @Success  200 {object} dtoV1.SingleResponse[dtoV1.WhoisResponse]
// @Failure  400 {object} dtoV1.ErrorResponse
// @Failure  422 {object} dtoV1.ErrorResponse
// @Router   /toolbox/whois [post]
func (h *ToolboxHandler) WHOIS(w http.ResponseWriter, r *http.Request) {
	var req dtoV1.WhoisRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	res, err := h.service.WHOIS(r.Context(), req.Domain)
	if err != nil {
		h.audit(r.Context(), "toolbox.whois", req.Domain, outcomeFromErr(err))
		h.mapToolboxError(w, r, err)
		return
	}
	h.audit(r.Context(), "toolbox.whois", req.Domain, "ok")
	respond(w, http.StatusOK, dtoV1.WhoisResponse{
		Registrar:    res.Registrar,
		RegisteredAt: res.RegisteredAt,
		UpdatedAt:    res.UpdatedAt,
		ExpiresAt:    res.ExpiresAt,
		DaysToExpiry: res.DaysToExpiry,
		Status:       res.Status,
		Privacy:      res.Privacy,
		DNSSEC:       res.DNSSEC,
		Nameservers:  res.Nameservers,
	})
}

// decodeJSON decodes the request body, writing a 400 on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respondError(w, r, http.StatusBadRequest, "VALIDATION_FAILED", "invalid JSON body")
		return false
	}
	return true
}

// mapToolboxError maps a toolbox service error to an RFC-7807 response.
func (h *ToolboxHandler) mapToolboxError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrToolboxValidation):
		respondError(w, r, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
	case errors.Is(err, service.ErrToolboxTargetNotRegistered):
		respondError(w, r, http.StatusForbidden, "TARGET_NOT_REGISTERED", "target must be a registered monitor host")
	case errors.Is(err, service.ErrToolboxTargetBlocked):
		respondError(w, r, http.StatusForbidden, "TARGET_BLOCKED", "target resolves to a blocked address")
	case errors.Is(err, service.ErrToolboxCertUnavailable):
		respondError(w, r, http.StatusUnprocessableEntity, "CERT_UNAVAILABLE", "certificate could not be retrieved")
	case errors.Is(err, service.ErrToolboxWhoisNoData):
		respondError(w, r, http.StatusUnprocessableEntity, "WHOIS_NO_DATA", "no registration data for domain")
	default:
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "toolbox request failed")
	}
}

// outcomeFromErr maps an error to a stable audit outcome string.
func outcomeFromErr(err error) string {
	switch {
	case errors.Is(err, service.ErrToolboxValidation):
		return "validation_failed"
	case errors.Is(err, service.ErrToolboxTargetNotRegistered):
		return "refused_unregistered"
	case errors.Is(err, service.ErrToolboxTargetBlocked):
		return "blocked_ssrf"
	case errors.Is(err, service.ErrToolboxCertUnavailable):
		return "cert_unavailable"
	case errors.Is(err, service.ErrToolboxWhoisNoData):
		return "whois_no_data"
	default:
		return "error"
	}
}

// audit emits a structured log entry for each toolbox run (FR-017). No table exists.
func (h *ToolboxHandler) audit(ctx context.Context, action, target, outcome string) {
	userID, _ := ctx.Value("user_id").(string)
	slog.Default().LogAttrs(ctx, slog.LevelInfo, "audit",
		slog.String("category", "toolbox"),
		slog.String("action", action),
		slog.String("user_id", userID),
		slog.String("target", target),
		slog.String("outcome", outcome),
	)
}
