package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/dtos"
	"github.com/faramarzQ/sms-gateway-service/internals/http/requests"
	"github.com/faramarzQ/sms-gateway-service/internals/http/responses"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"sync"
	"time"
)

const max_concurrent_sms_batch = 20

type SMSService struct {
	repo             *repositories.SMSRepository
	messagePublisher *message_broker.Publisher
	userService      *UserService
}

func NewSMSService(repo *repositories.SMSRepository, messagePublisher *message_broker.Publisher, userService *UserService) *SMSService {
	return &SMSService{
		repo:             repo,
		messagePublisher: messagePublisher,
		userService:      userService,
	}
}

func (service *SMSService) SendSMS(ctx context.Context, req requests.SendSMSRequest) error {
	smsMessage := dtos.SMSMessage{
		UserID:      req.UserID,
		PhoneNumber: req.PhoneNumber,
		Message:     req.Message,
		Type:        req.Type,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	userTrafficClass, err := service.userService.GetUserTrafficClass(ctx, req.UserID)
	if err != nil {
		return err
	}

	err = service.sendSingleSMS(ctx, smsMessage, *userTrafficClass)
	if err != nil {
		return err
	}

	return nil
}

func (service *SMSService) SendSMSBatch(ctx context.Context, req requests.SendSMSBatchRequest) []*error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	userTrafficClass, err := service.userService.GetUserTrafficClass(ctx, req.UserID)
	if err != nil {
		return []*error{&err}
	}

	semaphore := make(chan struct{}, max_concurrent_sms_batch)
	var wg sync.WaitGroup

	//response := responses.BatchSMSResult{}
	errors := make([]*error, 0)

	for i, message := range req.Messages {
		wg.Add(1)

		go func(id int) {
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

			err = service.sendSingleSMS(ctx, smsMessage, *userTrafficClass)
			if err != nil {
				errors = append(errors, &err)
			}

		}(i)
	}

	wg.Wait()

	return errors
}

func (service *SMSService) sendSingleSMS(ctx context.Context, sms dtos.SMSMessage, userTrafficClass value_objects.TrafficClass) error {
	smsRow, err := service.repo.StoreSMS(ctx, sms)
	if err != nil {
		return err
	}

	sms.ID = smsRow.ID

	body, err := json.Marshal(sms)
	if err != nil {
		return fmt.Errorf("marshal sms: %w", err)
	}

	routingKey := service.GetRoutingKey(sms.Type, userTrafficClass)
	if routingKey == "" {
		if updateErr := service.repo.UpdateStatus(
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

	if err := service.messagePublisher.Publish(
		ctx,
		routingKey,
		body,
	); err != nil {

		if updateErr := service.repo.UpdateStatus(
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

func (service *SMSService) GetRoutingKey(smsType value_objects.SMSType, userTrafficClass value_objects.TrafficClass) string {
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

func (service *SMSService) GetReport(ctx context.Context, userId uint64) (*responses.UserSMSReportResponse, error) {
	smsList, err := service.repo.GetUserSMS(ctx, userId)
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
