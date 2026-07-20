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
	"github.com/faramarzQ/sms-gateway-service/internals/http/responses"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"sync"
	"time"
)

const max_concurrent_sms_batch = 20

type SMSService struct {
	repo             *repositories.SMSRepository
	messagePublisher *message_broker.Publisher
	userService      *UserService
	userRepository   *repositories.UserRepository
	redis            *redis.Client
}

func NewSMSService(repo *repositories.SMSRepository,
	messagePublisher *message_broker.Publisher,
	userService *UserService,
	userRepository *repositories.UserRepository,
	redis *redis.Client) *SMSService {

	return &SMSService{
		repo:             repo,
		messagePublisher: messagePublisher,
		userService:      userService,
		userRepository:   userRepository,
		redis:            redis,
	}
}

func (s *SMSService) SendSMS(ctx context.Context, req requests.SendSMSRequest) (*responses.SendSMSErrorResponse, error) {
	smsMessage := dtos.SMSMessage{
		ClientID:    req.ClientID,
		UserID:      req.UserID,
		PhoneNumber: req.PhoneNumber,
		Message:     req.Message,
		Type:        req.Type,
	}

	hasBalance, err := s.userRepository.HasBalance(ctx, req.UserID, 1)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpErrors.ErrUserNotFound
		}

		return nil, err
	}

	if !hasBalance {
		return nil, httpErrors.ErrUserBalanceExceeded
	}

	// cache user request count in bucket
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := s.incrementUserRequestCount(ctx, req.UserID, 1); err != nil {
			logger.Logger.Error("increment request count failed",
				zap.Error(err),
			)
		}
	}()

	smsErrors := responses.SendSMSErrorResponse{}
	userTrafficClass, err := s.userService.GetUserTrafficClass(ctx, req.UserID)
	if err != nil {
		smsErrors.Errors = append(smsErrors.Errors, responses.SendSMSError{
			Message:  "failed sending sms",
			ClientId: req.ClientID,
		})
		return &smsErrors, err
	}

	err = s.storeAndPublishSingleSMS(ctx, smsMessage, *userTrafficClass)
	if err != nil {
		if httpErrors.IsDomainError(err) {
			smsErrors.Errors = append(smsErrors.Errors, responses.SendSMSError{
				Message:  err.Error(),
				ClientId: req.ClientID,
			})
			return &smsErrors, err
		}
		smsErrors.Errors = append(smsErrors.Errors, responses.SendSMSError{
			Message:  "failed sending sms",
			ClientId: req.ClientID,
		})
		return &smsErrors, err
	}

	return &smsErrors, nil
}

func (s *SMSService) storeAndPublishSingleSMS(ctx context.Context, sms dtos.SMSMessage, userTrafficClass value_objects.TrafficClass) error {

	//NOTE: if consistency matters more, do a check on redundancy of sms.ClientID
	//NOTE: if consistency matters more, commit StoreSMS and ConsumeBalance in a transaction
	//NOTE: if consistency matters more, use outbox pattern

	err := s.userRepository.ConsumeBalance(ctx, sms.UserID, 1)
	if err != nil {
		if rollbackErr := s.repo.UpdateStatus(
			ctx,
			sms.ID,
			models.SMSStatusFailed,
		); rollbackErr != nil {
			return fmt.Errorf(
				"publish sms failed: %v, consume balance failed: %w",
				err,
				rollbackErr,
			)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httpErrors.ErrUserNotFound
		}
		return err
	}

	smsRow, err := s.repo.StoreSMS(ctx, sms)
	if err != nil {
		return err
	}

	sms.ID = smsRow.ID

	body, err := json.Marshal(sms)
	if err != nil {
		return fmt.Errorf("marshal sms: %w", err)
	}

	routingKey := s.CalculateRoutingKey(sms.Type, userTrafficClass)

	if err := s.messagePublisher.Publish(
		ctx,
		routingKey,
		body,
	); err != nil {
		// rollback
		if rollbackErr := s.repo.UpdateStatus(
			ctx,
			sms.ID,
			models.SMSStatusFailed,
		); rollbackErr != nil {
			return fmt.Errorf(
				"publish sms failed: %v, update status failed: %w",
				err,
				rollbackErr,
			)
		}

		if rollbackErr := s.userRepository.IncreaseBalance(ctx, sms.UserID, 1); rollbackErr != nil {
			return fmt.Errorf(
				"publish sms failed: %v, rollback balance failed: %w",
				err,
				rollbackErr,
			)
		}
	}

	return nil
}

