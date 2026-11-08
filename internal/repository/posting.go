package repository

import (
	"context"
	"fmt"

	"smartwallet/internal/domain"
)

type PostingRepository struct {
	db DBTX
}

func NewPostingRepository(db DBTX) *PostingRepository {
	return &PostingRepository{db: db}
}

func (r *PostingRepository) Create(ctx context.Context, posting *domain.Posting) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO postings (
			id,
			transaction_id,
			account_id,
			amount_kopecks
		)
		VALUES ($1, $2, $3, $4)
	`,
		posting.ID,
		posting.TransactionID,
		posting.AccountID,
		posting.AmountKopecks,
	)
	if err != nil {
		return fmt.Errorf("create posting: %w", err)
	}

	return nil
}
