package dtos

import "github.com/faramarzQ/sms-gateway-service/internals/value_objects"

type TrafficClassChange struct {
	UserID uint64
	Class  value_objects.TrafficClass
}
