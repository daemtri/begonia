package runtime

import (
	"testing"
)

func TestParseServiceType(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected ServiceType
	}{
		{
			name:     "cluster",
			input:    "app10A",
			expected: ServiceTypeCluster,
		},
		{
			name:     "service",
			input:    "app011",
			expected: ServiceTypeService,
		},
		{
			name:     "invalid",
			input:    "invalid",
			expected: ServiceTypeService,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ParseServiceType(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %v, but got %v (%s)", tc.expected, actual, tc.name)
			}
		})
	}
}
