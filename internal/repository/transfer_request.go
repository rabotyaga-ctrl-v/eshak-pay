package repository

import (
	"context"
	"fmt"

	"errors"
	"smartwallet/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrTransferRequestNotFound = errors.New("transfer request not found")

type TransferRequestRepository struct {
	db DBTX
}

func NewTransferRequestRepository(db DBTX) *TransferRequestRepository {
	return &TransferRequestRepository{db: db}
}

func (r *TransferRequestRepository) Create(
	ctx context.Context,
	request *domain.TransferRequest,
) (bool, error) {
	tag, err := r.db.Exec(ctx, `
        INSERT INTO transfer_requests (
            id,
            user_id,
            idempotency_key,
            request_hash,
            status
        )
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (user_id, idempotency_key) DO NOTHING
    `,
		request.ID,
		request.UserID,
		request.IdempotencyKey,
		request.RequestHash,
		request.Status,
	)
	if err != nil {
		return false, fmt.Errorf("create transfer request: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *TransferRequestRepository) FindByKey(
	ctx context.Context,
	userID uuid.UUID,
	idempotencyKey string,
) (*domain.TransferRequest, error) {
	query := `
        SELECT
            id,
            user_id,
            idempotency_key,
            request_hash,
            status,
            rejection_reason,
            transaction_id,
            created_at,
            completed_at
        FROM transfer_requests
        WHERE user_id = $1
          AND idempotency_key = $2
    `

	var request domain.TransferRequest

	err := r.db.QueryRow(ctx, query, userID, idempotencyKey).Scan(
		&request.ID,
		&request.UserID,
		&request.IdempotencyKey,
		&request.RequestHash,
		&request.Status,
		&request.RejectionReason,
		&request.TransactionID,
		&request.CreatedAt,
		&request.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransferRequestNotFound
		}

		return nil, fmt.Errorf("find transfer request by key: %w", err)
	}

	return &request, nil
}

func (r *TransferRequestRepository) Complete(
	ctx context.Context,
	requestID uuid.UUID,
	status domain.TransferRequestStatus,
	transactionID *uuid.UUID,
	rejectionReason *string,
) error {
	tag, err := r.db.Exec(ctx, `
        UPDATE transfer_requests
        SET
            status = $2,
            transaction_id = $3,
            rejection_reason = $4,
            completed_at = NOW()
        WHERE id = $1
          AND status = 'PROCESSING'
    `,
		requestID,
		status,
		transactionID,
		rejectionReason,
	)
	if err != nil {
		return fmt.Errorf("complete transfer request: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("complete transfer request: request not found or already completed")
	}

	return nil
}
