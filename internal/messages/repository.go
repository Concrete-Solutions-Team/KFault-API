package messages

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

type MessageData struct {
	RoomID  string
	UserID  string
	Content string
}

func (r *Repository) InsertMessage(ctx context.Context, msg MessageData) error {
	sql := "INSERT INTO messages (room_id, user_id, content) VALUES ($1, $2, $3)"

	_, err := r.db.Query(ctx, sql, msg.RoomID, msg.UserID, msg.Content)
	if err != nil {
		return err
	}
	return nil
}
