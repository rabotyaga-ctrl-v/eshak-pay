package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"smartwallet/internal/domain"
	"smartwallet/internal/service"
	"strconv"
)

type WalletHandler struct {
	walletService *service.WalletService
}

func NewWalletHandler(
	walletService *service.WalletService,
) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

func (h *WalletHandler) GetHistory(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")

	var limit int64 = 20

	if limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}

		limit = parsedLimit
	}

	history, err := h.walletService.GetHistory(
		r.Context(),
		userID,
		limit,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidHistoryLimit) {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}

		log.Printf("get history failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := struct {
		Transactions []domain.TransactionHistoryItem `json:"transactions"`
	}{
		Transactions: history,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("encode history response: %v", err)
	}
}
