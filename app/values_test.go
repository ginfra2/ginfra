//go:build unit_test

/*
Package app
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Vlues tests.
*/

package app

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAppValues(t *testing.T) {
	assert.Contains(t, GetDebuggingLogPath(), PathDebuggingLog)
}
