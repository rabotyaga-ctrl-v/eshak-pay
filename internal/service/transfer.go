package service

import (
	"errors"

	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"smartwallet/internal/domain"
	"strings"

	"context"
	"smartwallet/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidTransferInput = errors.New("invalid transfer input")
	ErrTransferToSelf       = errors.New("cannot transfer to the same account")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidRecipient     = errors.New("invalid recipient account")
	ErrIdempotencyConflict  = errors.New("idempotency key already used with different request")
)

type TransferService struct {
	db *pgxpool.Pool
}

type TransferInput struct {
	UserID             uuid.UUID
	RecipientAccountID uuid.UUID
	AmountKopecks      int64
	IdempotencyKey     string
}

type TransferResult struct {
	RequestID       uuid.UUID
	Status          domain.TransferRequestStatus
	TransactionID   *uuid.UUID
	RejectionReason *string
}

func NewTransferService(db *pgxpool.Pool) *TransferService {
	return &TransferService{
		db: db,
	}
}

func calculateRequestHash(input TransferInput) string {
	payload := fmt.Sprintf(
		"transfer:v1:%s:%s:%d",
		input.UserID,
		input.RecipientAccountID,
		input.AmountKopecks,
	)

	hash := sha256.Sum256([]byte(payload))

	return hex.EncodeToString(hash[:])
}

func validateTransferInput(input *TransferInput) error {
	if input.UserID == uuid.Nil ||
		input.RecipientAccountID == uuid.Nil ||
		input.AmountKopecks <= 0 {
		return ErrInvalidTransferInput
	}

	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)

	if len(input.IdempotencyKey) == 0 ||
		len(input.IdempotencyKey) > 128 {
		return ErrInvalidTransferInput
	}

	return nil
}

func (s *TransferService) getExistingResult(
	ctx context.Context,
	repo *repository.TransferRequestRepository,
	userID uuid.UUID,
	idempotencyKey string,
	requestHash string,
) (*TransferResult, error) {
	request, err := repo.FindByKey(ctx, userID, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("find existing transfer request: %w", err)
	}

	if request.RequestHash != requestHash {
		return nil, ErrIdempotencyConflict
	}

	if request.Status == domain.TransferRequestStatusProcessing {
		return nil, fmt.Errorf("transfer request is still processing")
	}

	return &TransferResult{
		RequestID:       request.ID,
		Status:          request.Status,
		TransactionID:   request.TransactionID,
		RejectionReason: request.RejectionReason,
	}, nil
}

func (s *TransferService) registerTransferRequest(
	ctx context.Context,
	repo *repository.TransferRequestRepository,
	input TransferInput,
	requestHash string,
) (*domain.TransferRequest, *TransferResult, error) {
	request := &domain.TransferRequest{
		ID:             uuid.New(),
		UserID:         input.UserID,
		IdempotencyKey: input.IdempotencyKey,
		RequestHash:    requestHash,
		Status:         domain.TransferRequestStatusProcessing,
	}

	created, err := repo.Create(ctx, request)
	if err != nil {
		return nil, nil, fmt.Errorf("register transfer request: %w", err)
	}

	if !created {
		result, err := s.getExistingResult(
			ctx,
			repo,
			input.UserID,
			input.IdempotencyKey,
			requestHash,
		)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}

	return request, nil, nil
}

func lockTransferAccounts(
	ctx context.Context,
	accountRepo *repository.AccountRepository,
	sourceID uuid.UUID,
	destinationID uuid.UUID,
) (*domain.Account, *domain.Account, error) {
	var firstID, secondID uuid.UUID

	if sourceID.String() < destinationID.String() {
		firstID, secondID = sourceID, destinationID
	} else {
		firstID, secondID = destinationID, sourceID
	}

	first, err := accountRepo.LockByID(ctx, firstID)
	if err != nil {
		return nil, nil, fmt.Errorf("lock first account: %w", err)
	}

	second, err := accountRepo.LockByID(ctx, secondID)
	if err != nil {
		return nil, nil, fmt.Errorf("lock second account: %w", err)
	}

	if first.ID == sourceID {
		return first, second, nil
	}

	return second, first, nil
}

func validateTransferAccounts(
	source *domain.Account,
	destination *domain.Account,
	userID uuid.UUID,
) error {
	if source.UserID == nil || *source.UserID != userID {
		return ErrInvalidTransferInput
	}

	if source.AccountType != domain.AccountTypeUser ||
		destination.AccountType != domain.AccountTypeUser {
		return ErrInvalidRecipient
	}

	if source.Status != domain.AccountStatusActive ||
		destination.Status != domain.AccountStatusActive {
		return ErrInvalidRecipient
	}

	if source.ID == destination.ID {
		return ErrTransferToSelf
	}

	if source.Currency != destination.Currency {
		return ErrInvalidRecipient
	}

	if destination.UserID == nil {
		return ErrInvalidRecipient
	}

	return nil
}

