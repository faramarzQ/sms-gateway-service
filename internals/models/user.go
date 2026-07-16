package models

import (
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"time"
)

type User struct {
	ID           uint64                     `gorm:"primaryKey"`
	Name         string                     `gorm:"size:255;not null"`
	TrafficClass value_objects.TrafficClass `gorm:"type:varchar(20);not null;default:'standard'"`
	Balance      int64                      `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
