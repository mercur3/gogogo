package common

import (
	"errors"
	"fmt"
)

type ErrKind string

const (
	ErrNotFound      ErrKind = "not_found"
	ErrAlreadyExists ErrKind = "already_exists"
	ErrUnknown       ErrKind = "uknown"
)

// TypedErr represents an error with a concrete error "type" and an optional parent.
type TypedErr struct {
	Kind   ErrKind
	Msg    string
	Parent error
}

// NewTypedErr constructs a TypedErr and guarantees ErrType is never nil.
func NewTypedErr(kind ErrKind, msg string, parent error) *TypedErr {
	return &TypedErr{
		Kind:   kind,
		Msg:    msg,
		Parent: parent,
	}
}

func (e *TypedErr) Error() string {
	if e.Parent != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Parent)
	}
	return e.Msg
}

// Is implements errors.Is semantics for TypedErr. Checks Parent against target.
func (e *TypedErr) Is(target error) bool {
	if target == nil {
		return false
	}

	// // If target is a TypedErr, compare their ErrType and Parent.
	// if te, ok := target.(*TypedErr); ok {
	//     // Both ErrType and Parent should match (using errors.Is for transitive matching).
	//     return errors.Is(e.ErrType, te.ErrType) && errors.Is(e.Parent, te.Parent)
	// }

	// // Otherwise, check if target matches ErrType or Parent.
	return errors.Is(e.Parent, target)
}

func (e *TypedErr) Unwrap() error {
	return e.Parent
}
