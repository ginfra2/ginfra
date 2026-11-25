package kube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_kubeServiceConfigMapName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "config",
		},
		{
			name:     "simple name",
			input:    "service",
			expected: "configservice",
		},
		{
			name:     "name with special characters",
			input:    "service-name123",
			expected: "configservice-name123",
		},
		{
			name:     "name with spaces",
			input:    "my service",
			expected: "configmy service",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := kubeServiceConfigMapName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func Test_kubeServiceConfigBlobName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "configblob",
		},
		{
			name:     "simple name",
			input:    "service",
			expected: "configblobservice",
		},
		{
			name:     "name with special characters",
			input:    "service-name123",
			expected: "configblobservice-name123",
		},
		{
			name:     "name with spaces",
			input:    "my service",
			expected: "configblobmy service",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := kubeServiceConfigMapBlobName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
