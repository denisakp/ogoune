package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/service"
)

// Erasure of a person's data on request (spec 095). Interactive sessions
// only; the erasure itself re-authenticates the operator. A failed
// confirmation is 422 -- never 401, which would end the operator's session.

// ErasureAccounts godoc
// @Summary      Other accounts that can be erased
// @Description  Every account but the signed-in one, with its address and last sign-in. Interactive sessions only.
// @Tags         privacy
// @Produce      json
// @Success      200 {object} dtoV1.SingleResponse[[]dtoV1.ErasureAccountResponse]
// @Failure      401 {object} dtoV1.ErrorResponse
// @Failure      403 {object} dtoV1.ErrorResponse
// @Security     BearerAuth
// @Router       /me/privacy/accounts [get]
func (h *PrivacyHandler) ErasureAccounts(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	if userID == "" {
		respondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	accounts, err := h.erasure.OtherAccounts(r.Context(), userID)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL", "could not list accounts")
		return
	}
	out := make([]dtoV1.ErasureAccountResponse, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, dtoV1.ErasureAccountResponse{ID: a.ID, Email: a.Email, LastLoginAt: rfc3339Ptr(a.LastLoginAt)})
	}
	respond(w, http.StatusOK, out)
}

// PreviewErasure godoc
// @Summary      Preview an erasure
// @Description  What erasing an address, or a former account, would remove, keep and leave for manual review. Changes nothing. The address travels in the body, never in a URL. Interactive sessions only.
// @Tags         privacy
// @Accept       json
// @Produce      json
// @Param        body body dtoV1.ErasurePreviewRequest true "An address, or a former account"
// @Success      200 {object} dtoV1.SingleResponse[dtoV1.ErasurePreviewResponse]
// @Failure      400 {object} dtoV1.ErrorResponse
// @Failure      401 {object} dtoV1.ErrorResponse
// @Failure      403 {object} dtoV1.ErrorResponse
// @Failure      404 {object} dtoV1.ErrorResponse
// @Failure      422 {object} dtoV1.ErrorResponse
// @Security     BearerAuth
// @Router       /me/privacy/erasure/preview [post]
func (h *PrivacyHandler) PreviewErasure(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	if userID == "" {
		respondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	var req dtoV1.ErasurePreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	p, err := h.erasure.Preview(r.Context(), userID, service.ErasureSubject{Email: req.Email, AccountID: req.AccountID})
	if err != nil {
		respondErasureError(w, r, err)
		return
	}
	respond(w, http.StatusOK, mapErasurePreview(p))
}

// Erase godoc
// @Summary      Erase a person's data
// @Description  Removes the address from every place it was found (or erases a former account), in one transaction: all or nothing. Requires the address typed again, the current password and, when enabled, the two-factor code; any of them wrong is one 422, never 401. A notification channel changed meanwhile is 409, nothing changed. Rate-limited like sign-in. Interactive sessions only.
// @Tags         privacy
// @Accept       json
// @Produce      json
// @Param        body body dtoV1.ErasureRequest true "Subject, typed address and re-authentication"
// @Success      200 {object} dtoV1.SingleResponse[dtoV1.ErasureResultResponse]
// @Failure      400 {object} dtoV1.ErrorResponse
// @Failure      401 {object} dtoV1.ErrorResponse
// @Failure      403 {object} dtoV1.ErrorResponse
// @Failure      404 {object} dtoV1.ErrorResponse
// @Failure      409 {object} dtoV1.ErrorResponse
// @Failure      422 {object} dtoV1.ErrorResponse
// @Failure      429 {object} dtoV1.ErrorResponse
// @Security     BearerAuth
// @Router       /me/privacy/erasure [post]
func (h *PrivacyHandler) Erase(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	if userID == "" {
		respondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	var req dtoV1.ErasureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	verify := func(ctx context.Context) error {
		return h.auth.Reauthenticate(ctx, userID, req.Password, req.Code)
	}
	res, err := h.erasure.Erase(r.Context(), userID, service.ErasureSubject{Email: req.Email, AccountID: req.AccountID}, req.ConfirmEmail, verify)
	if err != nil {
		respondErasureError(w, r, err)
		return
	}
	respond(w, http.StatusOK, mapErasureResult(res))
}

// respondErasureError maps erasure errors. Messages never carry the address.
func respondErasureError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrErasureNotConfirmed):
		respondError(w, r, http.StatusUnprocessableEntity, "/problems/invalid-credentials", "the typed address, the password or the two-factor code is not correct")
	case errors.Is(err, service.ErrCannotEraseSelf):
		respondError(w, r, http.StatusUnprocessableEntity, "/problems/cannot-erase-self", "your own account and address cannot be erased here")
	case errors.Is(err, service.ErrLastAccount):
		respondError(w, r, http.StatusUnprocessableEntity, "/problems/last-account", "the erasure would leave no account able to sign in")
	case errors.Is(err, service.ErrValidationFailed):
		respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "give either an email address or an account")
	case errors.Is(err, service.ErrResourceNotFound):
		respondError(w, r, http.StatusNotFound, "NOT_FOUND", "account not found")
	case errors.Is(err, service.ErrErasureConflict):
		respondError(w, r, http.StatusConflict, "/problems/erasure-conflict", "a notification channel changed during the erasure; nothing was changed, try again")
	default:
		respondError(w, r, http.StatusInternalServerError, "INTERNAL", "the erasure failed; nothing was changed")
	}
}

