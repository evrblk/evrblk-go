package moab

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeMoabApi implements only the two methods MoabConsumer actually calls.
// Any other method panics (nil embedded interface), which is fine since the
// tests below never trigger them.
type fakeMoabApi struct {
	MoabApi

	mu              sync.Mutex
	tasks           []*Task
	reported        []*ReportStatusRequestEntry
	dequeueRequests []*DequeueRequest

	// reportStatusHook, if set, replaces the default always-succeeds-with-
	// RESULT_OK behavior entirely - for tests simulating RPC-level failures
	// or specific per-entry results. It's responsible for calling
	// recordReported itself if it wants reportedCount/reportedEntries to
	// reflect what it did.
	reportStatusHook func(req *ReportStatusRequest) (*ReportStatusResponse, error)
}

func (f *fakeMoabApi) Dequeue(ctx context.Context, req *DequeueRequest) (*DequeueResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.dequeueRequests = append(f.dequeueRequests, req)

	n := int(req.BatchSize)
	if n > len(f.tasks) {
		n = len(f.tasks)
	}
	out := f.tasks[:n]
	f.tasks = f.tasks[n:]
	return &DequeueResponse{Tasks: out}, nil
}

func (f *fakeMoabApi) ReportStatus(ctx context.Context, req *ReportStatusRequest) (*ReportStatusResponse, error) {
	f.mu.Lock()
	hook := f.reportStatusHook
	f.mu.Unlock()

	if hook != nil {
		return hook(req)
	}

	f.recordReported(req.Entries)

	entries := make([]*ReportStatusResponseEntry, len(req.Entries))
	for i, e := range req.Entries {
		entries[i] = &ReportStatusResponseEntry{TaskId: e.TaskId, Result: ReportStatusResponseEntry_RESULT_OK}
	}
	return &ReportStatusResponse{Entries: entries}, nil
}

func (f *fakeMoabApi) recordReported(entries []*ReportStatusRequestEntry) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reported = append(f.reported, entries...)
}

func (f *fakeMoabApi) reportedCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.reported)
}

func (f *fakeMoabApi) reportedEntries() []*ReportStatusRequestEntry {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*ReportStatusRequestEntry(nil), f.reported...)
}

func (f *fakeMoabApi) dequeueRequestsSnapshot() []*DequeueRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*DequeueRequest(nil), f.dequeueRequests...)
}

