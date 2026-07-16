package services

import (
	"context"
	"errors"
	httpErrors "github.com/faramarzQ/sms-gateway-service/internals/http/errors"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"gorm.io/gorm"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		repo: userRepo,
	}
}

func (s *UserService) GetUser(ctx context.Context, userId uint64) (*models.User, error) {
	user, err := s.repo.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpErrors.ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}
