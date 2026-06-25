package rooms

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateRoom(ctx context.Context, rd *RoomData, ownerID *string) (*RoomData, error) {
	if rd.Name == "" {
		return nil, ErrInvalidData
	}
	_, err := s.repo.GetRoomByName(ctx, rd.Name)
	fmt.Println(rd.Name)
	if err != ErrRoomNotFound {
		if err != nil {
			return nil, err
		}
	}
	res, err := s.repo.CreateRoom(ctx, rd, ownerID)
	fmt.Println(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) GetAllRooms(ctx context.Context) ([]*RoomData, error) {
	return s.repo.GetAllRooms(ctx)
}

func (s *Service) DeleteRoom(ctx context.Context, roomID string) error {
	return s.repo.DeleteRoom(ctx, roomID)
}