// TestStart_GracefulShutdown_DoesNotRaceLastSend pins down the actual race
// this feature closes: ctx is cancelled at the exact moment a worker is
// about to report the completion of the task it's mid-handling. Before the
// fix, the worker raced an unconditional ctx.Done() branch against the send
// to statusCh, so - even though the status reporter was still very much
// alive and receiving - the send could lose that race and the completion
// would be silently dropped. Repeating this many times makes that a near
// certainty if the race is still there, while the fix makes it impossible
// by construction (the reporter can't stop until every worker, including
// this one, has returned).
func TestStart_GracefulShutdown_DoesNotRaceLastSend(t *testing.T) {
	const iterations = 200

	for i := 0; i < iterations; i++ {
		api := &fakeMoabApi{tasks: []*Task{{Id: "the-task", Attempts: 1}}}

		consumer, err := NewMoabConsumer(api, "q", slog.Default(), ConsumerConfig{
			NumWorkers:          4,
			NumPollers:          1,
			ReportBatchSize:     100,
			ReportFlushInterval: time.Hour,
			PollingInterval:     time.Millisecond,
			DequeueTimeout:      time.Second,
			ReportStatusTimeout: time.Second,
			KeepAliveTimeout:    3 * time.Hour, // far longer than the test runs; heartbeating is irrelevant here
		})
		if err != nil {
			t.Fatalf("NewMoabConsumer: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())

		handling := make(chan struct{})
		release := make(chan struct{})
		var handlingClosedOnce sync.Once

		startDone := make(chan struct{})
		go func() {
			defer close(startDone)
			consumer.Start(ctx, func(ctx context.Context, task *Task) error {
				handlingClosedOnce.Do(func() { close(handling) })
				<-release
				return nil
			})
		}()

		<-handling // the worker is now blocked inside the handler

		// Cancel and release back-to-back so the worker's post-handler send
		// races ctx.Done() becoming ready as tightly as this test can force.
		cancel()
		close(release)

		select {
		case <-startDone:
		case <-time.After(2 * time.Second):
			t.Fatalf("iteration %d: Start did not return within 2s of ctx cancellation", i)
		}

		if got := api.reportedCount(); got != 1 {
			t.Fatalf("iteration %d: expected 1 reported status after graceful shutdown, got %d", i, got)
		}
	}
}

// TestStart_ShutdownTimeout_AbandonsHungHandler checks the other half of the
// contract: a HandlerFunc that ignores ctx cancellation and never returns
// must not hang Start forever. Once ShutdownTimeout elapses, Start gives up
// waiting on it and returns anyway.
func TestStart_ShutdownTimeout_AbandonsHungHandler(t *testing.T) {
	api := &fakeMoabApi{tasks: []*Task{{Id: "the-task", Attempts: 1}}}

	const shutdownTimeout = 50 * time.Millisecond
	consumer, err := NewMoabConsumer(api, "q", slog.Default(), ConsumerConfig{
		NumWorkers:          2,
		NumPollers:          1,
		PollingInterval:     time.Millisecond,
		DequeueTimeout:      time.Second,
		ReportStatusTimeout: time.Second,
		ShutdownTimeout:     shutdownTimeout,
		KeepAliveTimeout:    3 * time.Hour, // far longer than the test runs; heartbeating is irrelevant here
	})
	if err != nil {
		t.Fatalf("NewMoabConsumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	handling := make(chan struct{})
	release := make(chan struct{})
	t.Cleanup(func() { close(release) }) // let the abandoned handler finish so it isn't left running past the test

	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		consumer.Start(ctx, func(ctx context.Context, task *Task) error {
			close(handling)
			<-release // ignores ctx cancellation on purpose
			return nil
		})
	}()

	<-handling
	cancel()

	select {
	case <-startDone:
	case <-time.After(shutdownTimeout + 2*time.Second):
		t.Fatal("Start did not return after ShutdownTimeout elapsed")
	}
}

// TestStart_ShutdownTimeoutDoesNotFireBeforeCtxCancellation guards against
// the wind-down select racing ShutdownTimeout from the moment Start is
// called, rather than from ctx actually being cancelled: any run that simply
// lasts longer than ShutdownTimeout would then trip the "abandon" path while
// pollLoop/workerLoop are still healthy, permanently killing the status
// reporter and heartbeat loop (via drained) 30s (the default) into an
// otherwise-normal run, with no way to recover for the rest of it. Setting
// ShutdownTimeout far shorter than how long ctx actually stays alive
// reproduces that window directly: every task must still get reported long
// after ShutdownTimeout has elapsed, because ctx was never cancelled.
func TestStart_ShutdownTimeoutDoesNotFireBeforeCtxCancellation(t *testing.T) {
	const numTasks = 30

	tasks := make([]*Task, numTasks)
	for i := range tasks {
		tasks[i] = &Task{Id: fmt.Sprintf("task-%d", i), Attempts: 1}
	}
	api := &fakeMoabApi{tasks: tasks}

	const shutdownTimeout = 20 * time.Millisecond
	consumer, err := NewMoabConsumer(api, "q", slog.Default(), ConsumerConfig{
		NumWorkers:          2,
		NumPollers:          1,
		ReportBatchSize:     1,
		ReportFlushInterval: time.Millisecond,
		PollingInterval:     time.Millisecond,
		DequeueTimeout:      time.Second,
		ReportStatusTimeout: time.Second,
		ShutdownTimeout:     shutdownTimeout,
		KeepAliveTimeout:    3 * time.Hour, // far longer than the test runs; heartbeating is irrelevant here
	})
	if err != nil {
		t.Fatalf("NewMoabConsumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		consumer.Start(ctx, func(ctx context.Context, task *Task) error {
			// Each task takes longer than ShutdownTimeout to process in
			// aggregate (numTasks/NumWorkers handler calls, each sleeping a
			// bit), so plenty of tasks are still being handled well after
			// ShutdownTimeout has elapsed while ctx is still very much alive.
			time.Sleep(5 * time.Millisecond)
			return nil
		})
	}()

	// All numTasks must get reported eventually - including the ones handled
	// long after ShutdownTimeout elapsed - because ctx is never cancelled
	// during this wait.
	deadline := time.After(2 * time.Second)
	for api.reportedCount() != numTasks {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for all tasks to be reported, got %d/%d - the status reporter likely stopped early", api.reportedCount(), numTasks)
		case <-time.After(time.Millisecond):
		}
	}

	cancel()

	select {
	case <-startDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return within 2s of ctx cancellation")
	}
}

// TestStart_GracefulShutdown_FlushesPendingSubBatch reproduces the shutdown
// race this test guards against: cancelling ctx while a handful of tasks
// (fewer than ReportBatchSize) are in flight must not drop their completion
// reports, and Start must not return until they've been flushed.
func TestStart_GracefulShutdown_FlushesPendingSubBatch(t *testing.T) {
	const numTasks = 5

	tasks := make([]*Task, numTasks)
	for i := range tasks {
		tasks[i] = &Task{Id: fmt.Sprintf("task-%d", i), Attempts: 1}
	}
	api := &fakeMoabApi{tasks: tasks}

	consumer, err := NewMoabConsumer(api, "q", slog.Default(), ConsumerConfig{
		NumWorkers: 8,
		NumPollers: 2,
		// Larger than numTasks and a long interval, so the only way these
		// get reported is via the shutdown-time flush being exercised here.
		ReportBatchSize:     100,
		ReportFlushInterval: time.Hour,
		PollingInterval:     5 * time.Millisecond,
		DequeueTimeout:      time.Second,
		ReportStatusTimeout: time.Second,
		KeepAliveTimeout:    3 * time.Hour, // far longer than the test runs; heartbeating is irrelevant here
	})
	if err != nil {
		t.Fatalf("NewMoabConsumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var handled atomic.Int32
	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		consumer.Start(ctx, func(ctx context.Context, task *Task) error {
			handled.Add(1)
			return nil
		})
	}()

	// Give the pollers/workers a chance to dequeue and process everything.
	deadline := time.After(2 * time.Second)
	for handled.Load() != numTasks {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for tasks to be handled, handled=%d", handled.Load())
		case <-time.After(time.Millisecond):
		}
	}

	cancel()

	select {
	case <-startDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return within 2s of ctx cancellation")
	}

	if got := api.reportedCount(); got != numTasks {
		t.Fatalf("expected %d reported statuses after graceful shutdown, got %d", numTasks, got)
	}
}

// TestStart_SendsKeepAliveTimeoutOnDequeueAndReportStatus pins down that
// ConsumerConfig.KeepAliveTimeout is actually sent to the server - on every
// Dequeue call (as the lease this consumer wants for the batch) and on
// every ReportStatus entry (as the lease it wants if the report turns out
// to be a heartbeat) - not just used locally for the stale-report warning.
func TestStart_SendsKeepAliveTimeoutOnDequeueAndReportStatus(t *testing.T) {
	api := &fakeMoabApi{tasks: []*Task{{Id: "the-task", Attempts: 1}}}

	const keepAliveTimeout = 45 * time.Second
	consumer, err := NewMoabConsumer(api, "q", slog.Default(), ConsumerConfig{
		NumWorkers:          2,
		NumPollers:          1,
		PollingInterval:     time.Millisecond,
		DequeueTimeout:      time.Second,
		ReportStatusTimeout: time.Second,
		KeepAliveTimeout:    keepAliveTimeout,
	})
	if err != nil {
		t.Fatalf("NewMoabConsumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var handled atomic.Int32
	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		consumer.Start(ctx, func(ctx context.Context, task *Task) error {
			handled.Add(1)
			return nil
		})
	}()

	deadline := time.After(2 * time.Second)
	for handled.Load() != 1 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for the task to be handled")
		case <-time.After(time.Millisecond):
		}
	}

	cancel()

	select {
	case <-startDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return within 2s of ctx cancellation")
	}

	wantSeconds := int64(keepAliveTimeout / time.Second)

	for _, req := range api.dequeueRequestsSnapshot() {
		if req.KeepaliveTimeoutInSeconds != wantSeconds {
			t.Fatalf("Dequeue request has KeepaliveTimeoutInSeconds=%d, want %d", req.KeepaliveTimeoutInSeconds, wantSeconds)
		}
	}

	reported := api.reportedEntries()
	if len(reported) != 1 {
		t.Fatalf("expected 1 reported status, got %d", len(reported))
	}
	if reported[0].KeepaliveTimeoutInSeconds != wantSeconds {
		t.Fatalf("ReportStatus entry has KeepaliveTimeoutInSeconds=%d, want %d", reported[0].KeepaliveTimeoutInSeconds, wantSeconds)
	}
}

