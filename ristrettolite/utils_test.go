package ristrettolite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		zeroFunc func() any
		expected any
	}{
		{
			name: "Zero value of int",
			zeroFunc: func() any {
				return zeroValue[int]()
			},
			expected: 0,
		},
		{
			name: "Zero value of string",
			zeroFunc: func() any {
				return zeroValue[string]()
			},
			expected: "",
		},
		{
			name: "Zero value of struct",
			zeroFunc: func() any {
				type MyStruct struct {
					Field1 int
					Field2 string
				}
				return zeroValue[MyStruct]()
			},
			expected: struct {
				Field1 int
				Field2 string
			}{},
		},
		{
			name: "Zero value of pointer",
			zeroFunc: func() any {
				return zeroValue[*int]()
			},
			expected: (*int)(nil),
		},
		{
			name: "Zero value of slice",
			zeroFunc: func() any {
				return zeroValue[[]int]()
			},
			expected: []int(nil),
		},
		{
			name: "Zero value of map",
			zeroFunc: func() any {
				return zeroValue[map[string]int]()
			},
			expected: map[string]int(nil),
		},
		{
			name: "Zero value of bool",
			zeroFunc: func() any {
				return zeroValue[bool]()
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.zeroFunc()
			assert.EqualValues(t, tt.expected, actual, "Zero value mismatch for %s", tt.name)
		})
	}
}
