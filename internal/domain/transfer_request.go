package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransferRequestStatus string

const (
	TransferRequestStatusProcessing TransferRequestStatus = "PROCESSING"
	TransferRequestStatusExecuted   TransferRequestStatus = "EXECUTED"
	TransferRequestStatusRejected   TransferRequestStatus = "REJECTED"
)

type TransferRequest struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	IdempotencyKey  string
	RequestHash     string
	Status          TransferRequestStatus
	RejectionReason *string
	TransactionID   *uuid.UUID
	CreatedAt       time.Time
	CompletedAt     *time.Time
}
