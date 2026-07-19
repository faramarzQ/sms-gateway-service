package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/cache"
	"github.com/faramarzQ/sms-gateway-service/internals/dtos"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	repoDto "github.com/faramarzQ/sms-gateway-service/internals/repositories/dtos"
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

const TrafficClassificationWindowHours = 4
const BulkTrafficThreshold = 500

type TrafficClassifierService struct {
	redis          *redis.Client
	userRepository *repositories.UserRepository
}

func NewTrafficClassifierService(redis *redis.Client, userRepo *repositories.UserRepository) *TrafficClassifierService {
	return &TrafficClassifierService{
		redis:          redis,
		userRepository: userRepo,
	}
}

func (s *TrafficClassifierService) Execute() error {
	ctx := context.Background()
	users, err := s.userRepository.GetAllUsersTrafficInfo(ctx)
	if err != nil {
		return err
	}

	changes := make([]dtos.TrafficClassChange, 0)

	const maxConcurrentUsers = 100

	semaphore := make(chan struct{}, maxConcurrentUsers)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, user := range users {
		wg.Add(1)

		go func(user repoDto.UserTrafficInfo) {
			defer wg.Done()

			semaphore <- struct{}{}        // acquire
			defer func() { <-semaphore }() // release

			history, err := s.GetTrafficHistory(ctx, user.ID)
			if err != nil {
				logger.Logger.Error("error calculating traffic history: %W", zap.Error(err))
				return
			}

			class := s.DetermineTrafficClass(history)
			if class == user.TrafficClass {
				return
			}

			mu.Lock()
			changes = append(changes, dtos.TrafficClassChange{
				UserID: user.ID,
				Class:  class,
			})
			mu.Unlock()

		}(user)
	}

	wg.Wait()

	err = s.userRepository.BulkUpdateTrafficClass(ctx, changes)
	if err != nil {
		return err
	}

	err = s.cacheUserTrafficClass(ctx, changes)
	if err != nil {
		return err
	}

	return nil
}

func (s *TrafficClassifierService) GetTrafficHistory(
	ctx context.Context,
	userID uint64,
) ([]int64, error) {

	now := time.Now().UTC()

	keys := make([]string, 0, TrafficClassificationWindowHours)
	for i := 0; i < TrafficClassificationWindowHours; i++ {
		t := now.Add(-time.Duration(i) * time.Hour)

		key := fmt.Sprintf(
			cache.UserRequestsPerHour,
			userID,
			t.Format("2006010215"),
		)

		keys = append(keys, key)
	}

	values, err := s.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	counts := make([]int64, TrafficClassificationWindowHours)

	for i, v := range values {
		if v == nil {
			continue
		}

		count, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return nil, err
		}

		counts[i] = count
	}

	return counts, nil
}

func (s *TrafficClassifierService) DetermineTrafficClass(history []int64) value_objects.TrafficClass {
	for _, count := range history {
		if count >= BulkTrafficThreshold {
			return value_objects.TrafficClassBulk
		}
	}

	return value_objects.TrafficClassStandard
}

func (s *TrafficClassifierService) SetUserTrafficClass(
	ctx context.Context,
	userID uint64,
	class value_objects.TrafficClass,
) error {

	key := fmt.Sprintf(
		cache.UserTrafficClass,
		userID,
	)

	data := dtos.UserTrafficClass{
		Class: class,
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.redis.Set(
		ctx,
		key,
		bytes,
		24*time.Hour,
	).Err()
}

func (s *TrafficClassifierService) cacheUserTrafficClass(
	ctx context.Context,
	users []dtos.TrafficClassChange,
) error {
	if len(users) == 0 {
		return nil
	}

	pipe := s.redis.Pipeline()

	for _, user := range users {
		key := fmt.Sprintf(cache.UserTrafficClass, user.UserID)

		payload, err := json.Marshal(dtos.UserTrafficClass{
			Class: user.Class,
		})
		if err != nil {
			return fmt.Errorf("marshal traffic class: %w", err)
		}

		pipe.Set(
			ctx,
			key,
			payload,
			24*time.Hour,
		)
	}

	_, err := pipe.Exec(ctx)
	return err
}
