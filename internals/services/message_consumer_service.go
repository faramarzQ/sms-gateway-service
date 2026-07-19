package services

import (
	"context"
	"encoding/json"
	"github.com/faramarzQ/sms-gateway-service/internals/dtos"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"go.uber.org/zap"
	"time"
)

type MessageConsumerService struct {
	messageConsumer *message_broker.Consumer
	smsRepository   *repositories.SMSRepository
}

func NewMessageConsumerService(messageConsumer *message_broker.Consumer, smsRepository *repositories.SMSRepository) *MessageConsumerService {
	return &MessageConsumerService{
		messageConsumer: messageConsumer,
		smsRepository:   smsRepository,
	}
}

func (s *MessageConsumerService) Consume() error {
	err := s.consumeSMSAck()
	if err != nil {
		return err
	}

	return nil
}

func (s *MessageConsumerService) consumeSMSAck() error {
	ctx := context.Background()

	for {
		msgs, err := s.messageConsumer.Consume(message_broker.QueueSMSAck)
		if err != nil {
			logger.Logger.Error("failed to start consumer", zap.Error(err))

			time.Sleep(5 * time.Second)
			continue
		}

		logger.Logger.Info("SMS ACK consumer started")

		for msg := range msgs {

			var ack dtos.SMSAckMessage
			if err := json.Unmarshal(msg.Body, &ack); err != nil {
				logger.Logger.Error(
					"failed to parse SMS ACK",
					zap.Error(err),
				)

				_ = msg.Nack(false, false)
				continue
			}

			smsStatus := models.SMSStatus(string(ack.Status))

			err := s.smsRepository.UpdateStatus(ctx, ack.SMSID, smsStatus)
			if err != nil {
				logger.Logger.Error(
					"failed to update db status",
					zap.Error(err),
				)
			}

			if err := msg.Ack(false); err != nil {
				logger.Logger.Error(
					"failed to ack RabbitMQ message",
					zap.Error(err),
				)
			}
		}

		logger.Logger.Warn("consumer disconnected, reconnecting in 5 seconds...")

		time.Sleep(5 * time.Second)
	}
}
