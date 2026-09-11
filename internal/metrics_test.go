package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMetricLabelFromGrpcError_CodeMapping(t *testing.T) {
	cases := []struct {
		grpc codes.Code
		want string
	}{
		{codes.OK, ""},
		{codes.DeadlineExceeded, "timeout"},
		{codes.Canceled, "timeout"},
		{codes.Aborted, "invalid_request"},
		{codes.FailedPrecondition, "invalid_request"},
		{codes.AlreadyExists, "invalid_request"},
		{codes.InvalidArgument, "invalid_request"},
		{codes.OutOfRange, "invalid_request"},
		{codes.Unknown, "internal"},
		{codes.Unimplemented, "internal"},
		{codes.Internal, "internal"},
		{codes.Unavailable, "internal"},
		{codes.DataLoss, "internal"},
		{codes.NotFound, "not_found"},
		{codes.PermissionDenied, "permission_denied"},
		{codes.ResourceExhausted, "resource_exhausted"},
		{codes.Unauthenticated, "unauthenticated"},
	}

	for _, c := range cases {
		t.Run(c.grpc.String(), func(t *testing.T) {
			got := MetricLabelFromGrpcError(status.New(c.grpc, "msg").Err())
			assert.Equal(t, c.want, got)
		})
	}
}

func TestMetricLabelFromGrpcError_NonStatusFallback(t *testing.T) {
	assert.Equal(t, "internal", MetricLabelFromGrpcError(assertError("boom")))
}

type assertError string

func (e assertError) Error() string { return string(e) }
