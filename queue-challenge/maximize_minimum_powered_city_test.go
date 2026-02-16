package main

import "testing"

func TestMaxPower(t *testing.T) {
	tests := []struct {
		name     string
		stations []int
		r        int
		k        int
		expected int
	}{
		{
			name:     "example 1",
			stations: []int{1, 2, 4, 5, 0},
			r:        1,
			k:        2,
			expected: 5,
		},
		{
			name:     "example 2",
			stations: []int{4, 4, 4, 4},
			r:        0,
			k:        3,
			expected: 4,
		},
		{
			name:     "example 3",
			stations: []int{1, 2, 3},
			r:        1,
			k:        5,
			expected: 8,
		},
		{
			name:     "example 4",
			stations: []int{0, 0, 0},
			r:        1,
			k:        3,
			expected: 3,
		},
		{
			name:     "example 5",
			stations: []int{10},
			r:        0,
			k:        0,
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maxPower(tt.stations, tt.r, tt.k)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}
