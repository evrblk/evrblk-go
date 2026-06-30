package internal

import (
	evrblk "github.com/evrblk/evrblk-go"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorFromRpcError translates a gRPC transport error into the transport-neutral
// *evrblk.Error surfaced by the SDK. It returns nil for a nil error (and for a
// codes.OK status), so callers can pass the RPC result through unconditionally.
// The original error is retained as the cause for debugging via errors.Is/As.
func ErrorFromRpcError(err error) error {
	if err == nil {
		return nil
	}

	if st, ok := status.FromError(err); ok {
		details := detailsFromStatus(st)

		switch st.Code() {
		case codes.OK:
			return nil

		case codes.DeadlineExceeded,
			codes.Canceled:
			return evrblk.NewError(evrblk.Timeout, st.Message(), details, err)

		case codes.AlreadyExists:
			return evrblk.NewError(evrblk.AlreadyExists, st.Message(), details, err)

		case codes.Aborted,
			codes.FailedPrecondition,
			codes.InvalidArgument,
			codes.OutOfRange:
			return evrblk.NewError(evrblk.InvalidRequest, st.Message(), details, err)

		case codes.Unavailable:
			return evrblk.NewError(evrblk.Unavailable, st.Message(), details, err)

		case codes.Unknown,
			codes.Unimplemented,
			codes.Internal,
			codes.DataLoss:
			return evrblk.NewError(evrblk.InternalFailure, st.Message(), details, err)

		case codes.NotFound:
			return evrblk.NewError(evrblk.NotFound, st.Message(), details, err)

		case codes.PermissionDenied:
			return evrblk.NewError(evrblk.PermissionDenied, st.Message(), details, err)

		case codes.ResourceExhausted:
			return evrblk.NewError(evrblk.ResourceExhausted, st.Message(), details, err)

		case codes.Unauthenticated:
			return evrblk.NewError(evrblk.Unauthenticated, st.Message(), details, err)
		}
	}

	return evrblk.NewError(evrblk.InternalFailure, err.Error(), nil, err)
}

// detailsFromStatus extracts structured key/value context attached server-side
// as errdetails.ErrorInfo. Returns nil when there is nothing to extract.
func detailsFromStatus(st *status.Status) map[string]string {
	var details map[string]string
	for _, d := range st.Details() {
		info, ok := d.(*errdetails.ErrorInfo)
		if !ok {
			continue
		}
		for k, v := range info.GetMetadata() {
			if details == nil {
				details = make(map[string]string)
			}
			details[k] = v
		}
	}
	return details
}
