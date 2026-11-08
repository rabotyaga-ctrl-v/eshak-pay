package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeInitialCredit TransactionType = "INITIAL_CREDIT"
	TransactionTypeTransfer      TransactionType = "TRANSFER"
)

type Transaction struct {
	ID                   uuid.UUID
	InitiatedByUserID    *uuid.UUID
	SourceAccountID      uuid.UUID
	DestinationAccountID uuid.UUID
	Type                 TransactionType
	AmountKopecks        int64
	CreatedAt            time.Time
}
