/*
Package base
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Error facility.  This is an attempt to match errors to more modern structured logging schemes, including the logger
that this module uses.
*/
package base

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// #####################################################################################################################
// # ERRORS

// Flagst Common flags that are used to express what sort of system state is the result of the error.
type Flagst uint32

const ERRFLAG_SUBSYSTEM_FATAL = 0b1           // A fatal error that has spoiled a subsystem.
const ERRFLAG_SUBSYSTEM_SERIOUS = 0b10        // A recoverable error that has spoiled a subsystem
const ERRFLAG_NOT_FOUND = 0b100               // Something was not found during the operation.
const ERRFLAG_NOT_AUTHORIZED = 0b1000         // Operation is not authorized.
const ERRFLAG_RECOVERABLE = 0b10000           // Error should not cause lasting consequences.
const ERRFLAG_NOT_RECOVERABLE = 0b100000      // Error will cause lasting consequences.
const ERRFLAG_DUPLICATE = 0b1000000           // A duplicate caused the error.
const ERRFLAG_BAD_VALUE = 0b10000000          // Caused by bad value.
const ERRFLAG_BAD_ID = 0b100000000            // Bad id was given.
const ERRFLAG_INFORMATIVE_ONLY = 0b1000000000 // Information and the error should be disgarded at the most appropriate place.

type GAnnotation struct {
	Name  string
	Value any
}

func (gann *GAnnotation) PrintValue() string {
	return fmt.Sprintf("%v", gann.Value)
}

// GinfraError is a featured error for ginfra.  Given that errors in the underlying module (ginfra itself)
// are typically serious justifies shunting around more information about the error, including annotations, a parent
// error, and flags for additional context.   In a lot os up-stack use cases it may make less sense to use this.  I have
// used a similar scheme in other projects and found it dramatically improves both operations and bug resolution.
type GinfraError struct {
	error
	message     string
	parentErr   error
	annotations []GAnnotation

	// I admit I'm iffy about this one, so consider this a bleeding-edge tryout.  Since flag are meant to express
	// system state problems, they should be fairly rare, as most errors don't affact state.
	flags Flagst
}

// Error.  This implements the error interface got GinfraError.  It is not particularly pretty at this point but that
// can be improved.  Yes, I know it has both value and pointer receivers and that's not good.  The problem is most
// annotation data will be primitive and near primitive, while 'error' is an interface.
func (gerr *GinfraError) Error() string {
	var b strings.Builder

	b.WriteString("message: ")
	b.WriteString(gerr.message)

	if len(gerr.annotations) > 0 {
		b.WriteString(" | annotations:")
		for i, anno := range gerr.annotations {
			b.WriteString(" '")
			b.WriteString(anno.Name)
			b.WriteString("','")
			b.WriteString(anno.Value.(string))
			b.WriteString("'")
			if i < len(gerr.annotations)-1 {
				b.WriteString(",")
			}
		}
	}

	if gerr.parentErr != nil {
		b.WriteString(" | cause: ")
		b.WriteString(gerr.parentErr.Error())
		b.WriteString(" > ")
		b.WriteString(fmt.Sprintf("message %v", gerr.message))
	}

	return b.String()
}

// GetFlags retrieves the current value of the flags associated with the GinfraError instance.
func (gerr *GinfraError) GetFlags() Flagst {
	return gerr.flags
}

// GetAnnotations retrieves all annotations associated with the GinfraError instance as a slice of GAnnotation.
func (gerr *GinfraError) GetAnnotations() []GAnnotation {
	return gerr.annotations
}

// SeekAnnotation searches for an annotation in the GinfraError instance by its name (key) and returns a pointer to it.
// The annotations are not indexed and this is a linear search.
func (gerr *GinfraError) SeekAnnotation(key string) *GAnnotation {
	var r *GAnnotation

	for _, i := range gerr.annotations {
		if i.Name == key {
			r = &i
			break
		}
	}

	return r
}

