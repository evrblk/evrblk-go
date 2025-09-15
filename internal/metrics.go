package internal

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	TotalRequestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "evrblk_client_requests_total",
		Help: "Total number of requests",
	}, []string{"service", "method"})
	FailedRequestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "evrblk_client_requests_failed",
		Help: "Number of failed requests",
	}, []string{"service", "method", "error"})
	RequestsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:                            "evrblk_client_request_duration_seconds",
		Help:                            "Request duration",
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: time.Hour,
	}, []string{"service", "method"})
)

func init() {
	prometheus.MustRegister(TotalRequestsCounter)
	prometheus.MustRegister(FailedRequestsCounter)
	prometheus.MustRegister(RequestsDuration)
}

func MetricLabelFromGrpcError(err error) string {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.OK:
			return ""

		case codes.DeadlineExceeded:
		case codes.Canceled:
			return "timeout"

		case codes.Aborted:
		case codes.FailedPrecondition:
		case codes.AlreadyExists:
		case codes.InvalidArgument:
		case codes.OutOfRange:
			return "invalid_request"

		case codes.Unknown:
		case codes.Unimplemented:
		case codes.Internal:
		case codes.Unavailable:
		case codes.DataLoss:
			return "internal"

		case codes.NotFound:
			return "not_found"

		case codes.PermissionDenied:
			return "permission_denied"

		case codes.ResourceExhausted:
			return "resource_exhausted"

		case codes.Unauthenticated:
			return "unauthenticated"
		}
	}

	return "internal"
}

func MeasureSince(o prometheus.Observer, t1 time.Time) {
	o.Observe(time.Since(t1).Seconds())
}
