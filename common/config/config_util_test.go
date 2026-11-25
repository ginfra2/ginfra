package config

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestIfFileResolvePath(t *testing.T) {
	testCases := []struct {
		name      string
		uri       string
		path      string
		expResult string
	}{
		{
			name:      "no_file_scheme_and_no_path",
			uri:       "myFile",
			path:      "",
			expResult: "myFile",
		},
		{
			name:      "file_scheme_and_no_path",
			uri:       "file://myFile",
			path:      "",
			expResult: "file://myFile",
		},
		{
			name:      "file_scheme_and_path",
			uri:       "file://myFile",
			path:      "/myPath",
			expResult: "file:///myPath/myFile",
		},
		{
			name:      "file_path_with_triple_slashes",
			uri:       "file:///myFile",
			path:      "/myPath",
			expResult: "file:///myFile",
		},
		{
			name:      "file_path_no_slash",
			uri:       "file://myFile",
			path:      "myPath",
			expResult: "file:///myPath/myFile",
		},
		{
			name:      "additional colon.",
			uri:       "file://myFile",
			path:      "C:/myPath", // Needs some work.
			expResult: "file:///myPath/myFile",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualResult := IfFileResolvePath(tc.uri, tc.path)
			if actualResult != tc.expResult {
				t.Errorf("expected %s but got %s", tc.expResult, actualResult)
			}
		})
	}
}

func TestFlagify(t *testing.T) {
	testCases := []struct {
		name     string
		flag     string
		expected string
	}{
		{
			"Empty flag",
			"",
			"",
		},
		{
			"Non-empty flag",
			"flag",
			"--flag",
		},
		{
			"Flag with leading/trailing whitespaces",
			"  flag  ",
			"--  flag  ",
		},
		{
			"Flag with special characters",
			"$%@!",
			"--$%@!",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Flagify(tc.flag)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestCommandBuilder(t *testing.T) {
	testCases := map[string]struct {
		input    []string
		expected []string
	}{
		"empty input":                   {input: []string{}, expected: []string{}},
		"single empty str":              {input: []string{""}, expected: []string{}},
		"multiple empty str":            {input: []string{"", "", ""}, expected: []string{}},
		"single non-empty str":          {input: []string{"test"}, expected: []string{"test"}},
		"multiple non-empty str":        {input: []string{"test1", "test2"}, expected: []string{"test1", "test2"}},
		"mixed empty and non-empty str": {input: []string{"", "test", ""}, expected: []string{"test"}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			output := CommandBuilder(tc.input...)
			if !reflect.DeepEqual(output, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, output)
			}
		})
	}
}

func TestGetNewAuthToken(t *testing.T) {
	t.Run("Auth Token Read Write", func(t *testing.T) {
		ans := GetNewAuthToken(false)
		assert.Equal(t, AUTH_TOKEN_LENGTH, len(ans))
		assert.NotEqual(t, CONFIG_RO_FLAG[0], ans[0])
	})

	t.Run("Auth Token Read Only", func(t *testing.T) {
		ans := GetNewAuthToken(true)
		assert.Equal(t, AUTH_TOKEN_LENGTH, len(ans))
		assert.Equal(t, CONFIG_RO_FLAG[0], ans[0])
	})

}
