package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"smartwallet/internal/domain"
	"smartwallet/internal/repository"
	"smartwallet/internal/security"

	"github.com/jackc/pgx/v5/pgxpool"

	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput       = errors.New("email and password are required")
	ErrEmailTaken         = errors.New("email is already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrIdNotExists        = errors.New("Id not exists")
)

type UserService struct {
	userRepo       *repository.UserRepository
	db             *pgxpool.Pool
	accountService *AccountService
}

func NewUserService(
	userRepo *repository.UserRepository,
	db *pgxpool.Pool,
	accountService *AccountService,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		db:             db,
		accountService: accountService,
	}
}

func (s *UserService) Register(
	ctx context.Context,
	email string,
	password string,
) (uuid.UUID, error) {
	email = strings.TrimSpace(email)

	if email == "" || password == "" {
		return uuid.Nil, ErrInvalidInput
	}

	if len(email) > 254 {
		return uuid.Nil, ErrInvalidInput
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return uuid.Nil, ErrInvalidInput
	}

	passwordLength := utf8.RuneCountInString(password)

	if passwordLength < 12 || passwordLength > 128 {
		return uuid.Nil, ErrInvalidInput
	}

	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin registration transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	userRepo := repository.NewUserRepository(tx)

	if err := userRepo.Create(ctx, &user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return uuid.Nil, ErrEmailTaken
		}

		return uuid.Nil, fmt.Errorf("register user: %w", err)
	}

	if err := s.accountService.CreatePrimaryAccount(ctx, tx, user.ID); err != nil {
		return uuid.Nil, fmt.Errorf("create primary account: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit registration transaction: %w", err)
	}

	return user.ID, nil
}

func (s *UserService) Authenticate(
	ctx context.Context,
	email, password string,
) (uuid.UUID, error) {
	email = strings.TrimSpace(email)

	if email == "" || password == "" {
		return uuid.Nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return uuid.Nil, ErrInvalidCredentials
		}

		return uuid.Nil, fmt.Errorf("find user for authentication: %w", err)
	}

	valid, err := security.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return uuid.Nil, fmt.Errorf("verify password: %w", err)
	}

	if !valid {
		return uuid.Nil, ErrInvalidCredentials
	}

	return user.ID, nil
}

func (s *UserService) GetEmailById(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {

	user, err := s.userRepo.FindByUserId(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrIdNotExists
		}

		return "", fmt.Errorf("find email by userid: %w", err)
	}

	return user.Email, nil
}
