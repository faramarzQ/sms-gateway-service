package services

import (
	"context"
	"errors"
	httpErrors "github.com/faramarzQ/sms-gateway-service/internals/http/errors"
	"github.com/faramarzQ/sms-gateway-service/internals/http/requests"
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

func (s *UserService) IncreaseBalance(ctx context.Context, userId uint64, req requests.IncreaseBalanceRequest) error {
	userExists, err := s.repo.ExistsById(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httpErrors.ErrUserNotFound
		}
	}
	if !userExists {
		return httpErrors.ErrUserNotFound
	}

	if req.Amount < 0 {
		return errors.New("amount must be greater than zero")
	}

	err = s.repo.IncreaseBalance(ctx, userId, req.Amount)
	if err != nil {
		return err
	}

	return nil
}
