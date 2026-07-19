package dtos

type SMSAckMessage struct {
	SMSID  uint64 `json:"sms_id"`
	Status string `json:"status"`
}
