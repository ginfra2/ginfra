package testutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScanPackSeek(t *testing.T) {
	tests := []struct {
		name              string
		initText          string
		initOffset        int
		seekText          string
		expectedReturn    int
		expectedOffset    int
		expectedLastFound string
		expectedLfOffset  int
	}{
		{"empty text", "", 0, "anything", -1, 0, "", 0},
		{"empty seek", "some text", 0, "", 0, 1, "", 0},
		{"found at start", "findme at the beginning", 0, "findme", 0, 1, "findme", 0},
		{"found at middle", "found the findme at middle", 0, "findme", 10, 11, "findme", 10},
		{"found at end", "at the end findme", 0, "findme", 11, 12, "findme", 11},
		{"not found", "it's not here", 0, "findme", -1, 0, "", 0},
		{"start offset", "findme once, findme twice", 10, "findme", 13, 14, "findme", 13},
		{"offset exceeds", "findme but miss", 16, "findme", -1, 16, "", 0},
		{"offset equals length", "findme but offset too large", 30, "findme", -1, 30, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ScanPack instance
			sp, _ := NewScanPack(tt.initText, nil)
			// Set initial offset
			sp.offset = tt.initOffset

			// Perform seek
			res := sp.Seek(tt.seekText)

			// Assertions
			assert.Equal(t, tt.expectedReturn, res)
			assert.Equal(t, tt.expectedOffset, sp.offset)
			assert.Equal(t, tt.expectedLastFound, sp.lastFound)
			assert.Equal(t, tt.expectedLfOffset, sp.lfOffset)
		})
	}
}

func Test_TestCheckMapValue(t *testing.T) {
	var testCases = []struct {
		desc  string
		fail  bool
		m     map[string]interface{}
		value any
		path  []string
	}{
		{
			desc:  "Value not found",
			fail:  true,
			m:     map[string]interface{}{"key": "value"},
			value: "notfound",
			path:  []string{"key"},
		},
		{
			desc:  "Correct value",
			fail:  false,
			m:     map[string]interface{}{"key": "value"},
			value: "value",
			path:  []string{"key"},
		},
		{
			desc:  "Incorrect value",
			fail:  true,
			m:     map[string]interface{}{"key": "value"},
			value: "incorrect",
			path:  []string{"key"},
		},
		{
			desc:  "path not found",
			fail:  true,
			m:     map[string]interface{}{"key": "value"},
			value: "any",
			path:  []string{"path"},
		},
		{
			desc:  "nested path value",
			fail:  false,
			m:     map[string]interface{}{"nested": map[string]interface{}{"key": "value"}},
			value: "value",
			path:  []string{"nested", "key"},
		},
		{
			desc:  "nested path not found",
			fail:  false,
			m:     map[string]interface{}{"nested": map[string]interface{}{"key": "value"}},
			value: "value",
			path:  []string{"nested", "key"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			TestCheckMapValue(t, tc.fail, tc.m, tc.value, tc.path...)
		})
	}
}
