//go:build unit_test

/*
Package base
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Errors tests.
*/
package base

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGinfraError(t *testing.T) {
	tests := []struct {
		name string
		err  GinfraError
		want string
	}{
		{
			name: "EmptyGinfraError",
			err:  GinfraError{},
			want: "message: ",
		},
		{
			name: "GinfraErrorWithMessage",
			err:  GinfraError{message: "test"},
			want: "message: test",
		},
		{
			name: "GinfraErrorWithAnnotation",
			err: GinfraError{
				message: "test",
				annotations: []GAnnotation{
					{
						Name:  "key",
						Value: "value",
					},
				},
			},
			want: "message: test | annotations: 'key','value'",
		},
		{
			name: "GinfraErrorWithParentError",
			err: GinfraError{
				message:   "test",
				parentErr: errors.New("parent error"),
			},
			want: "message: test | cause: parent error > message test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestErrorAnnotations(t *testing.T) {
	testCases := []struct {
		name        string
		annotations []any
		wantPanic   bool
		expectVal   string
	}{
		{
			name:        "WithoutAnnotations",
			annotations: []any{},
			wantPanic:   false,
			expectVal:   "WithoutAnnotations",
		},
		{
			name:        "SingleAnnotationPair",
			annotations: []any{"key1", "value1"},
			wantPanic:   false,
			expectVal:   "'key1','value1'",
		},
		{
			name:        "MultipleAnnotationPairs",
			annotations: []any{"key1", "value1", "key2", "value2"},
			wantPanic:   false,
			expectVal:   "'key2','value2'",
		},
		{
			name:        "NoMatchingAnnotationPair",
			annotations: []any{"key1"},
			wantPanic:   true,
			expectVal:   "",
		},
		{
			name:        "AnnotateNamedArrayOfErrors",
			annotations: []any{"key1", []error{errors.New("error1"), errors.New("error2"), errors.New("error3")}},
			wantPanic:   false,
			expectVal:   "'key1_3','error3'",
		},
		{
			name:        "AnnotateUnnamedArrayOfErrors",
			annotations: []any{[]error{errors.New("error1"), errors.New("error2"), errors.New("error3")}},
			wantPanic:   false,
			expectVal:   "'Error # 3','error3'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewGinfraError(tc.name)

			func() {
				defer func() {
					if r := recover(); r != nil && !tc.wantPanic {
						t.Errorf("Annotations method panicked unexpectedly: %v", r)
					} else if r == nil && tc.wantPanic {
						t.Error("Annotations method expected to panic, but it did not")
					}
				}()
				_ = err.Annotations(tc.annotations...)
			}()

			// No panic expected, validate annotations
			s := err.Error()
			if s == "" {
				return
			}

			if !tc.wantPanic {
				if tc.expectVal != "" {
					assert.Contains(t, err.Error(), tc.expectVal)
				}
			}
		})
	}
}

func TestGinfraErrorAdd(t *testing.T) {
	e := NewGinfraError("Error")
	err := e.Annotate("Bork", "Blarg")
	assert.Error(t, err)

	s := e.Error()
	assert.Less(t, 0, strings.Index(s, "Bork"))
	assert.Less(t, 0, strings.Index(s, "Blarg"))
	s = err.Error()
	assert.Less(t, 0, strings.Index(s, "Bork"))
	assert.Less(t, 0, strings.Index(s, "Blarg"))

}

func TestGinfraErrorFlags(t *testing.T) {
	oe := errors.New("y")

	e := NewGinfraErrorFlags("x", 1)
	assert.Equal(t, 1, int(e.GetFlags()))
	_ = e.AddFlag(2)
	assert.Equal(t, 3, int(e.GetFlags()))

	e = NewGinfraErrorChildFlags("e", oe, 8)
	assert.Equal(t, 8, int(e.GetFlags()))

	e = NewGinfraErrorAFlags("x", 16, "a", "b")
	assert.Equal(t, 16, int(e.GetFlags()))

	e = NewGinfraErrorChildAFlags("x", oe, 32, "a", "b")
	assert.Equal(t, 32, int(e.GetFlags()))

	assert.Equal(t, 32, int(GetErrorFlags(e)))
	assert.Equal(t, 0, int(GetErrorFlags(oe)))
}

func TestGinfraErrorFeatures(t *testing.T) {
	t.Run("print annotation", func(t *testing.T) {
		g := GAnnotation{
			Value: "bork",
		}
		g.PrintValue()
	})

	t.Run("annotation features", func(t *testing.T) {
		g := NewGinfraErrorA("Error", "anno1", "ANNO1", "anno2", "ANNO2", "anno3", "ANNO3")

		ga := g.GetAnnotations()
		assert.Len(t, ga, 3)

		gs1 := g.SeekAnnotation("anno2")
		assert.Equal(t, "anno2", gs1.Name)
		assert.Equal(t, "ANNO2", gs1.Value)

		gs2 := g.SeekAnnotation("xxxxx")
		assert.Nil(t, gs2)

		gn := NewGinfraError("ginfraerror")

		g = g.Annotations(errors.New("plainerror"), gn)
		gns1 := g.SeekAnnotation("Error # 1")
		assert.Equal(t, "Error # 1", gns1.Name)
		assert.Equal(t, "plainerror", gns1.Value)
		gns2 := g.SeekAnnotation("Error # 2")
		assert.Equal(t, "Error # 2", gns2.Name)
		assert.Equal(t, "message: ginfraerror", gns2.Value)

	})
}

func TestErrorJson(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := NewGinfraErrorFlags("test message", ERRFLAG_BAD_VALUE)
		jsonStr := err.ErrorJson()
		assert.Contains(t, jsonStr, `"message":"test message"`)
		assert.Contains(t, jsonStr, `"flags":128`) // ERRFLAG_BAD_VALUE is 0b10000000 = 128
	})

	t.Run("with annotations", func(t *testing.T) {
		err := NewGinfraErrorA("anno message", "key1", "val1", "key2", "val2")
		jsonStr := err.ErrorJson()
		assert.Contains(t, jsonStr, `"message":"anno message"`)
		assert.Contains(t, jsonStr, `"annotations":{"key1":"val1","key2":"val2"}`)
	})

	t.Run("with plain error cause", func(t *testing.T) {
		parent := errors.New("root cause")
		err := NewGinfraErrorChild("child message", parent)
		jsonStr := err.ErrorJson()
		assert.Contains(t, jsonStr, `"message":"child message"`)
		assert.Contains(t, jsonStr, `"cause":"root cause"`)
	})

	t.Run("with nested GinfraError cause", func(t *testing.T) {
		grandparent := errors.New("grandparent")
		parent := NewGinfraErrorChildA("parent message", grandparent, "pk", "pv")
		parent.AddFlag(ERRFLAG_NOT_FOUND)
		err := NewGinfraErrorChild("child message", parent)

		jsonStr := err.ErrorJson()
		assert.Contains(t, jsonStr, `"message":"child message"`)
		assert.Contains(t, jsonStr, `"cause":{"annotations":{"pk":"pv"},"cause":"grandparent","flags":4,"message":"parent message"}`)
	})
}
