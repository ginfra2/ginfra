/*
Package testutil
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Utilities for tests.
*/
package testutil

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type ScanPack struct {
	text      string
	offset    int
	lastFound string
	lfOffset  int
}

func NewScanPack(t string, err error) (*ScanPack, error) {
	return &ScanPack{
		text:   t,
		offset: 0,
	}, err
}

func NewScanPackB(b []byte, err error) (*ScanPack, error) {
	return &ScanPack{
		text:   string(b),
		offset: 0,
	}, err
}

// Seek searches for the given text in the scan pack's text starting from the current offset.
// If the text is found, it updates the current offset.
// Returns the index of the first occurrence of the text if found, or -1 if not found.
func (sp *ScanPack) Seek(text string) int {
	if len(sp.text) <= sp.offset {
		return -1
	}
	if i := strings.Index(sp.text[sp.offset:], text); i >= 0 {
		o := sp.offset + i
		sp.offset = o + 1
		sp.lastFound = text
		sp.lfOffset = o
		return o
	}
	return -1
}

func (sp *ScanPack) Seeks(texts ...string) int {
	var i int
	for _, s := range texts {
		i = sp.Seek(s)
		if i < 0 {
			break
		}
	}
	return i
}

func TestAssertValue(t *testing.T, err error, left, right any) {
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.Equal(t, left, right, "Unexpected value.")
	}
}

// TestCheckMapValue validates the values in a nested map[string]interface{} based on a given
// path. It checks if the value at the provided path matches the expected value. It returns the
// inner map[string]interface{} at the last key of the path. If any value in the path is missing,
// an error is raised. The function takes in a testing.T instance, the map to be checked, the
// expected value, and the path keys.  If the target value is another map, it will return it.
// If fail is true, it expects to fail.
func TestCheckMapValue(t *testing.T, fail bool, m map[string]interface{}, value any, path ...string) map[string]interface{} {
	var r map[string]interface{}

	for i, n := range path {
		if i == (len(path) - 1) {
			if v, ok := m[n]; ok {
				switch v.(type) {
				case map[string]interface{}:
					r = v.(map[string]interface{})
					if fail {
						assert.NotEqual(t, fmt.Sprintf("%v", value), n, "Name should not be found.")
					} else {
						assert.Equal(t, fmt.Sprintf("%v", value), n, "Name not correct.")
					}

				default:
					if fail {
						assert.NotEqual(t, fmt.Sprintf("%v", value), fmt.Sprintf("%v", v),
							fmt.Sprintf("Value should not be correct.  Expected %v, got %v.  path=%v", value, v, path))
					} else {
						assert.Equal(t, fmt.Sprintf("%v", value), fmt.Sprintf("%v", v),
							fmt.Sprintf("Value not correct.  Expected %v, got %v.  path=%v", value, v, path))
					}
				}
			} else {
				if !fail {
					t.Error("Value not found.", path)
				}
			}

		} else {
			if mn, ok := m[n]; ok {
				// If the assertion fails, things are pretty messed up anyway,
				m = mn.(map[string]interface{})
			} else {
				if !fail {
					t.Error("Path not found.", path)
				}

			}

		}
	}
	return r
}

func FreshTargetDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return errors.New("Could not remove previous files (" + path + ").  Clean out the tmp/test/ dir.")
	}
	return os.MkdirAll(path, os.ModePerm)
}

// SkipFlakyCheck brutal hack.  This is because of some truly strange behavior on one of my bigger servers.
func SkipFlakyCheck() bool {
	if _, err := os.Stat("/SKIPFLAKY"); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(base.GetGinfraHome(), "SKIPFLAKY")); err == nil {
		return true
	}
	return false
}
