package repositories

import (
	"context"
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetUserById(ctx context.Context, id uint64) (*models.User, error) {
	user := models.User{}
	err := r.db.WithContext(ctx).
		First(&user, id).Error

	fmt.Println(id, user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) ExistsById(ctx context.Context, id uint64) (bool, error) {
	var exists bool

	err := r.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).
		Scan(&exists).
		Error

	return exists, err
}

func (r *UserRepository) IncreaseBalance(ctx context.Context, id uint64, amount int64) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("balance", gorm.Expr("balance + ?", amount)).
		Error
}

func (r *UserRepository) GetUserTrafficClass(ctx context.Context, userId uint64) (*value_objects.TrafficClass, error) {
	var trafficClass value_objects.TrafficClass

	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("traffic_class").
		Where("id = ?", userId).
		Scan(&trafficClass).Error

	if err != nil {
		return nil, fmt.Errorf("get user traffic class: %w", err)
	}

	return &trafficClass, nil
}
