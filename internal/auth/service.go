package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) Register(ctx context.Context, u *UserReq) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {

		return "", fmt.Errorf("Bcrypt failed: %v", err)
	}

	id, err := s.repo.CreateUser(ctx, &User{
		Username:     u.Username,
		ID:           uuid.New(),
		PasswordHash: hash,
	})
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		UserID:   id.String(),
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed to sign token")
	}

	return tokenString, nil
}

func (s *Service) Login(ctx context.Context, u *UserReq) (string, error) {
	user, err := s.repo.GetByUsername(ctx, u.Username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(u.Password))
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed to sign token")
	}

	return tokenString, nil
}

func (s *Service) LogOut(ctx context.Context, authInfo *AuthInfo) error {
	err := s.repo.ExpireToken(ctx, authInfo.Token, &authInfo.Claims.ExpiresAt.Time)
	if err != nil {
		return fmt.Errorf("failed to save expired token: %w", err)
	}

	return nil
}
