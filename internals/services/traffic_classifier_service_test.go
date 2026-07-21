package services

import (
	"testing"

	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"
)

func TestDetermineTrafficClass(t *testing.T) {
	service := &TrafficClassifierService{}

	tests := []struct {
		name     string
		history  []int64
		expected value_objects.TrafficClass
	}{
		{
			name:     "empty history",
			history:  []int64{},
			expected: value_objects.TrafficClassStandard,
		},
		{
			name:     "all below threshold",
			history:  []int64{10, 50, 200, 499},
			expected: value_objects.TrafficClassStandard,
		},
		{
			name:     "equal to threshold",
			history:  []int64{10, 500, 20, 30},
			expected: value_objects.TrafficClassBulk,
		},
		{
			name:     "above threshold",
			history:  []int64{100, 650, 10, 20},
			expected: value_objects.TrafficClassBulk,
		},
		{
			name:     "multiple hours above threshold",
			history:  []int64{700, 600, 800, 900},
			expected: value_objects.TrafficClassBulk,
		},
		{
			name:     "threshold on last hour",
			history:  []int64{1, 2, 3, 500},
			expected: value_objects.TrafficClassBulk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.DetermineTrafficClass(tt.history)

			if result != tt.expected {
				t.Fatalf(
					"expected %v, got %v",
					tt.expected,
					result,
				)
			}
		})
	}
}
