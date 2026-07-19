package dtos

import "github.com/faramarzQ/sms-gateway-service/internals/value_objects"

type SMSMessage struct {
	ID          uint64                `json:"id"`
	ClientID    string                `json:"client_id"`
	UserID      uint64                `json:"user_id"`
	PhoneNumber string                `json:"phone_number"`
	Message     string                `json:"message"`
	Type        value_objects.SMSType `json:"type"`
}
