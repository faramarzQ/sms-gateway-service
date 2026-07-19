package repositories

import (
	"context"
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/dtos"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	"gorm.io/gorm"
)

type SMSRepository struct {
	db *gorm.DB
}

func NewSMSRepository(db *gorm.DB) *SMSRepository {
	return &SMSRepository{
		db: db,
	}
}

func (repo *SMSRepository) StoreSMS(ctx context.Context, smsDto dtos.SMSMessage) (*models.SMS, error) {
	sms := models.SMS{
		UserID:      smsDto.UserID,
		PhoneNumber: smsDto.PhoneNumber,
		Message:     smsDto.Message,
		Type:        smsDto.Type,
		Status:      models.SMSStatusPending,
	}

	err := repo.db.Create(&sms).Error
	if err != nil {
		return nil, err
	}

	return &sms, nil
}

func (r *SMSRepository) UpdateStatus(
	ctx context.Context,
	id uint64,
	status models.SMSStatus,
) error {

	return r.db.
		Model(&models.SMS{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}

func (r *SMSRepository) GetUserSMS(
	ctx context.Context,
	userID uint64,
) ([]models.SMS, error) {

	var sms []models.SMS

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sms).Error

	if err != nil {
		return nil, fmt.Errorf("get user sms report: %w", err)
	}

	return sms, nil
}