func mapErasurePreview(p *domain.ErasurePreview) dtoV1.ErasurePreviewResponse {
	out := dtoV1.ErasurePreviewResponse{
		Kind:               string(p.Kind),
		Channels:           mapErasureHits(p.Channels),
		ReportRecipient:    p.ReportRecipient,
		ReportsSent:        p.ReportsSent,
		Sessions:           p.Sessions,
		APIKeys:            p.APIKeys,
		UpdatesUnlinked:    p.UpdatesUnlinked,
		ManualReview:       mapErasureManual(p.ManualReview),
		PreviouslyErasedAt: rfc3339Ptr(p.PreviouslyErased),
		TransportChange:    mapTransportChange(p.TransportChange),
	}
	if p.Account != nil {
		out.Account = &dtoV1.ErasureAccountResponse{ID: p.Account.ID, Email: p.Account.Email, LastLoginAt: rfc3339Ptr(p.Account.LastLoginAt)}
	}
	return out
}

func mapErasureResult(res *domain.ErasureResult) dtoV1.ErasureResultResponse {
	return dtoV1.ErasureResultResponse{
		RecordID:               res.RecordID,
		Changes:                res.Changes,
		DisabledChannels:       mapErasureHits(res.DisabledChannels),
		ManualReview:           mapErasureManual(res.ManualReview),
		TransportChange:        mapTransportChange(res.TransportChange),
		ReportRecipientCleared: res.ReportRecipientCleared,
	}
}

func mapErasureHits(in []domain.ErasureChannelHit) []dtoV1.ErasureChannelHitResponse {
	out := make([]dtoV1.ErasureChannelHitResponse, 0, len(in))
	for _, h := range in {
		out = append(out, dtoV1.ErasureChannelHitResponse{ID: h.ID, Name: h.Name, Type: h.Type, Fields: h.Fields, WillDisable: h.WillDisable})
	}
	return out
}

func mapErasureManual(in []domain.ErasureManualItem) []dtoV1.ErasureManualItemResponse {
	out := make([]dtoV1.ErasureManualItemResponse, 0, len(in))
	for _, m := range in {
		out = append(out, dtoV1.ErasureManualItemResponse{ChannelID: m.ChannelID, ChannelName: m.ChannelName, ChannelType: m.ChannelType, Reason: m.Reason})
	}
	return out
}

func mapTransportChange(tc *domain.ErasureTransportChange) *dtoV1.ErasureTransportChangeResponse {
	if tc == nil {
		return nil
	}
	out := &dtoV1.ErasureTransportChangeResponse{FromChannel: dtoV1.ErasureChannelRef{ID: tc.FromChannelID, Name: tc.FromChannelName}}
	if tc.ToChannelID != "" {
		out.ToChannel = &dtoV1.ErasureChannelRef{ID: tc.ToChannelID, Name: tc.ToChannelName}
	}
	return out
}
