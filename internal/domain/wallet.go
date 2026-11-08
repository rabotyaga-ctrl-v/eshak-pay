package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionHistoryItem struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Type          string    `json:"type"`
	AmountKopecks int64     `json:"amount_kopecks"`
	CreatedAt     time.Time `json:"created_at"`
}
