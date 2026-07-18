package models

import (
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"time"
)

type SMSStatus string

const SMSStatusPending SMSStatus = "pending"
const SMSStatusRejected SMSStatus = "rejected"
const SMSStatusDelivered SMSStatus = "delivered"

type SMS struct {
	ID uint64 `gorm:"primaryKey"`

	ClientID string `gorm:"size:100;not null;index"`

	UserID uint64 `gorm:"index;not null"`

	PhoneNumber string `gorm:"size:20;not null"`

	Message string `gorm:"type:text;not null"`

	Type value_objects.SMSType `gorm:"type:varchar(20);not null"`

	Status SMSStatus `gorm:"type:varchar(20);not null;default:'pending'"`

	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
}