func (s *SMSService) SendSMSBatch(ctx context.Context, req requests.SendSMSBatchRequest) (*responses.SendSMSErrorResponse, error) {
	// NOTE: if availability matters more, store all messages as batch in db

	hasBalance, err := s.userRepository.HasBalance(ctx, req.UserID, len(req.Messages))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httpErrors.ErrUserNotFound
		}

		return nil, err
	}

	if !hasBalance {
		return nil, httpErrors.ErrUserBalanceExceeded
	}

	// cache user request count in bucket
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := s.incrementUserRequestCount(cacheCtx, req.UserID, len(req.Messages)); err != nil {
			logger.Logger.Error("increment request count failed",
				zap.Error(err),
			)
		}
	}()

	userTrafficClass, err := s.userService.GetUserTrafficClass(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	semaphore := make(chan struct{}, max_concurrent_sms_batch)
	var wg sync.WaitGroup

	smsErrors := responses.SendSMSErrorResponse{}
	var mu sync.Mutex

	for i, message := range req.Messages {

		semaphore <- struct{}{}

		wg.Add(1)

		go func(id int, message requests.SMSRequest) {
			defer wg.Done()
			defer func() {
				// release semaphore
				<-semaphore
			}()

			smsMessage := dtos.SMSMessage{
				UserID:      req.UserID,
				PhoneNumber: message.PhoneNumber,
				Message:     message.Message,
				Type:        message.Type,
			}

			err := s.storeAndPublishSingleSMS(ctx, smsMessage, *userTrafficClass)
			if err != nil {
				if httpErrors.IsDomainError(err) {
					mu.Lock()
					smsErrors.Errors = append(smsErrors.Errors, responses.SendSMSError{
						Message:  err.Error(),
						ClientId: message.ClientID,
					})
					mu.Unlock()

					return
				}

				logger.Logger.Error(err.Error())

				mu.Lock()
				smsErrors.Errors = append(smsErrors.Errors, responses.SendSMSError{
					Message:  "failed sending sms",
					ClientId: message.ClientID,
				})
				mu.Unlock()

				return
			}

		}(i, message)
	}

	wg.Wait()

	return &smsErrors, nil
}

func (*SMSService) CalculateRoutingKey(smsType value_objects.SMSType, userTrafficClass value_objects.TrafficClass) string {
	switch userTrafficClass {
	case value_objects.TrafficClassStandard:
		switch smsType {
		case value_objects.SMSTypeOrdinary:
			return "standard"

		case value_objects.SMSTypeExpress:
			return "standard.express"
		}

	case value_objects.TrafficClassBulk:
		switch smsType {
		case value_objects.SMSTypeOrdinary:
			return "bulk"

		case value_objects.SMSTypeExpress:
			return "bulk.express"
		}
	}

	logger.Logger.Error(fmt.Sprintf("failed calculate routing key: sms type: %s , traffic class: %s", smsType, userTrafficClass))
	return "standard"
}

func (s *SMSService) GetReport(ctx context.Context, userId uint64) (*responses.UserSMSReportResponse, error) {
	smsList, err := s.repo.GetUserSMS(ctx, userId)
	if err != nil {
		return nil, err
	}

	response := responses.UserSMSReportResponse{
		UserID: userId,
		Total:  int64(len(smsList)),
		SMS:    make([]responses.SMSReportResponse, 0, len(smsList)),
	}

	for _, sms := range smsList {
		response.SMS = append(response.SMS, responses.SMSReportResponse{
			ID:          sms.ID,
			PhoneNumber: sms.PhoneNumber,
			Message:     sms.Message,
			Type:        sms.Type,
			Status:      string(sms.Status),
			CreatedAt:   sms.CreatedAt,
		})
	}

	return &response, nil

}

func (s *SMSService) incrementUserRequestCount(
	ctx context.Context,
	userID uint64,
	value int,
) error {

	now := time.Now().UTC()
	key := fmt.Sprintf(
		cache.UserRequestsPerHour,
		userID,
		now.Format("2006010215"),
	)

	_, err := s.redis.IncrBy(ctx, key, int64(value)).Result()
	if err != nil {
		return fmt.Errorf("increment request count: %w", err)
	}

	return nil
}
