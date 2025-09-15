package evrblk

import (
	"fmt"
)

type ErrorCode uint32

const (
	Ok ErrorCode = iota
	InternalFailure
	Timeout
	InvalidRequest
	Unauthenticated
	PermissionDenied
	NotFound
	ResourceExhausted
)

type Error struct {
	Message string
	Code    ErrorCode
	Details map[string]string
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
	default:
		return fmt.Sprintf("internal failure: %s", e.Message)
	}
}
