package repository

import (
	"context"
	"fmt"

	"smartwallet/internal/domain"

	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountRepository struct {
	db DBTX
}

func NewAccountRepository(db DBTX) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(
	ctx context.Context,
	account *domain.Account,
) error {
	query := `
		INSERT INTO accounts (
			id,
			user_id,
			account_type,
			is_primary,
			currency,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		account.ID,
		account.UserID,
		account.AccountType,
		account.IsPrimary,
		account.Currency,
		account.Status,
	)
	if err != nil {
		return fmt.Errorf("create account: %w", err)
	}

	return nil
}

func (r *AccountRepository) FindByID(
	ctx context.Context,
	accountID uuid.UUID,
) (*domain.Account, error) {
	query := `
		SELECT
			id,
			user_id,
			account_type,
			is_primary,
			currency,
			status,
			created_at,
			closed_at
		FROM accounts
		WHERE id = $1
	`

	var account domain.Account

	err := r.db.QueryRow(ctx, query, accountID).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountType,
		&account.IsPrimary,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.ClosedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("find account by id: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) LockByID(
	ctx context.Context,
	accountID uuid.UUID,
) (*domain.Account, error) {
	query := `
        SELECT
            id,
            user_id,
            account_type,
            is_primary,
            currency,
            status,
            created_at,
            closed_at
        FROM accounts
        WHERE id = $1
        FOR UPDATE
    `

	var account domain.Account

	err := r.db.QueryRow(ctx, query, accountID).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountType,
		&account.IsPrimary,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.ClosedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("lock account by id: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.Account, error) {
	query := `
		SELECT
			id,
			user_id,
			account_type,
			is_primary,
			currency,
			status,
			created_at,
			closed_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("find accounts by user id: %w", err)
	}
	defer rows.Close()

	var accounts []domain.Account

	for rows.Next() {
		var account domain.Account

		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.AccountType,
			&account.IsPrimary,
			&account.Currency,
			&account.Status,
			&account.CreatedAt,
			&account.ClosedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}

	return accounts, nil
}

func (r *AccountRepository) FindPrimaryByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.Account, error) {
	query := `
		SELECT
			id,
			user_id,
			account_type,
			is_primary,
			currency,
			status,
			created_at,
			closed_at
		FROM accounts
		WHERE user_id = $1
		  AND is_primary = TRUE
		  AND status = 'ACTIVE'
	`

	var account domain.Account

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountType,
		&account.IsPrimary,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.ClosedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("find primary account: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) GetBalance(
	ctx context.Context,
	accountID uuid.UUID,
) (int64, error) {
	var balance int64

	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_kopecks), 0)
		FROM postings
		WHERE account_id = $1
	`, accountID).Scan(&balance)

	if err != nil {
		return 0, fmt.Errorf("get account balance: %w", err)
	}

	return balance, nil
}
