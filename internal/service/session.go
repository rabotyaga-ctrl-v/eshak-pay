package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"smartwallet/internal/domain"
	"smartwallet/internal/repository"
	"time"

	"github.com/google/uuid"
)

const sessionDuration = time.Hour

var ErrInvalidSession = errors.New("invalid session")

type SessionService struct {
	sessionRepo *repository.SessionRepository
}

func NewSessionService(sessionRepo *repository.SessionRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

func generateToken() (string, error) {
	tokenBytes := make([]byte, 32)

	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return token, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *SessionService) Create(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().Add(sessionDuration),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return token, nil
}

func (s *SessionService) ValidateToken(
	ctx context.Context,
	token string,
) (uuid.UUID, error) {
	if token == "" {
		return uuid.Nil, ErrInvalidSession
	}

	session, err := s.sessionRepo.FindActiveByTokenHash(
		ctx,
		hashToken(token),
	)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return uuid.Nil, ErrInvalidSession
		}

		return uuid.Nil, fmt.Errorf("validate session: %w", err)
	}

	return session.UserID, nil
}

func (s *SessionService) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	tokenHash := hashToken(token)

	session, err := s.sessionRepo.FindActiveByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil
		}

		return fmt.Errorf("find session for revocation: %w", err)
	}

	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}
