package messages

import (
	"context"
	"fmt"
	"log"

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
	sql := "INSERT INTO messages (room_id, user_id, content) VALUES ($1, $2, $3) RETURNING content::text"

	s, err := r.db.Exec(ctx, sql, msg.RoomID, msg.UserID, msg.Content)
	if err != nil {
		return err
	}
	log.Println(s)
	return nil
}
func (r *Repository) GetMessagesByRoom(ctx context.Context, roomID string, limit int) ([]ChatPayload, error) {
	if limit <= 0 {
		limit = 50
	}
	sql := `
        SELECT room_id, user_id, content
        FROM messages 
        WHERE room_id = $1 
        ORDER BY created_at DESC 
        LIMIT $2;`

	rows, err := r.db.Query(ctx, sql, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("querying messages: %w", err)
	}
	defer rows.Close()

	var messages []ChatPayload
	for rows.Next() {
		var m ChatPayload
		if err := rows.Scan(&m.RoomID, &m.Sender, &m.Text); err != nil {
			return nil, fmt.Errorf("scanning message: %w", err)
		}
		messages = append(messages, m)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
