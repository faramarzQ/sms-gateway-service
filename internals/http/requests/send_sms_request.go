package requests

import "github.com/faramarzQ/sms-gateway-service/internals/value_objects"

type SMSRequest struct {
	PhoneNumber string                `json:"phone_number" binding:"required"`
	Message     string                `json:"message" binding:"required,max=160"`
	Type        value_objects.SMSType `json:"type" binding:"required,oneof=ordinary express"`
}

type SendSMSRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	SMSRequest
}

type SendSMSBatchRequest struct {
	UserID   uint64       `json:"user_id" binding:"required"`
	Messages []SMSRequest `json:"messages" binding:"required,min=1,dive"`
}
