package evrblk

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func FromRpcError(err error) error {
	if st, ok := status.FromError(err); ok {
		details := make(map[string]string)

		//for _, d := range st.Details() {
		//	d
		//}

		switch st.Code() {
		case codes.OK:
			return nil

		case codes.DeadlineExceeded,
			codes.Canceled:
			return &Error{
				Message: st.Message(),
				Code:    Timeout,
				Details: details,
			}

		case codes.Aborted,
			codes.FailedPrecondition,
			codes.AlreadyExists,
			codes.InvalidArgument,
			codes.OutOfRange:
			return &Error{
				Message: st.Message(),
				Code:    InvalidRequest,
				Details: details,
			}

		case codes.Unknown,
			codes.Unimplemented,
			codes.Internal,
			codes.Unavailable,
			codes.DataLoss:
			return &Error{
				Message: st.Message(),
				Code:    InternalFailure,
				Details: details,
			}

		case codes.NotFound:
			return &Error{
				Message: st.Message(),
				Code:    NotFound,
				Details: details,
			}

		case codes.PermissionDenied:
			return &Error{
				Message: st.Message(),
				Code:    PermissionDenied,
				Details: details,
			}

		case codes.ResourceExhausted:
			return &Error{
				Message: st.Message(),
				Code:    ResourceExhausted,
				Details: details,
			}

		case codes.Unauthenticated:
			return &Error{
				Message: st.Message(),
				Code:    Unauthenticated,
				Details: details,
			}
		}
	}

	return &Error{
		Message: err.Error(),
		Code:    InternalFailure,
		Details: make(map[string]string),
	}
}