// TestRunHandlerWithHeartbeat_RenewsLeaseWhileHandlerRuns pins down the
// actual point of KeepAliveTimeout: a handler that runs for multiple
// heartbeat intervals gets its lease renewed - standalone
// ReportStatus(IN_PROGRESS) calls, not batched with the final report -
// before it ever reports SUCCEEDED.
// TestHeartbeatLoop_RenewsLeaseOfTaskStillQueuedInBuffer pins down the fix
// this test is named for: a task can sit dequeued-but-unclaimed in bufCh
// for a while (every worker busy with something else), and heartbeatLoop
// must renew its lease during that wait - not just once some worker
// eventually gets around to it. Forces exactly that ordering with a single
// worker kept busy on one task while a second sits in the buffer untouched.
func TestHeartbeatLoop_RenewsLeaseOfTaskStillQueuedInBuffer(t *testing.T) {
	const (
		busyTaskId   = "busy-task"
		queuedTaskId = "queued-task"
	)
	api := &fakeMoabApi{tasks: []*Task{
		{Id: busyTaskId, Attempts: 1},
		{Id: queuedTaskId, Attempts: 1},
	}}

	const keepAliveTimeout = 100 * time.Millisecond // heartbeatInterval = 50ms, sweep = 12.5ms
	consumer, err := NewMoabConsumer(api, "q", slog.Default(), ConsumerConfig{
		NumWorkers:          1, // only one worker: it claims busyTaskId first, leaving queuedTaskId stuck in bufCh
		NumPollers:          1,
		DequeueBatchSize:    2,
		ReportBatchSize:     1,
		ReportFlushInterval: 10 * time.Millisecond,
		PollingInterval:     time.Millisecond,
		DequeueTimeout:      10 * time.Millisecond,
		ReportStatusTimeout: 10 * time.Millisecond,
		KeepAliveTimeout:    keepAliveTimeout,
	})
	if err != nil {
		t.Fatalf("NewMoabConsumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	busyTaskStarted := make(chan struct{})
	release := make(chan struct{})

	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		consumer.Start(ctx, func(ctx context.Context, task *Task) error {
			if task.Id == busyTaskId {
				close(busyTaskStarted)
				<-release
			}
			return nil
		})
	}()

	<-busyTaskStarted

	// While the one worker is still stuck on busyTaskId, give heartbeatLoop
	// several sweep intervals to renew queuedTaskId purely because it's
	// sitting in the buffer - no worker has touched it yet.
	time.Sleep(150 * time.Millisecond)

	var queuedTaskHeartbeats int
	for _, e := range api.reportedEntries() {
		if e.Status != ReportStatusRequestEntry_STATUS_IN_PROGRESS {
			t.Fatalf("neither task's handler has returned yet, so no final report should exist; got status %v for %s", e.Status, e.TaskId)
		}
		if e.TaskId == queuedTaskId {
			queuedTaskHeartbeats++
		}
	}
	if queuedTaskHeartbeats < 2 {
		t.Fatalf("expected at least 2 heartbeats for the task still sitting in the buffer, got %d", queuedTaskHeartbeats)
	}

	close(release)

	succeededCount := func() int {
		n := 0
		for _, e := range api.reportedEntries() {
			if e.Status == ReportStatusRequestEntry_STATUS_SUCCEEDED {
				n++
			}
		}
		return n
	}

	deadline := time.After(2 * time.Second)
	for succeededCount() < 2 {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for both tasks to report SUCCEEDED, got %d", succeededCount())
		case <-time.After(time.Millisecond):
		}
	}

	cancel()
	select {
	case <-startDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return within 2s of ctx cancellation")
	}
}

