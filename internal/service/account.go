package service

import (
	"context"
	"fmt"

	"smartwallet/internal/domain"
	"smartwallet/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	systemAccountID = "00000000-0000-0000-0000-000000000001"
	initialBalance  = int64(5_000_000) // 50 000 RUB в копейках
)

type AccountService struct {
	accountRepo *repository.AccountRepository
}

func NewAccountService(db repository.DBTX) *AccountService {
	return &AccountService{
		accountRepo: repository.NewAccountRepository(db),
	}
}

func (s *AccountService) CreatePrimaryAccount(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
) error {

	accountRepo := repository.NewAccountRepository(tx)
	transactionRepo := repository.NewTransactionRepository(tx)
	postingRepo := repository.NewPostingRepository(tx)

	accountID := uuid.New()

	account := &domain.Account{
		ID:          accountID,
		UserID:      &userID,
		AccountType: domain.AccountTypeUser,
		IsPrimary:   true,
		Currency:    "RUB",
		Status:      domain.AccountStatusActive,
	}

	if err := accountRepo.Create(ctx, account); err != nil {
		return fmt.Errorf("create primary account: %w", err)
	}

	transactionID := uuid.New()
	systemID := uuid.MustParse(systemAccountID)

	transaction := &domain.Transaction{
		ID:                   transactionID,
		SourceAccountID:      systemID,
		DestinationAccountID: accountID,
		Type:                 domain.TransactionTypeInitialCredit,
		AmountKopecks:        initialBalance,
	}

	if err := transactionRepo.Create(ctx, transaction); err != nil {
		return fmt.Errorf("create initial credit transaction: %w", err)
	}

	systemPosting := &domain.Posting{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     systemID,
		AmountKopecks: -initialBalance,
	}

	if err := postingRepo.Create(ctx, systemPosting); err != nil {
		return fmt.Errorf("create system posting: %w", err)
	}

	userPosting := &domain.Posting{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     accountID,
		AmountKopecks: initialBalance,
	}

	if err := postingRepo.Create(ctx, userPosting); err != nil {
		return fmt.Errorf("create user posting: %w", err)
	}

	return nil
}

func (s *AccountService) GetPrimaryAccountBalance(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.Account, int64, error) {
	account, err := s.accountRepo.FindPrimaryByUserID(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("find primary account: %w", err)
	}

	balance, err := s.accountRepo.GetBalance(ctx, account.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("get primary account balance: %w", err)
	}

	return account, balance, nil
}
