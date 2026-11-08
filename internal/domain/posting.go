package domain

import (
	"time"

	"github.com/google/uuid"
)

type Posting struct {
	ID            uuid.UUID
	TransactionID uuid.UUID
	AccountID     uuid.UUID
	AmountKopecks int64
	CreatedAt     time.Time
}
