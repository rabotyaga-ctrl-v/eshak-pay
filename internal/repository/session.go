package repository

import (
	"context"
	"smartwallet/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"errors"
	"fmt"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(
	ctx context.Context,
	session *domain.Session,
) error {
	query := `
		INSERT INTO sessions (
    		id,
    		user_id,
    		token_hash,
    		expires_at
		)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query, session.ID, session.UserID, session.TokenHash, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *SessionRepository) FindActiveByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {

	query := `
        select id, user_id, token_hash, expires_at,
		revoked_at, created_at from sessions
		where token_hash = $1 
		and revoked_at IS NULL
		and expires_at > now()
    `

	var session domain.Session

	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.CreatedAt,
	)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf("find active session: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) Revoke(
	ctx context.Context,
	sessionID uuid.UUID,
) error {
	query := `
		UPDATE sessions
		SET revoked_at = now()
		WHERE id = $1
		AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}
