package repository

import (
	"context"
	"fmt"

	"smartwallet/internal/domain"
)

type TransactionRepository struct {
	db DBTX
}

func NewTransactionRepository(db DBTX) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO transactions (
			id,
			initiated_by_user_id,
			source_account_id,
			destination_account_id,
			type,
			amount_kopecks
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		tx.ID,
		tx.InitiatedByUserID,
		tx.SourceAccountID,
		tx.DestinationAccountID,
		tx.Type,
		tx.AmountKopecks,
	)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	return nil
}
