package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/cache"
	"github.com/faramarzQ/sms-gateway-service/internals/dtos"
	httpErrors "github.com/faramarzQ/sms-gateway-service/internals/http/errors"
	"github.com/faramarzQ/sms-gateway-service/internals/http/requests"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type UserService struct {
	repo  repositories.UserRepositoryInterface
	redis *redis.Client
}

func NewUserService(userRepo repositories.UserRepositoryInterface, redis *redis.Client) *UserService {
	return &UserService{
		repo:  userRepo,
		redis: redis,
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

func (s *UserService) GetUserTrafficClass(ctx context.Context, userId uint64) (*value_objects.TrafficClass, error) {
	key := fmt.Sprintf(cache.UserTrafficClass, userId)

	cachedUser, err := s.redis.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if err != redis.Nil {
		var trafficClass dtos.UserTrafficClass
		if err := json.Unmarshal(cachedUser, &trafficClass); err != nil {
			return nil, fmt.Errorf("unmarshal traffic class: %w", err)
		}

		return &trafficClass.Class, nil
	}

	// cache miss:

	trafficClass, err := s.repo.GetUserTrafficClass(ctx, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpErrors.ErrUserNotFound
		}
		return nil, err
	}

	trafficClassDto := dtos.UserTrafficClass{
		Class: *trafficClass,
	}

	data, err := json.Marshal(trafficClassDto)
	if err != nil {
		return nil, err
	}

	err = s.redis.Set(
		ctx,
		key,
		data,
		24*time.Hour,
	).Err()
	if err != nil {
		logger.Logger.Error("redis set failed", zap.Error(err))
	}

	return trafficClass, nil
}
