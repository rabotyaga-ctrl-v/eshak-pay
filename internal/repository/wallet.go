package repository

import (
	"context"
	"fmt"

	"smartwallet/internal/domain"

	"github.com/google/uuid"
)

type WalletRepository struct {
	db DBTX
}

func NewWalletRepository(db DBTX) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetHistory(
	ctx context.Context,
	userID uuid.UUID,
	limit int64,
) ([]domain.TransactionHistoryItem, error) {

	rows, err := r.db.Query(ctx, `
		SELECT * FROM get_transaction_history($1, $2)
	`,
		userID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get history transactions: %w", err)
	}
	defer rows.Close()

	var history []domain.TransactionHistoryItem

	for rows.Next() {
		var item domain.TransactionHistoryItem

		err := rows.Scan(
			&item.TransactionID,
			&item.Type,
			&item.AmountKopecks,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan history transactions: %w", err)
		}

		history = append(history, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history transactions: %w", err)
	}

	return history, nil
}