// AddFlag adds the specified flag to the existing flags of the GinfraError and returns the updated GinfraError instance.
func (gerr *GinfraError) AddFlag(flag Flagst) *GinfraError {
	gerr.flags = gerr.flags + flag
	return gerr
}

// Annotate appends a new annotation with the specified name and value to the GinfraError and returns the updated instance.
func (gerr *GinfraError) Annotate(name string, value any) *GinfraError {
	a := GAnnotation{
		Name:  name,
		Value: value,
	}
	gerr.annotations = append(gerr.annotations, a)
	return gerr
}

func errorAnnotation(err error, num int) GAnnotation {
	return GAnnotation{
		Name:  fmt.Sprintf("Error # %d", num),
		Value: err.Error(),
	}
}

// Annotations adds multiple annotations to the GinfraError instance, mostly expecting name/value pairs or standalone
// errors (though the errors can be named too).
func (gerr *GinfraError) Annotations(anno ...any) *GinfraError {
	var (
		name string
		enum = 1
	)
	hasName := false
	for _, item := range anno {
		if hasName {
			switch ie := item.(type) {
			case []error:
				for i, ee := range ie {
					gerr.annotations = append(gerr.annotations, GAnnotation{
						Name:  name + "_" + strconv.Itoa(i+1),
						Value: fmt.Sprintf("%v", ee), // TODO break out all the type handling from print.go because this is just stupid.
					})
				}

			default:
				gerr.annotations = append(gerr.annotations, GAnnotation{
					Name:  name,
					Value: fmt.Sprintf("%v", item), // TODO break out all the type handling from print.go because this is just stupid.
				})
			}
			hasName = false

		} else {
			switch ie := item.(type) {
			case *GinfraError:
				gerr.annotations = append(gerr.annotations, errorAnnotation(ie, enum))
				enum = enum + 1

			// They have to be separate because of the type that is passed to the errorAnnotation.
			case error:
				gerr.annotations = append(gerr.annotations, errorAnnotation(ie, enum))
				enum = enum + 1

			case []error:
				for _, ee := range ie {
					gerr.annotations = append(gerr.annotations, errorAnnotation(ee, enum))
					enum = enum + 1
				}

			default:
				name = fmt.Sprintf("%v", item)
				hasName = true
			}

		}
	}

	if hasName {
		panic("Annotation name does not have associated value.")
	}

	return gerr
}

// NewGinfraError creates a new GinfraError instance with the specified message and no parent error or annotations.
func NewGinfraError(gmessage string) *GinfraError {
	return &GinfraError{
		message:   gmessage,
		parentErr: nil,
	}
}

// NewGinfraErrorChild creates a new child GinfraError with the given message and parent error.
// If err is nil, it will be ignored.  It returns the created GinfraError.
func NewGinfraErrorChild(gmessage string, err error) *GinfraError {
	e := GinfraError{
		message:   gmessage,
		parentErr: err,
	}
	if err != nil {
		if ee, ok := err.(*GinfraError); ok {
			e.flags = ee.flags
		}
	}
	return &e
}

// NewGinfraErrorA creates a new GinfraError with the specified message and annotations.
// It returns the newly created GinfraError instance.
func NewGinfraErrorA(gmessage string, annotations ...any) *GinfraError {
	e := GinfraError{
		message:   gmessage,
		parentErr: nil,
	}
	return e.Annotations(annotations...)
}

// NewGinfraErrorChildA creates a new child GinfraError with the given message, parent error and annotations.
// If err is nil, it will be ignored.
// It returns the created GinfraError.
func NewGinfraErrorChildA(gmessage string, err error, annotations ...any) *GinfraError {
	e := NewGinfraErrorA(gmessage, annotations...)
	if err != nil {
		e.parentErr = err
		if ee, ok := err.(*GinfraError); ok {
			e.flags = ee.flags
		}
	}
	return e
}

