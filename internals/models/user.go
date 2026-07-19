package models

import (
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"time"
)

type User struct {
	ID           uint64                     `json:"id" gorm:"primaryKey"`
	Name         string                     `json:"name" gorm:"size:255;not null"`
	TrafficClass value_objects.TrafficClass `json:"traffic_class" gorm:"type:varchar(20);not null;default:'standard'"`
	Balance      int64                      `json:"balance" gorm:"not null"`
	CreatedAt    time.Time                  `json:"created_at"`
	UpdatedAt    time.Time                  `json:"updated_at"`
}
