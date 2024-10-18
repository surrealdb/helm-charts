package tests

import "testing"

func TestShouldUpdateSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		path     string
		subject  string
		expected bool
	}{
		{
			name:     "empty env",
			env:      "",
			path:     "deployment.yaml",
			subject:  "default",
			expected: false,
		},
		{
			name:     "wildcard env",
			env:      "*",
			path:     "deployment.yaml",
			subject:  "default",
			expected: true,
		},
		{
			name:     "path env",
			env:      "deployment.yaml/*",
			path:     "deployment.yaml",
			subject:  "default",
			expected: true,
		},
		{
			name:     "path/subject env",
			env:      "deployment.yaml/default",
			path:     "deployment.yaml",
			subject:  "default",
			expected: true,
		},
		{
			name:     "path/different subject env",
			env:      "deployment.yaml/other",
			path:     "deployment.yaml",
			subject:  "default",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := shouldUpdateSnapshot(tt.env, tt.path, tt.subject)
			if actual != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}
