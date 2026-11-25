//go:build unit_test

/*
Package base
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Test runners.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package base

import (
	"testing"
	"time"
)

func TestTruth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: false,
		},
		{
			name:     "Lowercase f",
			input:    "f",
			expected: false,
		},
		{
			name:     "Uppercase F",
			input:    "F",
			expected: false,
		},
		{
			name:     "Fully lowercase false",
			input:    "false",
			expected: false,
		},
		{
			name:     "Mixed case false",
			input:    "fAlSe",
			expected: false,
		},
		{
			name:     "Zero as string",
			input:    "0",
			expected: false,
		},
		{
			name:     "Random other string",
			input:    "R2D2",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Truth(tt.input)
			if actual != tt.expected {
				t.Errorf("Truth(%s) = %v, expected %v", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestGeneratePassword(t *testing.T) {
	// define a structure for test case
	type test struct {
		name   string
		length int
	}

	// initialize test cases
	var tests = []test{
		{"Zero", 0},
		{"Positive", 10},
		{"Large number", 5000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := GeneratePassword(tt.length)

			if len(got) != tt.length {
				t.Errorf("GeneratePassword() = %v, want %v", len(got), tt.length)
			}

			// check characters of the generated password
			for _, c := range got {
				found := false
				for _, passchar := range passchars {
					if c == int32(passchar) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GeneratePassword() = %v, contains invalid characters", got)
					break
				}
			}
		})
	}
}

func TestGenerateCodeword(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "ZeroLength",
			length: 0,
		},
		{
			name:   "LengthOne",
			length: 1,
		},
		{
			name:   "LengthFive",
			length: 5,
		},
		{
			name:   "LengthTen",
			length: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateCodeword(tt.length)
			if len(result) != tt.length {
				t.Errorf("Expected length = %d, got %d", tt.length, len(result))
			}
			for _, char := range result {
				if !charInCodechars(char) {
					t.Errorf("Invalid character in codeword: %v", char)
				}
			}
		})
	}
}

func charInCodechars(char rune) bool {
	for _, c := range codechars {
		if char == c {
			return true
		}
	}
	return false
}

func TestGenerateAlphaword(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "ZeroLength",
			length: 0,
		},
		{
			name:   "SmallLength",
			length: 1,
		},
		{
			name:   "MediumLength",
			length: 10,
		},
		{
			name:   "LargeLength",
			length: 5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateAlphaword(tt.length)
			if len(result) != tt.length {
				t.Errorf("Expected length = %d, got %d", tt.length, len(result))
			}
			for _, char := range result {
				if !charInAlphachars(char) {
					t.Errorf("Generated string contains invalid character: %v", char)
				}
			}
		})
	}
}

func charInAlphachars(char int32) bool {
	for _, c := range alphachars {
		if char == c {
			return true
		}
	}
	return false
}

func TestSnipId(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "empty_id",
			id:   "",
			want: "",
		},
		{
			name: "short_id",
			id:   "123456",
			want: "123456",
		},
		{
			name: "exact_length_id",
			id:   "123456789012",
			want: "123456789012",
		},
		{
			name: "long_id",
			id:   "12345678901234567890",
			want: "123456789012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnipId(tt.id); got != tt.want {
				t.Errorf("SnipId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLock(t *testing.T) {
	tests := []struct {
		name       string
		timeout    time.Duration
		shouldLock bool
		initFunc   func() *GMutex
	}{
		{
			name:       "SuccessfulLock",
			timeout:    50 * time.Millisecond,
			shouldLock: true,
			initFunc: func() *GMutex {
				return &GMutex{ch: make(chan struct{}, 1)}
			},
		},
		{
			name:       "TimeoutLock",
			timeout:    10 * time.Millisecond,
			shouldLock: false,
			initFunc: func() *GMutex {
				mtx := &GMutex{ch: make(chan struct{}, 1)}
				mtx.Lock(5 * time.Millisecond) // Block the lock
				return mtx
			},
		},
		{
			name:       "NilMutexLock",
			timeout:    10 * time.Millisecond,
			shouldLock: false,
			initFunc: func() *GMutex {
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtx := tt.initFunc()
			var got bool
			if mtx != nil {
				got = mtx.Lock(tt.timeout)
			}

			if got != tt.shouldLock {
				t.Errorf("GMutex.Lock() = %v, want %v", got, tt.shouldLock)
			}

			// Cleanup: unlock in case of locked state
			if got && mtx != nil {
				mtx.Unlock()
			}
		})
	}
}
