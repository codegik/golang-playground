package main

import (
	"reflect"
	"testing"
)

func TestGeneratePermutation(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "Example 1",
			input:    []int{0, 1, 1, 0, 3},
			expected: []int{4, 1, 3, 5, 2},
		},
		{
			name:     "Example 2",
			input:    []int{0, 1, 2},
			expected: []int{3, 2, 1},
		},
		{
			name:     "Example 3",
			input:    []int{0, 0, 0, 0},
			expected: []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generatePermutation(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("generatePermutation(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
