package internal

import (
	"errors"
	"testing"

	evrblk "github.com/evrblk/evrblk-go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorFromRpcError_NilReturnsNil(t *testing.T) {
	assert.Nil(t, ErrorFromRpcError(nil))
}

func TestErrorFromRpcError_OKReturnsNil(t *testing.T) {
	assert.Nil(t, ErrorFromRpcError(status.New(codes.OK, "").Err()))
}

func TestErrorFromRpcError_CodeMapping(t *testing.T) {
	cases := []struct {
		grpc codes.Code
		want evrblk.ErrorCode
	}{
		{codes.DeadlineExceeded, evrblk.Timeout},
		{codes.Canceled, evrblk.Timeout},
		{codes.AlreadyExists, evrblk.AlreadyExists},
		{codes.Aborted, evrblk.InvalidRequest},
		{codes.FailedPrecondition, evrblk.InvalidRequest},
		{codes.InvalidArgument, evrblk.InvalidRequest},
		{codes.OutOfRange, evrblk.InvalidRequest},
		{codes.Unavailable, evrblk.Unavailable},
		{codes.Unknown, evrblk.InternalFailure},
		{codes.Unimplemented, evrblk.InternalFailure},
		{codes.Internal, evrblk.InternalFailure},
		{codes.DataLoss, evrblk.InternalFailure},
		{codes.NotFound, evrblk.NotFound},
		{codes.PermissionDenied, evrblk.PermissionDenied},
		{codes.ResourceExhausted, evrblk.ResourceExhausted},
		{codes.Unauthenticated, evrblk.Unauthenticated},
	}

	for _, c := range cases {
		t.Run(c.grpc.String(), func(t *testing.T) {
			err := ErrorFromRpcError(status.New(c.grpc, "msg").Err())

			var e *evrblk.Error
			require.ErrorAs(t, err, &e)
			assert.Equal(t, c.want, e.Code)
			assert.Equal(t, "msg", e.Message)
			assert.Empty(t, e.Details)
		})
	}
}

func TestErrorFromRpcError_PreservesCause(t *testing.T) {
	cause := status.New(codes.NotFound, "missing").Err()

	err := ErrorFromRpcError(cause)

	// The translated error still unwraps to the original transport error, so
	// errors.Is and the gRPC status helper keep working for debugging.
	assert.True(t, errors.Is(err, cause))
	assert.Equal(t, codes.NotFound, status.Code(errors.Unwrap(err)))
}

func TestErrorFromRpcError_ExtractsErrorInfoDetails(t *testing.T) {
	st, err := status.New(codes.AlreadyExists, "taken").WithDetails(&errdetails.ErrorInfo{
		Reason: "NAME_TAKEN",
		Domain: "grackle",
		Metadata: map[string]string{
			"namespace_name": "foo",
		},
	})
	require.NoError(t, err)

	var e *evrblk.Error
	require.ErrorAs(t, ErrorFromRpcError(st.Err()), &e)
	assert.Equal(t, evrblk.AlreadyExists, e.Code)
	assert.Equal(t, map[string]string{"namespace_name": "foo"}, e.Details)
}

func TestErrorFromRpcError_NonStatusFallback(t *testing.T) {
	cause := errors.New("not a grpc status error")

	err := ErrorFromRpcError(cause)

	var e *evrblk.Error
	require.ErrorAs(t, err, &e)
	assert.Equal(t, evrblk.InternalFailure, e.Code)
	assert.Equal(t, cause.Error(), e.Message)
	assert.True(t, errors.Is(err, cause))
}
