package responses

import (
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
	"time"
)

type UserSMSReportResponse struct {
	UserID uint64              `json:"user_id"`
	Total  int64               `json:"total"`
	SMS    []SMSReportResponse `json:"sms"`
}

type SMSReportResponse struct {
	ID          uint64                `json:"id"`
	PhoneNumber string                `json:"phone_number"`
	Message     string                `json:"message"`
	Type        value_objects.SMSType `json:"type"`
	Status      string                `json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
}
