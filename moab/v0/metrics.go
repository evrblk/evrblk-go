package moab

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics below are all scoped to a single MoabConsumer by the queue_name
// label. They're distinct from the generic per-RPC metrics in the internal
// package (evrblk_client_requests_total and friends, recorded by
// client_gen.go for every Dequeue/ReportStatus call) - those measure the RPC
// layer across every evrblk-go client; these measure MoabConsumer's own
// task-processing loop.
var (
	consumerTasksProcessedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "moab_consumer_tasks_processed_total",
		Help: "Total number of tasks a MoabConsumer's HandlerFunc has returned for, by outcome (result=succeeded|failed).",
	}, []string{"queue_name", "result"})

	consumerTaskPanicsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "moab_consumer_task_panics_total",
		Help: "Total number of HandlerFunc calls that panicked (recovered and counted/reported as failed).",
	}, []string{"queue_name"})

	consumerHandlerDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:                            "moab_consumer_handler_duration_seconds",
		Help:                            "Time spent inside HandlerFunc, by outcome (result=succeeded|failed).",
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: time.Hour,
	}, []string{"queue_name", "result"})

	consumerBufferDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:                            "moab_consumer_buffer_duration_seconds",
		Help:                            "Time a dequeued task spent waiting in MoabConsumer's internal buffer before a worker started handling it.",
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: time.Hour,
	}, []string{"queue_name"})

	consumerInFlightTasks = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "moab_consumer_in_flight_tasks",
		Help: "Current number of tasks a MoabConsumer holds a lease on - waiting in its buffer or being handled.",
	}, []string{"queue_name"})

	consumerOldestInFlightTaskAgeSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "moab_consumer_oldest_in_flight_task_age_seconds",
		Help: "How long ago the longest-held in-flight task was dequeued, or 0 if none are in flight.",
	}, []string{"queue_name"})

	consumerEmptyDequeueResponsesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "moab_consumer_empty_dequeue_responses_total",
		Help: "Total number of Dequeue calls that returned no tasks.",
	}, []string{"queue_name"})

	consumerDequeueErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "moab_consumer_dequeue_errors_total",
		Help: "Total number of Dequeue calls that returned an error.",
	}, []string{"queue_name"})

	consumerHeartbeatRejectionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "moab_consumer_heartbeat_rejections_total",
		Help: "Total number of heartbeats rejected because this consumer no longer holds the task's lease (it was reclaimed and possibly redelivered elsewhere).",
	}, []string{"queue_name"})

	consumerDroppedStatusReportsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "moab_consumer_dropped_status_reports_total",
		Help: "Total number of buffered task status reports dropped to bound memory during a prolonged ReportStatus outage.",
	}, []string{"queue_name"})
)

func init() {
	prometheus.MustRegister(consumerTasksProcessedTotal)
	prometheus.MustRegister(consumerTaskPanicsTotal)
	prometheus.MustRegister(consumerHandlerDuration)
	prometheus.MustRegister(consumerBufferDuration)
	prometheus.MustRegister(consumerInFlightTasks)
	prometheus.MustRegister(consumerOldestInFlightTaskAgeSeconds)
	prometheus.MustRegister(consumerEmptyDequeueResponsesTotal)
	prometheus.MustRegister(consumerDequeueErrorsTotal)
	prometheus.MustRegister(consumerHeartbeatRejectionsTotal)
	prometheus.MustRegister(consumerDroppedStatusReportsTotal)
}
