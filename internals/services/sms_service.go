package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/cache"
	"github.com/faramarzQ/sms-gateway-service/internals/dtos"
	"github.com/faramarzQ/sms-gateway-service/internals/http/requests"
	"github.com/faramarzQ/sms-gateway-service/internals/http/responses"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

const max_concurrent_sms_batch = 20

type SMSService struct {
	repo             *repositories.SMSRepository
	messagePublisher *message_broker.Publisher
	userService      *UserService
	redis            *redis.Client
}

func NewSMSService(repo *repositories.SMSRepository, messagePublisher *message_broker.Publisher, userService *UserService, redis *redis.Client) *SMSService {
	return &SMSService{
		repo:             repo,
		messagePublisher: messagePublisher,
		userService:      userService,
		redis:            redis,
	}
}

func (s *SMSService) SendSMS(ctx context.Context, req requests.SendSMSRequest) (*responses.SendSMSErrorResponse, error) {
	smsMessage := dtos.SMSMessage{
		UserID:      req.UserID,
		PhoneNumber: req.PhoneNumber,
		Message:     req.Message,
		Type:        req.Type,
	}

	err := s.incrementUserRequestCount(ctx, req.UserID, 12)
	if err != nil {
		return nil, err
	}

	smsErrors := responses.SendSMSErrorResponse{}
	userTrafficClass, err := s.userService.GetUserTrafficClass(ctx, req.UserID)
	if err != nil {
		smsErrors.Errors = append(smsErrors.Errors, responses.SendSMSError{
			Message:  "failed sending sms",
			ClientId: req.ClientID,
		})
		return &smsErrors, err
	}

	err = s.sendSingleSMS(ctx, smsMessage, *userTrafficClass)
	if err != nil {
		smsErrors.Errors = append(smsErrors.Errors, responses.SendSMSError{
			Message:  "failed sending sms",
			ClientId: req.ClientID,
		})
		return &smsErrors, err
	}

	return nil, nil
}

func (s *SMSService) SendSMSBatch(ctx context.Context, req requests.SendSMSBatchRequest) (*responses.SendSMSErrorResponse, error) {
	//ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	//defer cancel()
	//TODO: add timeout

	err := s.incrementUserRequestCount(ctx, req.UserID, len(req.Messages))
	if err != nil {
		return nil, err
	}

	userTrafficClass, err := s.userService.GetUserTrafficClass(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	semaphore := make(chan struct{}, max_concurrent_sms_batch)
	var wg sync.WaitGroup

	smsErrors := responses.SendSMSErrorResponse{}
	var mu sync.Mutex

	for i, message := range req.Messages {
		wg.Add(1)

		go func(id int, message requests.SMSRequest) {
			defer wg.Done()

			// acquire semaphore
			semaphore <- struct{}{}
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

			err = s.sendSingleSMS(ctx, smsMessage, *userTrafficClass)
			if err != nil {
				logger.Logger.Error(err.Error())

				mu.Lock()
				smsErrors.Errors = append(smsErrors.Errors, responses.SendSMSError{
					Message:  "failed sending sms",
					ClientId: message.ClientID,
				})
				mu.Unlock()
			}

		}(i, message)
	}

	wg.Wait()

	return &smsErrors, nil
}

func (s *SMSService) sendSingleSMS(ctx context.Context, sms dtos.SMSMessage, userTrafficClass value_objects.TrafficClass) error {
	smsRow, err := s.repo.StoreSMS(ctx, sms)
	if err != nil {
		return err
	}

	sms.ID = smsRow.ID

	body, err := json.Marshal(sms)
	if err != nil {
		return fmt.Errorf("marshal sms: %w", err)
	}

	routingKey := s.GetRoutingKey(sms.Type, userTrafficClass)
	if routingKey == "" {
		if updateErr := s.repo.UpdateStatus(
			ctx,
			sms.ID,
			models.SMSStatusRejected,
		); updateErr != nil {
			return fmt.Errorf(
				"publish sms failed: %v, update status failed: %w",
				err,
				updateErr,
			)
		}
	}

	if err := s.messagePublisher.Publish(
		ctx,
		routingKey,
		body,
	); err != nil {

		if updateErr := s.repo.UpdateStatus(
			ctx,
			sms.ID,
			models.SMSStatusRejected,
		); updateErr != nil {
			return fmt.Errorf(
				"publish sms failed: %v, update status failed: %w",
				err,
				updateErr,
			)
		}
	}

	return nil
}

func (*SMSService) GetRoutingKey(smsType value_objects.SMSType, userTrafficClass value_objects.TrafficClass) string {
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

	return ""
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

	count, err := s.redis.IncrBy(ctx, key, int64(value)).Result()
	if err != nil {
		return fmt.Errorf("increment request count: %w", err)
	}

	fmt.Println(count)

	return nil
}
