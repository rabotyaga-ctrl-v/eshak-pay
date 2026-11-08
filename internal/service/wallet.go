package service

import (
	"context"
	"errors"
	"fmt"
	"smartwallet/internal/domain"
	"smartwallet/internal/repository"

	"github.com/google/uuid"
)

var ErrInvalidHistoryLimit = errors.New("invalid history limit")

type WalletService struct {
	walletRepo *repository.WalletRepository
}

func NewWalletService(db repository.DBTX) *WalletService {
	return &WalletService{
		walletRepo: repository.NewWalletRepository(db),
	}
}

func (w *WalletService) GetHistory(
	ctx context.Context,
	userID uuid.UUID,
	limit int64,
) ([]domain.TransactionHistoryItem, error) {

	switch limit {
	case 20, 50, 100:
		// допустимое значение
	default:
		return nil, fmt.Errorf("invalid history limit")
	}

	rows, err := w.walletRepo.GetHistory(
		ctx,
		userID,
		limit,
	)
	if err != nil {
		return nil, ErrInvalidHistoryLimit
	}

	return rows, nil
}