func TestRunHandlerWithHeartbeat_RenewsLeaseWhileHandlerRuns(t *testing.T) {
	api := &fakeMoabApi{tasks: []*Task{{Id: "the-task", Attempts: 1}}}
	logs := &recordingHandler{}

	const keepAliveTimeout = 100 * time.Millisecond // heartbeatInterval = 50ms
	consumer, err := NewMoabConsumer(api, "q", slog.New(logs), ConsumerConfig{
		NumWorkers:          2,
		NumPollers:          1,
		ReportBatchSize:     1,
		ReportFlushInterval: 10 * time.Millisecond,
		PollingInterval:     time.Millisecond,
		DequeueTimeout:      10 * time.Millisecond,
		ReportStatusTimeout: 10 * time.Millisecond,
		KeepAliveTimeout:    keepAliveTimeout,
	})
	if err != nil {
		t.Fatalf("NewMoabConsumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		consumer.Start(ctx, func(ctx context.Context, task *Task) error {
			// Long enough, relative to the 50ms heartbeat interval, to
			// guarantee at least two heartbeats land before it returns.
			time.Sleep(220 * time.Millisecond)
			return nil
		})
	}()

	deadline := time.After(2 * time.Second)
	for api.reportedCount() == 0 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for the final status report")
		case <-time.After(time.Millisecond):
		}
	}

	cancel()
	select {
	case <-startDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return within 2s of ctx cancellation")
	}

	reported := api.reportedEntries()

	var heartbeats, finals int
	for _, e := range reported {
		switch e.Status {
		case ReportStatusRequestEntry_STATUS_IN_PROGRESS:
			heartbeats++
			if e.KeepaliveTimeoutInSeconds != int64(keepAliveTimeout/time.Second) {
				t.Fatalf("heartbeat has KeepaliveTimeoutInSeconds=%d, want %d", e.KeepaliveTimeoutInSeconds, int64(keepAliveTimeout/time.Second))
			}
		case ReportStatusRequestEntry_STATUS_SUCCEEDED:
			finals++
		}
	}

	if heartbeats < 2 {
		t.Fatalf("expected at least 2 heartbeats before completion, got %d (entries: %d)", heartbeats, len(reported))
	}
	if finals != 1 {
		t.Fatalf("expected exactly 1 final SUCCEEDED report, got %d", finals)
	}
	// The final report must be the last entry - a heartbeat racing in after
	// the handler already returned would be a bug (runHandlerWithHeartbeat
	// must stop heartbeating the instant the handler is done).
	if reported[len(reported)-1].Status != ReportStatusRequestEntry_STATUS_SUCCEEDED {
		t.Fatalf("expected the final report to be the last entry sent, got status %v last", reported[len(reported)-1].Status)
	}

	// The whole point of heartbeating is that a task actively kept alive
	// this way must never trip the "exceeded keepalive timeout" staleness
	// warning - that warning must track the most recently renewed lease,
	// not the very first one from the original Dequeue.
	for _, r := range logs.records() {
		if r.Level == slog.LevelWarn {
			t.Fatalf("unexpected warning despite successful heartbeats: %q", r.Message)
		}
	}
}

