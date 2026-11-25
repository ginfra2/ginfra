/*
Package base
Copyright (c) 2023, 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Various tools and helpers.  They should all be obvious in what they do.
*/
package base

import (
	"crypto/rand"
	"go/ast"
	"go/parser"
	"go/token"
	"math/big"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}

func GetRandomFileName(suffix string) string {

	return strconv.Itoa(mrand.Intn(428301510)) + "_" + suffix
}

const (
	passchars              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+[].:"
	passcharsclean         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codechars              = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alphachars             = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	COMMON_PASSWORD_LENGTH = 16
)

// GeneratePassword Get a password.
func GeneratePassword(length int) string {
	b := make([]byte, length)
	m := big.NewInt(int64(len(passchars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, m)
		if err != nil || n == nil {
			b[i] = passchars[mrand.Intn(len(passchars))]
		} else {
			b[i] = passchars[int(n.Int64())]
		}
	}
	return string(b)
}

// GeneratePasswordClean a clean password that wont require encoding.
func GeneratePasswordClean(length int) string {
	b := make([]byte, length)
	m := big.NewInt(int64(len(passcharsclean)))
	for i := range b {
		n, err := rand.Int(rand.Reader, m)
		if err != nil || n == nil {
			b[i] = passcharsclean[mrand.Intn(len(passcharsclean))]
		} else {
			b[i] = passcharsclean[int(n.Int64())]
		}
	}
	return string(b)
}

// GenerateCodeword generates a random alphanumeric string of the specified length using uppercase letters and numbers.
func GenerateCodeword(length int) string {
	b := make([]byte, length)
	m := big.NewInt(int64(len(codechars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, m)
		if err != nil || n == nil {
			b[i] = codechars[mrand.Intn(len(codechars))]
		} else {
			b[i] = codechars[int(n.Int64())]
		}
	}
	return string(b)
}

// GenerateAlphaword generates a random string of uppercase alphabetic characters with a specified length.
func GenerateAlphaword(length int) string {
	b := make([]byte, length)
	m := big.NewInt(int64(len(alphachars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, m)
		if err != nil || n == nil {
			b[i] = alphachars[mrand.Intn(len(alphachars))]
		} else {
			b[i] = alphachars[int(n.Int64())]
		}
	}
	return string(b)
}

// Truth checks if a string value represents a truthy or falsy value.
// It returns true if the value is truthy and false otherwise.
// A value is considered truthy if it is not an empty string and not equal
// to "f", "F", "false", or "0" (case-insensitive).
// The comparison for falsy values is case-sensitive.
func Truth(v string) bool {
	if v == "" || v == "f" || v == "F" || strings.ToLower(v) == "false" || v == "0" {
		return false
	}
	return true
}

// SnipId trims the given id to a maximum length of 12 characters if its length exceeds 12.
func SnipId(id string) string {
	if len(id) > 12 {
		return id[:ShortIdLength]
	}
	return id
}

// GetGoAST parses a Go source file and returns its abstract syntax tree and any associated error.
func GetGoAST(targetFile string) (f *ast.File, err error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, targetFile, nil, parser.AllErrors)
}

type GMutex struct {
	ch chan struct{}
}

func NewGMutex() *GMutex {
	return &GMutex{ch: make(chan struct{}, 1)}
}

func (mtx *GMutex) Lock(timeout time.Duration) bool {
	select {
	case <-time.After(timeout):
		return false
	case mtx.ch <- struct{}{}:
		return true
	}
}

func (mtx *GMutex) Unlock() {
	<-mtx.ch
}

func (mtx *GMutex) Locked() bool {
	return len(mtx.ch) > 0
}

func StringPtr(v string) *string { return &v }

func StringPtrOrNil(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func IntPtr(v int) *int { return &v }

func Ptr2StringOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func PtrInt2StringOrEmpty(v *int) string {
	if v == nil {
		return "nil"
	}
	return strconv.Itoa(*v)
}

// PtrInt2Int will return the value of the integer pointed to by v.  If v is nil, it will return 0.
func PtrInt2Int(v *int) int {
	if v == nil {
		return 0;
	}
	return *v
}

func PtrFloat2Float(v *float32) float32 {
	if v == nil {
		return 0;
	}
	return *v
}

func PtrBool2Bool(v *bool) bool {
	if v == nil {
		return false;
	}
	return *v
}

