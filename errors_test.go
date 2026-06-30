package evrblk

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	t.Run("allocates an empty details map when nil", func(t *testing.T) {
		e := NewError(NotFound, "nope", nil, nil)
		assert.NotNil(t, e.Details)
		assert.Empty(t, e.Details)
		assert.Equal(t, NotFound, e.Code)
		assert.Equal(t, "nope", e.Message)
	})

	t.Run("retains provided details and cause", func(t *testing.T) {
		cause := errors.New("boom")
		e := NewError(InvalidRequest, "bad", map[string]string{"field": "name"}, cause)
		assert.Equal(t, map[string]string{"field": "name"}, e.Details)
		assert.Same(t, cause, e.Unwrap())
	})
}

func TestError_Error(t *testing.T) {
	cases := map[ErrorCode]string{
		InternalFailure:   "internal failure: x",
		Timeout:           "timeout: x",
		InvalidRequest:    "invalid request: x",
		Unauthenticated:   "unauthenticated: x",
		PermissionDenied:  "permission denied: x",
		NotFound:          "not found: x",
		ResourceExhausted: "resource exhausted: x",
		Unavailable:       "unavailable: x",
		AlreadyExists:     "already exists: x",
	}
	for code, want := range cases {
		assert.Equal(t, want, (&Error{Code: code, Message: "x"}).Error())
	}

	// Unknown codes fall back to internal failure.
	assert.Equal(t, "internal failure: x", (&Error{Code: ErrorCode(9999), Message: "x"}).Error())
}

func TestError_UnwrapReachesCause(t *testing.T) {
	sentinel := errors.New("root cause")
	err := error(NewError(NotFound, "missing", nil, sentinel))

	assert.True(t, errors.Is(err, sentinel))
	assert.Same(t, sentinel, errors.Unwrap(err))

	// Wrapping the *Error further still lets errors.As recover it.
	wrapped := fmt.Errorf("call failed: %w", err)
	var e *Error
	assert.True(t, errors.As(wrapped, &e))
	assert.Equal(t, NotFound, e.Code)
}

func TestCodeOf(t *testing.T) {
	assert.Equal(t, Ok, CodeOf(nil))
	assert.Equal(t, Ok, CodeOf(errors.New("plain")))
	assert.Equal(t, NotFound, CodeOf(NewError(NotFound, "missing", nil, nil)))

	// Works through wrapping.
	wrapped := fmt.Errorf("wrap: %w", NewError(Unavailable, "down", nil, nil))
	assert.Equal(t, Unavailable, CodeOf(wrapped))
}

func TestIsHelpers(t *testing.T) {
	type matcher struct {
		fn   func(error) bool
		code ErrorCode
	}
	matchers := []matcher{
		{IsInternalFailure, InternalFailure},
		{IsTimeout, Timeout},
		{IsInvalidRequest, InvalidRequest},
		{IsUnauthenticated, Unauthenticated},
		{IsPermissionDenied, PermissionDenied},
		{IsNotFound, NotFound},
		{IsResourceExhausted, ResourceExhausted},
		{IsUnavailable, Unavailable},
		{IsAlreadyExists, AlreadyExists},
	}

	for _, m := range matchers {
		err := error(NewError(m.code, "x", nil, nil))
		// The matching helper returns true.
		assert.True(t, m.fn(err), "matcher for code %d should match", m.code)

		// Every other matcher returns false for this error.
		for _, other := range matchers {
			if other.code == m.code {
				continue
			}
			assert.False(t, other.fn(err), "matcher for code %d should not match code %d", other.code, m.code)
		}
	}

	// Nil and non-evrblk errors never match.
	assert.False(t, IsNotFound(nil))
	assert.False(t, IsNotFound(errors.New("plain")))
}
