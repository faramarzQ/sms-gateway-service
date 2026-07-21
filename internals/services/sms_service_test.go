package services

import (
	"testing"

	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
)

func TestCalculateRoutingKey(t *testing.T) {
	service := &SMSService{}

	tests := []struct {
		name         string
		smsType      value_objects.SMSType
		trafficClass value_objects.TrafficClass
		expected     string
	}{
		{
			name:         "standard ordinary",
			smsType:      value_objects.SMSTypeOrdinary,
			trafficClass: value_objects.TrafficClassStandard,
			expected:     "standard",
		},
		{
			name:         "standard express",
			smsType:      value_objects.SMSTypeExpress,
			trafficClass: value_objects.TrafficClassStandard,
			expected:     "standard.express",
		},
		{
			name:         "bulk ordinary",
			smsType:      value_objects.SMSTypeOrdinary,
			trafficClass: value_objects.TrafficClassBulk,
			expected:     "bulk",
		},
		{
			name:         "bulk express",
			smsType:      value_objects.SMSTypeExpress,
			trafficClass: value_objects.TrafficClassBulk,
			expected:     "bulk.express",
		},
		{
			name:         "unknown traffic class defaults to standard",
			smsType:      value_objects.SMSTypeOrdinary,
			trafficClass: value_objects.TrafficClass("invalid"),
			expected:     "standard",
		},
		{
			name:         "unknown sms type defaults to standard",
			smsType:      value_objects.SMSType("invalid"),
			trafficClass: value_objects.TrafficClassStandard,
			expected:     "standard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateRoutingKey(tt.smsType, tt.trafficClass)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
