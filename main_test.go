/*
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Entrypoint test.
*/
package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestEntry(t *testing.T) {

	t.Run("Good", func(t *testing.T) {
		os.Args = []string{
			"ginfra",
			"tool",
			"info",
		}
		assert.Equal(t, 0, Main(true))
	})

	t.Run("Bad", func(t *testing.T) {
		os.Args = []string{
			"ginfra",
			"bork",
		}
		assert.NotEqual(t, 0, Main(true))
	})

}
