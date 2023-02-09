package main

import (
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
)

func TestInTimeSpan(t *testing.T) {
	start, _ := time.Parse("15", "22")
	end, _ := time.Parse("15", "07")
	tt := []struct {
		check    string
		expected bool
	}{
		{
			check:    "22",
			expected: true,
		},
		{
			check:    "08",
			expected: false,
		},
		{
			check:    "05",
			expected: true,
		},
		{
			check:    "14",
			expected: false,
		},
	}
	for _, tc := range tt {
		check, _ := time.Parse("15", tc.check)
		result := inTimeSpan(start, end, check)
		assert.Equal(t, result, tc.expected)
	}

}
