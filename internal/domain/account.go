package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountTypeUser   AccountType = "USER"
	AccountTypeSystem AccountType = "SYSTEM"
)

type AccountStatus string

const (
	AccountStatusActive AccountStatus = "ACTIVE"
	AccountStatusClosed AccountStatus = "CLOSED"
)

type Account struct {
	ID          uuid.UUID
	UserID      *uuid.UUID
	AccountType AccountType
	IsPrimary   bool
	Currency    string
	Status      AccountStatus
	CreatedAt   time.Time
	ClosedAt    *time.Time
}
