package mocks

import (
	"context"

	"github.com/faramarzQ/sms-gateway-service/internals/dtos"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	"github.com/stretchr/testify/mock"
)

type SMSRepositoryMock struct {
	mock.Mock
}

func (m *SMSRepositoryMock) StoreSMS(
	ctx context.Context,
	sms dtos.SMSMessage,
) (*models.SMS, error) {

	args := m.Called(ctx, sms)

	var result *models.SMS
	if args.Get(0) != nil {
		result = args.Get(0).(*models.SMS)
	}

	return result, args.Error(1)
}

func (m *SMSRepositoryMock) UpdateStatus(
	ctx context.Context,
	id uint64,
	status models.SMSStatus,
) error {

	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *SMSRepositoryMock) GetUserSMS(
	ctx context.Context,
	userID uint64,
) ([]models.SMS, error) {

	args := m.Called(ctx, userID)

	var result []models.SMS
	if args.Get(0) != nil {
		result = args.Get(0).([]models.SMS)
	}

	return result, args.Error(1)
}