// recordingHandler is a minimal slog.Handler that just remembers every
// record it receives, for tests that need to assert on what was logged.
type recordingHandler struct {
	mu   sync.Mutex
	recs []slog.Record
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.recs = append(h.recs, r)
	return nil
}

func (h *recordingHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *recordingHandler) WithGroup(_ string) slog.Handler      { return h }

func (h *recordingHandler) records() []slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]slog.Record(nil), h.recs...)
}

func TestConsumerConfig_Validate(t *testing.T) {
	// safe is a config that comfortably satisfies every cross-field check
	// on its own; each test case perturbs exactly one field away from it
	// rather than sharing a base that might itself sit on a boundary.
	safe := func() ConsumerConfig {
		return ConsumerConfig{
			KeepAliveTimeout:    10 * time.Second, // heartbeatInterval = 5s
			ReportFlushInterval: 2 * time.Second,
			ReportStatusTimeout: 2 * time.Second,
			DequeueTimeout:      2 * time.Second,
		}
	}

	tests := []struct {
		name      string
		configure func() ConsumerConfig
		wantErr   bool
	}{
		{
			name: "KeepAliveTimeout zero - rejected, no zero-value escape hatch",
			configure: func() ConsumerConfig {
				c := safe()
				c.KeepAliveTimeout = 0
				return c
			},
			wantErr: true,
		},
		{
			name: "KeepAliveTimeout negative - rejected",
			configure: func() ConsumerConfig {
				c := safe()
				c.KeepAliveTimeout = -time.Second
				return c
			},
			wantErr: true,
		},
		{
			name:      "sane config with KeepAliveTimeout set",
			configure: safe,
			wantErr:   false,
		},
		{
			name: "ReportFlushInterval at half of KeepAliveTimeout - rejected",
			configure: func() ConsumerConfig {
				c := safe()
				c.ReportFlushInterval = 5 * time.Second
				return c
			},
			wantErr: true,
		},
		{
			name: "ReportFlushInterval just under half of KeepAliveTimeout - accepted",
			configure: func() ConsumerConfig {
				c := safe()
				c.ReportFlushInterval = 4999 * time.Millisecond
				return c
			},
			wantErr: false,
		},
		{
			name: "ReportStatusTimeout at half of KeepAliveTimeout - rejected",
			configure: func() ConsumerConfig {
				c := safe()
				c.ReportStatusTimeout = 5 * time.Second
				return c
			},
			wantErr: true,
		},
		{
			name: "DequeueTimeout at KeepAliveTimeout - rejected",
			configure: func() ConsumerConfig {
				c := safe()
				c.DequeueTimeout = 10 * time.Second
				return c
			},
			wantErr: true,
		},
		{
			name: "DequeueTimeout just under KeepAliveTimeout - accepted",
			configure: func() ConsumerConfig {
				c := safe()
				c.DequeueTimeout = 9999 * time.Millisecond
				return c
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.configure().Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReportStatusLoop_RetriesFailedFlushInsteadOfDroppingIt pins down the
// bug this closes: an RPC-level ReportStatus failure (network blip, server
// unavailable) must not make the batch vanish - the next flush attempt must
// retry the exact same entries until one actually succeeds.
func TestReportStatusLoop_RetriesFailedFlushInsteadOfDroppingIt(t *testing.T) {
	api := &fakeMoabApi{tasks: []*Task{{Id: "the-task", Attempts: 1}}}

	var attempts atomic.Int32
	api.reportStatusHook = func(req *ReportStatusRequest) (*ReportStatusResponse, error) {
		if attempts.Add(1) <= 2 {
			return nil, fmt.Errorf("simulated transient failure")
		}
		api.recordReported(req.Entries)
		entries := make([]*ReportStatusResponseEntry, len(req.Entries))
		for i, e := range req.Entries {
			entries[i] = &ReportStatusResponseEntry{TaskId: e.TaskId, Result: ReportStatusResponseEntry_RESULT_OK}
		}
		return &ReportStatusResponse{Entries: entries}, nil
	}

	consumer, err := NewMoabConsumer(api, "q", slog.Default(), ConsumerConfig{
		NumWorkers:          2,
		NumPollers:          1,
		ReportBatchSize:     1,
		ReportFlushInterval: 5 * time.Millisecond,
		PollingInterval:     time.Millisecond,
		DequeueTimeout:      10 * time.Millisecond,
		ReportStatusTimeout: 10 * time.Millisecond,
		KeepAliveTimeout:    time.Hour, // far longer than the test runs; heartbeating is irrelevant here
	})
	if err != nil {
		t.Fatalf("NewMoabConsumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		consumer.Start(ctx, func(ctx context.Context, task *Task) error { return nil })
	}()

	deadline := time.After(2 * time.Second)
	for api.reportedCount() == 0 {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for the retried report to succeed (attempts so far: %d)", attempts.Load())
		case <-time.After(time.Millisecond):
		}
	}

	if got := attempts.Load(); got < 3 {
		t.Fatalf("expected at least 3 ReportStatus attempts (2 failures + 1 success), got %d", got)
	}
	if got := api.reportedCount(); got != 1 {
		t.Fatalf("expected exactly 1 recorded report once the retry succeeded, got %d", got)
	}

	cancel()
	select {
	case <-startDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return within 2s of ctx cancellation")
	}
}

// TestSendHeartbeatBatch_StopsHeartbeatingRejectedTask pins down the other
// half of processing ReportStatus's per-entry results correctly: once a
// heartbeat comes back non-OK, this consumer no longer holds a valid lease
// on that task, so it must stop sending it more heartbeats (they'd just be
// rejected again) rather than retrying forever.
func TestSendHeartbeatBatch_StopsHeartbeatingRejectedTask(t *testing.T) {
	api := &fakeMoabApi{tasks: []*Task{{Id: "the-task", Attempts: 1}}}
	logs := &recordingHandler{}

	var heartbeats atomic.Int32
	api.reportStatusHook = func(req *ReportStatusRequest) (*ReportStatusResponse, error) {
		entries := make([]*ReportStatusResponseEntry, len(req.Entries))
		for i, e := range req.Entries {
			result := ReportStatusResponseEntry_RESULT_OK
			if e.Status == ReportStatusRequestEntry_STATUS_IN_PROGRESS {
				heartbeats.Add(1)
				// Reject every heartbeat: the lease is never actually held.
				result = ReportStatusResponseEntry_RESULT_STALE_ATTEMPT
			}
			entries[i] = &ReportStatusResponseEntry{TaskId: e.TaskId, Result: result}
		}
		api.recordReported(req.Entries)
		return &ReportStatusResponse{Entries: entries}, nil
	}

	const keepAliveTimeout = 40 * time.Millisecond // heartbeatInterval = 20ms, sweep = 5ms
	consumer, err := NewMoabConsumer(api, "q", slog.New(logs), ConsumerConfig{
		NumWorkers:          1,
		NumPollers:          1,
		ReportFlushInterval: 5 * time.Millisecond,
		PollingInterval:     time.Millisecond,
		DequeueTimeout:      5 * time.Millisecond,
		ReportStatusTimeout: 5 * time.Millisecond,
		KeepAliveTimeout:    keepAliveTimeout,
	})
	if err != nil {
		t.Fatalf("NewMoabConsumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release := make(chan struct{})
	t.Cleanup(func() { close(release) })

	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		consumer.Start(ctx, func(ctx context.Context, task *Task) error {
			<-release // stays "in flight" long enough for several heartbeat sweeps
			return nil
		})
	}()

	// Long enough for the first heartbeat (and, if the fix weren't in place,
	// several more) to have fired.
	deadline := time.After(2 * time.Second)
	for heartbeats.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for the first heartbeat attempt")
		case <-time.After(time.Millisecond):
		}
	}
	time.Sleep(8 * keepAliveTimeout)

	if got := heartbeats.Load(); got != 1 {
		t.Fatalf("expected exactly 1 heartbeat attempt (rejected, so no retries) for the whole wait, got %d", got)
	}

	foundWarning := false
	for _, r := range logs.records() {
		if r.Level == slog.LevelWarn {
			foundWarning = true
		}
	}
	if !foundWarning {
		t.Fatal("expected a Warn log about the rejected heartbeat")
	}
}

// TestReportStatusLoop_BoundsBufferedEntriesDuringProlongedOutage pins down
// maxBufferedReportEntriesMultiplier: retrying failed reports forever (see
// TestReportStatusLoop_RetriesFailedFlushInsteadOfDroppingIt) must not let
// the buffer grow without bound if the outage never ends - old entries are
// dropped, loudly, once it's grown past the cap.
func TestReportStatusLoop_BoundsBufferedEntriesDuringProlongedOutage(t *testing.T) {
	const numTasks = 15 // > maxBufferedReportEntriesMultiplier * ReportBatchSize (10*1)

	tasks := make([]*Task, numTasks)
	for i := range tasks {
		tasks[i] = &Task{Id: fmt.Sprintf("task-%d", i), Attempts: 1}
	}
	api := &fakeMoabApi{tasks: tasks}
	logs := &recordingHandler{}

	api.reportStatusHook = func(req *ReportStatusRequest) (*ReportStatusResponse, error) {
		return nil, fmt.Errorf("simulated prolonged outage")
	}

	consumer, err := NewMoabConsumer(api, "q", slog.New(logs), ConsumerConfig{
		NumWorkers:          numTasks,
		NumPollers:          1,
		DequeueBatchSize:    int32(numTasks),
		ReportBatchSize:     1,
		ReportFlushInterval: time.Hour, // only the count-based trigger should matter here
		PollingInterval:     time.Millisecond,
		DequeueTimeout:      5 * time.Millisecond,
		ReportStatusTimeout: 5 * time.Millisecond,
		KeepAliveTimeout:    3 * time.Hour, // far longer than the test runs; heartbeating is irrelevant here
	})
	if err != nil {
		t.Fatalf("NewMoabConsumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var handled atomic.Int32
	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		consumer.Start(ctx, func(ctx context.Context, task *Task) error {
			handled.Add(1)
			return nil
		})
	}()

	deadline := time.After(2 * time.Second)
	for handled.Load() != numTasks {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for all tasks to be handled, got %d", handled.Load())
		case <-time.After(time.Millisecond):
		}
	}

	// Give the reporter time to have attempted (and failed) flushing well
	// past the cap.
	time.Sleep(100 * time.Millisecond)

	foundDropLog := false
	for _, r := range logs.records() {
		if r.Level == slog.LevelError && strings.Contains(r.Message, "dropping oldest buffered status reports") {
			foundDropLog = true
		}
	}
	if !foundDropLog {
		t.Fatal("expected a log about dropping buffered entries once the cap was exceeded")
	}

	if got := api.reportedCount(); got != 0 {
		t.Fatalf("the outage never ends in this test, so nothing should have ever been successfully recorded; got %d", got)
	}

	cancel()
	select {
	case <-startDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return within 2s of ctx cancellation")
	}
}
