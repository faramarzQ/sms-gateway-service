package dtos

import "github.com/faramarzQ/sms-gateway-service/internals/value_objects"

type UserTrafficClass struct {
	Class value_objects.TrafficClass `json:"class"`
}