// NewGinfraErrorFlags creates a new GinfraError instance with the specified message and sets the given flags.
func NewGinfraErrorFlags(gmessage string, flags Flagst) *GinfraError {
	e := NewGinfraError(gmessage)
	e.flags = flags
	return e
}

// NewGinfraErrorChildFlags creates a new GinfraError with a specified message, parent error, and associated flags.
func NewGinfraErrorChildFlags(gmessage string, err error, flags Flagst) *GinfraError {
	e := NewGinfraErrorChild(gmessage, err)
	e.flags = flags
	return e
}

// NewGinfraErrorAFlags creates a new GinfraError with the specified message, flags, and optional annotations.
func NewGinfraErrorAFlags(gmessage string, flags Flagst, annotations ...any) *GinfraError {
	e := NewGinfraErrorA(gmessage, annotations...)
	e.flags = flags
	return e
}

// NewGinfraErrorChildAFlags creates a new GinfraError instance with a message, parent error, flags, and annotations.
func NewGinfraErrorChildAFlags(gmessage string, err error, flags Flagst, annotations ...any) *GinfraError {
	e := NewGinfraErrorChildA(gmessage, err, annotations...)
	e.flags = flags
	return e
}

// GetErrorFlags extracts and returns the error flags from a given error if it is of type GinfraError, otherwise returns 0.
func GetErrorFlags(err error) Flagst {
	if v, ok := err.(*GinfraError); ok {
		return v.GetFlags()
	}
	return 0
}

// GetErrorJSON get error as json.
func GetErrorJSON(err error) string {
	if v, ok := err.(*GinfraError); ok {
		return v.ErrorJson()
	}
	return "{ \"error\": \"" + err.Error() + "\" }"
}

// ErrorJson returns a JSON string representation of the error including message, annotations, flags, and parent error.
// If the parent error is also a GinfraError, it will be recursively expanded as JSON.
func (gerr *GinfraError) ErrorJson() string {
	type jsonError struct {
		Message     string            `json:"message"`
		Annotations map[string]string `json:"annotations,omitempty"`
		Flags       uint32            `json:"flags,omitempty"`
		Cause       interface{}       `json:"cause,omitempty"`
	}

	je := jsonError{
		Message: gerr.message,
		Flags:   uint32(gerr.flags),
	}

	if len(gerr.annotations) > 0 {
		je.Annotations = make(map[string]string)
		for _, anno := range gerr.annotations {
			je.Annotations[anno.Name] = anno.PrintValue()
		}
	}

	if gerr.parentErr != nil {
		// Check if parent is also a GinfraError and recursively expand it
		if parentGerr, ok := gerr.parentErr.(*GinfraError); ok {
			// Recursively build the parent error as a nested structure
			je.Cause = buildJsonErrorStruct(parentGerr)
		} else {
			// For non-GinfraError parents, just use the error string
			je.Cause = gerr.parentErr.Error()
		}
	}

	jsonBytes, err := json.Marshal(je)
	if err != nil {
		// Fallback to basic JSON if marshaling fails
		return fmt.Sprintf(`{"message":"%s","error":"failed to marshal JSON"}`, gerr.message)
	}

	return string(jsonBytes)
}

// buildJsonErrorStruct recursively builds the JSON structure for a GinfraError
func buildJsonErrorStruct(gerr *GinfraError) map[string]interface{} {
	result := map[string]interface{}{
		"message": gerr.message,
	}

	if gerr.flags > 0 {
		result["flags"] = uint32(gerr.flags)
	}

	if len(gerr.annotations) > 0 {
		annotations := make(map[string]string)
		for _, anno := range gerr.annotations {
			annotations[anno.Name] = anno.PrintValue()
		}
		result["annotations"] = annotations
	}

	if gerr.parentErr != nil {
		if parentGerr, ok := gerr.parentErr.(*GinfraError); ok {
			result["cause"] = buildJsonErrorStruct(parentGerr)
		} else {
			result["cause"] = gerr.parentErr.Error()
		}
	}

	return result
}
