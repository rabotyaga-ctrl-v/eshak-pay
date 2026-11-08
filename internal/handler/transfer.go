package handler

import (
	"encoding/json"
	"smartwallet/internal/service"

	"errors"
	"log"
	"net/http"
	"smartwallet/internal/domain"

	"github.com/google/uuid"
)

type TransferHandler struct {
	transferService *service.TransferService
}

func NewTransferHandler(
	transferService *service.TransferService,
) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
	}
}

type transferRequest struct {
	RecipientAccountID uuid.UUID `json:"recipient_account_id"`
	AmountKopecks      int64     `json:"amount_kopecks"`
}

func (h *TransferHandler) Transfer(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	var req transferRequest

	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := h.transferService.Transfer(
		r.Context(),
		service.TransferInput{
			UserID:             userID,
			RecipientAccountID: req.RecipientAccountID,
			AmountKopecks:      req.AmountKopecks,
			IdempotencyKey:     idempotencyKey,
		},
	)

	if err != nil &&
		(result == nil ||
			result.Status != domain.TransferRequestStatusRejected) {

		switch {
		case errors.Is(err, service.ErrInvalidTransferInput):
			http.Error(w, "invalid transfer input", http.StatusBadRequest)
			return

		case errors.Is(err, service.ErrIdempotencyConflict):
			http.Error(w, "idempotency conflict", http.StatusConflict)
			return

		default:
			log.Printf("transfer failed: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if result == nil {
		http.Error(w, "empty transfer result", http.StatusInternalServerError)
		return
	}

	response := struct {
		RequestID       uuid.UUID                    `json:"request_id"`
		Status          domain.TransferRequestStatus `json:"status"`
		TransactionID   *uuid.UUID                   `json:"transaction_id,omitempty"`
		RejectionReason *string                      `json:"rejection_reason,omitempty"`
	}{
		RequestID:       result.RequestID,
		Status:          result.Status,
		TransactionID:   result.TransactionID,
		RejectionReason: result.RejectionReason,
	}

	w.Header().Set("Content-Type", "application/json")

	if result.Status == domain.TransferRequestStatusRejected {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("encode transfer response: %v", err)
	}
}
