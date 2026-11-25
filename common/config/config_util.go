/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Configuration utilities as well as defined values.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package config

import (
	"math/rand"
	"strings"
)

// #####################################################################################################################
// # UTILITIES

// IfFileResolvePath resolves a file URI by optionally appending a given path and ensures correct "file://" formatting.
func IfFileResolvePath(uri, path string) string {
	if strings.HasPrefix(uri, "file://") && !strings.HasPrefix(uri, "file:///") {
		if len(path) > 0 {
			if path[0] == '/' {
				uri = "file://" + path + "/" + uri[7:]
			} else {
				uri = "file:///" + path + "/" + uri[7:]
			}

		} else {
			uri = "file://" + uri[7:]
		}

		idx := strings.Index(uri[5:], ":")
		if idx >= 1 {
			uri = "file:///" + uri[5+idx+2:]
		}
	}
	return uri
}

// Flagify converts a non-empty flag string into a CLI-style flag by prepending it with "--".
// If the input flag is empty, it returns the flag unchanged.
func Flagify(flag string) string {
	if len(flag) > 0 {
		return "--" + flag
	}
	return flag
}

// CommandBuilder constructs and returns a slice of non-empty strings from the provided input arguments.
func CommandBuilder(args ...string) []string {
	r := []string{}
	for _, a := range args {
		if len(a) > 0 {
			r = append(r, a)
		}
	}
	return r
}

// GetNewAuthToken generates a new authentication token of fixed length based on the provided read-only flag.
// If ro is true, the token starts with the read-only flag; otherwise, it starts with 'A'.
// The remaining characters are randomly generated uppercase letters.
// Returns the generated authentication token as a string.
func GetNewAuthToken(ro bool) string {
	var result [AUTH_TOKEN_LENGTH]byte
	for i := 1; i < AUTH_TOKEN_LENGTH; i++ {
		result[i] = byte(rand.Intn(90-65) + 65)
	}
	if ro {
		result[0] = CONFIG_RO_FLAG[0] // Always lead with the flag
	} else {
		result[0] = 65 // Always lead with an A
	}

	return string(result[:])
}
