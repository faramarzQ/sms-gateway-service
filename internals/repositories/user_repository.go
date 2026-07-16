package repositories

import (
	"context"
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
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
