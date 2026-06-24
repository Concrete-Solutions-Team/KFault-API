package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAlreadyExists = errors.New("User with this username already exists")
var ErrInternal = errors.New("Internal Server Error")

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash []byte
}

type UserReq struct {
	Username string
	Password string
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) (uuid.UUID, error) {
	sql :=
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id::text"
	var id uuid.UUID
	err := r.db.QueryRow(ctx, sql, user.Username, user.PasswordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, ErrAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("Internal server error: %w", err)
	}

	return id, nil
}
func (r *Repository) ExpireToken(ctx context.Context, token string, expiresAt *time.Time) error {
	sql :=
		"INSERT INTO expired_tokens (token, expires_at) VALUES ($1, $2)"

	_, err := r.db.Exec(ctx, sql, token, *expiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return fmt.Errorf("Internal server error: %w", err)
	}

	return nil
}

func (r *Repository) IsTokenRevoked(ctx context.Context, tokenString string) bool {
	sql :=
		"SELECT token FROM expired_tokens WHERE token = $1"

	var token string
	err := r.db.QueryRow(ctx, sql, tokenString).Scan(&token)

	if err != nil {
		return false
	}
	if token != "" {
		return true
	}

	return false
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var userID uuid.UUID
	var passHash string
	err := r.db.QueryRow(ctx,
		"SELECT id, password_hash FROM users WHERE username = $1 LIMIT 1;",
		username,
	).Scan(&userID, &passHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("invalid credentials %w", err)
		}
		return nil, fmt.Errorf("internal server error: %w", err)
	}

	return &User{
		ID:           userID,
		Username:     username,
		PasswordHash: []byte(passHash),
	}, nil
}
