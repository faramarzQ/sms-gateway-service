package repositories

import (
	"context"
	"fmt"
	dtos "github.com/faramarzQ/sms-gateway-service/internals/dtos"
	"github.com/faramarzQ/sms-gateway-service/internals/models"
	repoDtos "github.com/faramarzQ/sms-gateway-service/internals/repositories/dtos"
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"gorm.io/gorm"
	"strings"
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

func (r *UserRepository) GetAllUsersTrafficInfo(
	ctx context.Context,
) ([]repoDtos.UserTrafficInfo, error) {
	var users []repoDtos.UserTrafficInfo

	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("id", "traffic_class").
		Scan(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) UpdateTrafficClass(
	ctx context.Context,
	userID uint64,
	class value_objects.TrafficClass,
) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("traffic_class", class).Error
}

func (r *UserRepository) BulkUpdateTrafficClass(
	ctx context.Context,
	changes []dtos.TrafficClassChange,
) error {
	if len(changes) == 0 {
		return nil
	}

	var (
		caseSQL strings.Builder
		ids     []uint64
		args    []interface{}
	)

	caseSQL.WriteString("UPDATE users SET traffic_class = CASE id ")

	for _, change := range changes {
		caseSQL.WriteString("WHEN ? THEN ? ")
		args = append(args, change.UserID, change.Class)
		ids = append(ids, change.UserID)
	}

	caseSQL.WriteString("END WHERE id IN ?")

	args = append(args, ids)

	return r.db.WithContext(ctx).
		Exec(caseSQL.String(), args...).Error
}

func (r *UserRepository) GetBalance(
	ctx context.Context,
	userID uint64,
) (int64, error) {
	var result struct {
		Balance int64
	}

	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("balance").
		Where("id = ?", userID).
		Take(&result).
		Error
	if err != nil {
		return 0, err
	}

	return result.Balance, nil
}

func (r *UserRepository) HasBalance(
	ctx context.Context,
	userID uint64,
	amount int,
) (bool, error) {
	var result struct {
		Balance int
	}

	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("balance").
		Where("id = ?", userID).
		Take(&result).
		Error
	if err != nil {
		return false, err
	}

	return result.Balance >= amount, nil
}

func (r *UserRepository) ConsumeBalance(
	ctx context.Context,
	userID uint64,
	amount uint64,
) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND balance >= ?", userID, amount).
		Update("balance", gorm.Expr("balance - ?", amount)).Error

}