func validateTransferBalance(
	balance int64,
	amount int64,
) error {
	if amount <= 0 {
		return ErrInvalidTransferInput
	}

	if balance < amount {
		return ErrInsufficientFunds
	}

	return nil
}

func (s *TransferService) rejectTransfer(
	ctx context.Context,
	tx pgx.Tx,
	requestID uuid.UUID,
	reason error,
) (*TransferResult, error) {
	reasonText := reason.Error()

	requestRepo := repository.NewTransferRequestRepository(tx)

	if err := requestRepo.Complete(
		ctx,
		requestID,
		domain.TransferRequestStatusRejected,
		nil,
		&reasonText,
	); err != nil {
		return nil, fmt.Errorf("reject transfer request: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit rejected transfer: %w", err)
	}

	return &TransferResult{
		RequestID:       requestID,
		Status:          domain.TransferRequestStatusRejected,
		RejectionReason: &reasonText,
	}, reason
}

func createTransferRecords(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
	sourceID uuid.UUID,
	destinationID uuid.UUID,
	amount int64,
) (uuid.UUID, error) {
	transactionID := uuid.New()

	transactionRepo := repository.NewTransactionRepository(tx)
	postingRepo := repository.NewPostingRepository(tx)

	transaction := &domain.Transaction{
		ID:                   transactionID,
		InitiatedByUserID:    &userID,
		SourceAccountID:      sourceID,
		DestinationAccountID: destinationID,
		Type:                 domain.TransactionTypeTransfer,
		AmountKopecks:        amount,
	}

	if err := transactionRepo.Create(ctx, transaction); err != nil {
		return uuid.Nil, fmt.Errorf("create transfer transaction: %w", err)
	}

	sourcePosting := &domain.Posting{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     sourceID,
		AmountKopecks: -amount,
	}

	if err := postingRepo.Create(ctx, sourcePosting); err != nil {
		return uuid.Nil, fmt.Errorf("create source posting: %w", err)
	}

	destinationPosting := &domain.Posting{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     destinationID,
		AmountKopecks: amount,
	}

	if err := postingRepo.Create(ctx, destinationPosting); err != nil {
		return uuid.Nil, fmt.Errorf("create destination posting: %w", err)
	}

	return transactionID, nil
}

func (s *TransferService) Transfer(
	ctx context.Context,
	input TransferInput,
) (*TransferResult, error) {
	if err := validateTransferInput(&input); err != nil {
		return nil, err
	}

	requestHash := calculateRequestHash(input)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transfer transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	requestRepo := repository.NewTransferRequestRepository(tx)

	request, existingResult, err := s.registerTransferRequest(
		ctx,
		requestRepo,
		input,
		requestHash,
	)
	if err != nil {
		return nil, err
	}

	if existingResult != nil {
		return existingResult, nil
	}

	accountRepo := repository.NewAccountRepository(tx)

	sourceAccount, err := accountRepo.FindPrimaryByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("find sender account: %w", err)
	}

	if sourceAccount.ID == input.RecipientAccountID {
		return s.rejectTransfer(ctx, tx, request.ID, ErrTransferToSelf)
	}

	sourceAccount, destinationAccount, err := lockTransferAccounts(
		ctx,
		accountRepo,
		sourceAccount.ID,
		input.RecipientAccountID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return s.rejectTransfer(ctx, tx, request.ID, ErrInvalidRecipient)
		}

		return nil, fmt.Errorf("lock transfer accounts: %w", err)
	}

	if err := validateTransferAccounts(
		sourceAccount,
		destinationAccount,
		input.UserID,
	); err != nil {
		return s.rejectTransfer(ctx, tx, request.ID, err)
	}

	balance, err := accountRepo.GetBalance(ctx, sourceAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("get sender balance: %w", err)
	}

	if err := validateTransferBalance(balance, input.AmountKopecks); err != nil {
		return s.rejectTransfer(ctx, tx, request.ID, err)
	}

	transactionID, err := createTransferRecords(
		ctx,
		tx,
		input.UserID,
		sourceAccount.ID,
		destinationAccount.ID,
		input.AmountKopecks,
	)
	if err != nil {
		return nil, fmt.Errorf("create transfer records: %w", err)
	}

	transactionIDPtr := &transactionID

	if err := requestRepo.Complete(
		ctx,
		request.ID,
		domain.TransferRequestStatusExecuted,
		transactionIDPtr,
		nil,
	); err != nil {
		return nil, fmt.Errorf("complete transfer request: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transfer: %w", err)
	}

	return &TransferResult{
		RequestID:     request.ID,
		Status:        domain.TransferRequestStatusExecuted,
		TransactionID: transactionIDPtr,
	}, nil
}
