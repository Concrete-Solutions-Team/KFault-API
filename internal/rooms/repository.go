package rooms

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInternal     = errors.New("Internal Server Error")
	ErrRoomNotFound = errors.New("Room not found")
	ErrDbFailure    = errors.New("Database query failed")
	ErrInvalidData  = errors.New("Invalid data passed")
	ErrDuplicate    = errors.New("Name already in use")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

type RoomData struct {
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"`
	RoomID  string `json:"room_id"`
}

func (r *Repository) CreateRoom(ctx context.Context, rd *RoomData, ownerID *string) (*RoomData, error) {
	sql := "INSERT INTO rooms (name, owner_id) VALUES ($1, $2) RETURNING id::text, name::text, owner_id::text;"
	var room RoomData
	err := r.db.QueryRow(ctx, sql, rd.Name, ownerID).Scan(&room.RoomID, &room.Name, &room.OwnerID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &room, nil
}

func (r *Repository) DeleteRoom(ctx context.Context, roomID string) error {
	sql := "DELETE FROM rooms WHERE id = $1;"

	result, err := r.db.Exec(ctx, sql, roomID)
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no room found with that ID")
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetRoom(ctx context.Context, roomID string) (*RoomData, error) {
	sql := "SELECT id, name, owner_id FROM rooms WHERE id = $1;"
	var room RoomData
	err := r.db.QueryRow(ctx, sql, roomID).Scan(&room.RoomID, &room.Name, &room.OwnerID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("room not found")
		}
		log.Println("databse query failed, err: ", err)
		return nil, fmt.Errorf("database query failed")
	}

	return &room, nil
}

func (r *Repository) GetRoomByName(ctx context.Context, name string) (*RoomData, error) {
	sql := "SELECT id, name, owner_id FROM rooms WHERE name = $1;"
	var room RoomData
	err := r.db.QueryRow(ctx, sql, name).Scan(&room.RoomID, &room.Name, &room.OwnerID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		log.Println("databse query failed, err: ", err)
		return nil, ErrDbFailure
	}

	return &room, nil
}

func (r *Repository) GetAllRooms(ctx context.Context) ([]*RoomData, error) {
	sql := "SELECT id, name, owner_id FROM rooms;"
	rows, err := r.db.Query(ctx, sql)
	if err != nil {
		log.Println("database query failed, err: ", err)
		return nil, ErrDbFailure
	}
	defer rows.Close()

	var rooms []*RoomData
	for rows.Next() {
		var room RoomData
		if err := rows.Scan(&room.RoomID, &room.Name, &room.OwnerID); err != nil {
			log.Println("database scan failed, err: ", err)
			return nil, ErrDbFailure
		}
		rooms = append(rooms, &room)
	}

	if err := rows.Err(); err != nil {
		log.Println("database iteration error, err: ", err)
		return nil, ErrDbFailure
	}

	return rooms, nil
}
