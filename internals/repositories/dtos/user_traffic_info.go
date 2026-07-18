package dtos

import "github.com/faramarzQ/sms-gateway-service/internals/value_objects"

type UserTrafficInfo struct {
	ID           uint64
	TrafficClass value_objects.TrafficClass
}
