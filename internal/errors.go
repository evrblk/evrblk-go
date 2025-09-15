package internal

import (
	evrblk "github.com/evrblk/evrblk-go"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorFromRpcError(err error) error {
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
			return &evrblk.Error{
				Message: st.Message(),
				Code:    evrblk.Timeout,
				Details: details,
			}

		case codes.Aborted,
			codes.FailedPrecondition,
			codes.AlreadyExists,
			codes.InvalidArgument,
			codes.OutOfRange:
			return &evrblk.Error{
				Message: st.Message(),
				Code:    evrblk.InvalidRequest,
				Details: details,
			}

		case codes.Unknown,
			codes.Unimplemented,
			codes.Internal,
			codes.Unavailable,
			codes.DataLoss:
			return &evrblk.Error{
				Message: st.Message(),
				Code:    evrblk.InternalFailure,
				Details: details,
			}

		case codes.NotFound:
			return &evrblk.Error{
				Message: st.Message(),
				Code:    evrblk.NotFound,
				Details: details,
			}

		case codes.PermissionDenied:
			return &evrblk.Error{
				Message: st.Message(),
				Code:    evrblk.PermissionDenied,
				Details: details,
			}

		case codes.ResourceExhausted:
			return &evrblk.Error{
				Message: st.Message(),
				Code:    evrblk.ResourceExhausted,
				Details: details,
			}

		case codes.Unauthenticated:
			return &evrblk.Error{
				Message: st.Message(),
				Code:    evrblk.Unauthenticated,
				Details: details,
			}
		}
	}

	return &evrblk.Error{
		Message: err.Error(),
		Code:    evrblk.InternalFailure,
		Details: make(map[string]string),
	}
}
