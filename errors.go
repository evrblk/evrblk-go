package evrblk

import (
	"errors"
	"fmt"
)

type ErrorCode uint32

const (
	// Ok is the zero value of ErrorCode and means "no error". It never appears
	// on a non-nil *Error; it is the value CodeOf returns for a nil error or an
	// error that is not (and does not wrap) an *Error.
	Ok ErrorCode = iota
	// InternalFailure means the service hit an unexpected error that is not the
	// caller's fault (e.g. an internal, unknown, unimplemented, or data-loss
	// condition). Retrying may or may not help.
	InternalFailure
	// Timeout means the request did not complete in time or was canceled before
	// the service produced a result.
	Timeout
	// InvalidRequest means the request was malformed or not valid for the
	// current state of the resource (e.g. a bad argument, a failed
	// precondition, an out-of-range value, or an aborted optimistic-concurrency
	// conflict). The caller should fix the request before retrying.
	InvalidRequest
	// Unauthenticated means the request had missing or invalid credentials.
	Unauthenticated
	// PermissionDenied means the caller is authenticated but not authorized to
	// perform the request.
	PermissionDenied
	// NotFound means the requested resource does not exist.
	NotFound
	// ResourceExhausted means a quota or limit has been reached (e.g. the
	// maximum number of namespaces).
	ResourceExhausted
	// Unavailable means the service was temporarily unreachable or not ready.
	// It is transient and the request is generally safe to retry.
	Unavailable
	// AlreadyExists means a resource with the same client-provided identity
	// (e.g. a name) already exists.
	AlreadyExists
)

// Error is the transport-neutral error surfaced by every SDK call. Callers
// should branch on Code (or use the Is* helpers) rather than inspecting the
// underlying transport error. The original transport error, when present, is
// retained as the cause and reachable through errors.Is / errors.As.
type Error struct {
	Message string
	Code    ErrorCode
	Details map[string]string

	cause error
}

// NewError builds an *Error. details may be nil (an empty map is allocated).
// cause is the underlying error this was translated from, and may be nil.
func NewError(code ErrorCode, message string, details map[string]string, cause error) *Error {
	if details == nil {
		details = make(map[string]string)
	}
	return &Error{
		Message: message,
		Code:    code,
		Details: details,
		cause:   cause,
	}
}

func (e *Error) Error() string {
	switch e.Code {
	case InternalFailure:
		return fmt.Sprintf("internal failure: %s", e.Message)
	case Timeout:
		return fmt.Sprintf("timeout: %s", e.Message)
	case InvalidRequest:
		return fmt.Sprintf("invalid request: %s", e.Message)
	case Unauthenticated:
		return fmt.Sprintf("unauthenticated: %s", e.Message)
	case PermissionDenied:
		return fmt.Sprintf("permission denied: %s", e.Message)
	case NotFound:
		return fmt.Sprintf("not found: %s", e.Message)
	case ResourceExhausted:
		return fmt.Sprintf("resource exhausted: %s", e.Message)
	case Unavailable:
		return fmt.Sprintf("unavailable: %s", e.Message)
	case AlreadyExists:
		return fmt.Sprintf("already exists: %s", e.Message)
	default:
		return fmt.Sprintf("internal failure: %s", e.Message)
	}
}

// Unwrap returns the underlying transport error this Error was translated from,
// so errors.Is / errors.As keep working down to the original cause. It returns
// nil when there is no cause.
func (e *Error) Unwrap() error {
	return e.cause
}

// CodeOf returns the ErrorCode of err if it is (or wraps) an *Error, and Ok
// otherwise. A nil error also returns Ok.
func CodeOf(err error) ErrorCode {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return Ok
}

func hasCode(err error, code ErrorCode) bool {
	return CodeOf(err) == code
}

// Is* helpers report whether err is (or wraps) an *Error with the given code.

func IsInternalFailure(err error) bool   { return hasCode(err, InternalFailure) }
func IsTimeout(err error) bool           { return hasCode(err, Timeout) }
func IsInvalidRequest(err error) bool    { return hasCode(err, InvalidRequest) }
func IsUnauthenticated(err error) bool   { return hasCode(err, Unauthenticated) }
func IsPermissionDenied(err error) bool  { return hasCode(err, PermissionDenied) }
func IsNotFound(err error) bool          { return hasCode(err, NotFound) }
func IsResourceExhausted(err error) bool { return hasCode(err, ResourceExhausted) }
func IsUnavailable(err error) bool       { return hasCode(err, Unavailable) }
func IsAlreadyExists(err error) bool     { return hasCode(err, AlreadyExists) }
