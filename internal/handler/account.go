package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"smartwallet/internal/repository"
	"smartwallet/internal/service"

	"github.com/google/uuid"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(
	accountService *service.AccountService,
) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

func (h *AccountHandler) GetPrimaryAccount(
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

	account, balance, err := h.accountService.GetPrimaryAccountBalance(
		r.Context(),
		userID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}

		log.Printf("get primary account: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		ID             uuid.UUID `json:"id"`
		Currency       string    `json:"currency"`
		Status         string    `json:"status"`
		BalanceKopecks int64     `json:"balance_kopecks"`
	}{
		ID:             account.ID,
		Currency:       account.Currency,
		Status:         string(account.Status),
		BalanceKopecks: balance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("encode account response: %v", err)
	}
}
