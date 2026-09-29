package moab

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"runtime/debug"
	"sync"
	"time"
)

// HandlerFunc defines the handler that is run on for each task. ctx is
// Start's ctx, and is cancelled the moment shutdown begins - a handler that
// wants shutdown to stay within ShutdownTimeout should observe ctx.Done()
// and return promptly rather than running its work to completion regardless.
type HandlerFunc func(ctx context.Context, task *Task) error

// ConsumerConfig configures the tunables of a MoabConsumer. Any field left at
// its zero value is replaced with the corresponding value from
// DefaultConsumerConfig by NewMoabConsumer.
type ConsumerConfig struct {
	// NumWorkers is the number of goroutines concurrently calling the
	// consumer's HandlerFunc.
	NumWorkers int

	// NumPollers is the number of goroutines concurrently calling Dequeue.
	NumPollers int

	// BufferSize is the capacity of the channel buffering dequeued tasks
	// between pollers and workers. heartbeatLoop renews a task's lease for
	// as long as it's tracked as in-flight, whether it's sitting in this
	// buffer or already claimed by a worker, so BufferSize no longer risks
	// lease expiry by itself. It's purely a throughput/latency knob now:
	// too small and a poller's next Dequeue has to wait for a worker to
	// free up a slot; too large and tasks sit queued (and get wastefully
	// re-heartbeated in the meantime) longer than necessary before a
	// worker starts them. Keep it a small multiple of NumWorkers - enough
	// to keep workers saturated without much slack.
	// TODO: size this dynamically against actual throughput instead of a
	// fixed multiple.
	BufferSize int

	// DequeueBatchSize is the max number of tasks requested per Dequeue call.
	DequeueBatchSize int32

	// ReportBatchSize is the max number of status entries batched into a
	// single ReportStatus call. int32, matching DequeueBatchSize, even
	// though it's never itself put on the wire - kept consistent so the two
	// batch-size knobs don't silently drift into different types.
	ReportBatchSize int32

	// ReportFlushInterval is the max time completed task statuses sit
	// buffered before being flushed via ReportStatus, regardless of whether
	// ReportBatchSize has been reached. Must be under
	// KeepAliveTimeout/heartbeatDivisor (Validate enforces it): a
	// SUCCEEDED/FAILED report stuck in the batch for too long risks
	// arriving after the lease it's closing out has already been reclaimed.
	ReportFlushInterval time.Duration

	// PollingInterval is how long a poller waits before retrying after an
	// empty Dequeue response or a Dequeue error.
	PollingInterval time.Duration

	// DequeueTimeout bounds each individual Dequeue RPC call. Must be under
	// KeepAliveTimeout (Validate enforces it) - a Dequeue call that itself
	// takes a meaningful fraction of the lease leaves that much less of it
	// for the handler.
	DequeueTimeout time.Duration

	// ReportStatusTimeout bounds each individual ReportStatus RPC call -
	// every heartbeat and every final SUCCEEDED/FAILED report. Must be
	// under KeepAliveTimeout/heartbeatDivisor (Validate enforces it): a
	// heartbeat that itself takes too long to complete can lose its race
	// against the lease it was sent to renew.
	ReportStatusTimeout time.Duration

	// KeepAliveTimeout is sent as the keepalive override (in seconds) on
	// every Dequeue and ReportStatus call this consumer makes - it, not the
	// server, decides how long it needs to hold a task's lease. Required;
	// NewMoabConsumer rejects zero. On the wire, 0 legitimately means
	// "inherit the queue's default" (see DequeueRequest's field of the same
	// name), but that's a choice for a caller that genuinely doesn't care;
	// this consumer does, since the value also drives its own behavior -
	// heartbeatLoop's cadence, the stale-report Warn, and Validate's checks
	// on ReportFlushInterval/ReportStatusTimeout/DequeueTimeout below - and
	// "0, so skip all of that" is a footgun this type would rather not
	// offer.
	KeepAliveTimeout time.Duration

	// ShutdownTimeout bounds how long Start waits, once its ctx is
	// cancelled, for in-flight HandlerFunc calls to finish before giving up
	// on them. A handler still running when it elapses is abandoned - Start
	// returns anyway, after flushing whatever statuses have already come
	// in. That handler's eventual result (if it ever returns) is never
	// reported; the task is left for the queue's own keepalive-timeout
	// reclaim to redeliver.
	ShutdownTimeout time.Duration
}

// DefaultConsumerConfig returns the tunables MoabConsumer used before they
// became configurable.
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		NumWorkers:          32,
		NumPollers:          1,
		BufferSize:          32 * 16,
		DequeueBatchSize:    10,
		ReportBatchSize:     10,
		ReportFlushInterval: 5 * time.Second,
		PollingInterval:     time.Second,
		DequeueTimeout:      5 * time.Second,
		ReportStatusTimeout: 5 * time.Second,
		ShutdownTimeout:     30 * time.Second,
	}
}

// withDefaults returns a copy of c with every zero-valued field replaced by
// its DefaultConsumerConfig value. KeepAliveTimeout is intentionally left
// alone: it has no default, since Validate requires the caller to set it
// explicitly rather than silently falling back to something.
func (c ConsumerConfig) withDefaults() ConsumerConfig {
	d := DefaultConsumerConfig()

	if c.NumWorkers <= 0 {
		c.NumWorkers = d.NumWorkers
	}
	if c.NumPollers <= 0 {
		c.NumPollers = d.NumPollers
	}
	if c.BufferSize <= 0 {
		c.BufferSize = d.BufferSize
	}
	if c.DequeueBatchSize <= 0 {
		c.DequeueBatchSize = d.DequeueBatchSize
	}
	if c.ReportBatchSize <= 0 {
		c.ReportBatchSize = d.ReportBatchSize
	}
	if c.ReportFlushInterval <= 0 {
		c.ReportFlushInterval = d.ReportFlushInterval
	}
	if c.PollingInterval <= 0 {
		c.PollingInterval = d.PollingInterval
	}
	if c.DequeueTimeout <= 0 {
		c.DequeueTimeout = d.DequeueTimeout
	}
	if c.ReportStatusTimeout <= 0 {
		c.ReportStatusTimeout = d.ReportStatusTimeout
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = d.ShutdownTimeout
	}

	return c
}

const (
	// heartbeatDivisor sets how often an in-flight task is heartbeated (every
	// KeepAliveTimeout/heartbeatDivisor) and how large ReportStatusTimeout and
	// ReportFlushInterval are allowed to be relative to KeepAliveTimeout
	// (ConsumerConfig.Validate rejects either one at or above that same
	// fraction). All three delay when a status report - a heartbeat, or the
	// final SUCCEEDED/FAILED - actually reaches the server; keeping all of them
	// comfortably inside half the lease is what keeps a report from landing
	// after the server has already reclaimed the task as abandoned.
	heartbeatDivisor = 2

	// heartbeatSweepDivisor sets how often heartbeatLoop checks for in-flight
	// tasks due for a heartbeat, relative to the heartbeat interval itself
	// (every heartbeatInterval/heartbeatSweepDivisor). Finer than the interval
	// it's checking against, so a task becoming due right after one sweep is
	// still caught well before the next full interval elapses, rather than
	// sitting overdue for up to an entire heartbeatInterval.
	heartbeatSweepDivisor = 4

	// maxBufferedReportEntriesMultiplier bounds how many completed-task status
	// reports reportStatusLoop will hold onto (as a multiple of ReportBatchSize)
	// while retrying a ReportStatus call that keeps failing at the RPC level.
	// Retrying is safe and worth doing - the server's attempt-fencing makes a
	// retried report idempotent - but unbounded retries would let a prolonged
	// outage grow this buffer forever, so once it's exceeded the oldest entries
	// are dropped (loudly) to make room, rather than either silently dropping
	// everything or growing without limit.
	maxBufferedReportEntriesMultiplier = 10
)

// heartbeatInterval is how often an in-flight task is heartbeated:
// comfortably before KeepAliveTimeout, so the renewal lands well ahead of
// the deadline it's extending.
func (c ConsumerConfig) heartbeatInterval() time.Duration {
	return c.KeepAliveTimeout / heartbeatDivisor
}

// Validate reports whether c makes sense, once withDefaults has already
// filled in every unset field. KeepAliveTimeout must be positive - see its
// doc comment for why this type doesn't offer a "skip all of this" zero
// value - and every other check here exists to keep some other duration
// from eating too far into the lease it defines.
func (c ConsumerConfig) Validate() error {
	if c.KeepAliveTimeout <= 0 {
		return fmt.Errorf("KeepAliveTimeout must be greater than 0")
	}

	if c.ReportFlushInterval >= c.heartbeatInterval() {
		return fmt.Errorf("ReportFlushInterval (%s) must be less than KeepAliveTimeout/%d (%s): a completed task's status report must not be able to sit buffered long enough to miss its own lease",
			c.ReportFlushInterval, heartbeatDivisor, c.heartbeatInterval())
	}

	if c.ReportStatusTimeout >= c.heartbeatInterval() {
		return fmt.Errorf("ReportStatusTimeout (%s) must be less than KeepAliveTimeout/%d (%s): a heartbeat or final report that itself takes this long can lose its race against the lease it's renewing or closing out",
			c.ReportStatusTimeout, heartbeatDivisor, c.heartbeatInterval())
	}

	if c.DequeueTimeout >= c.KeepAliveTimeout {
		return fmt.Errorf("DequeueTimeout (%s) must be less than KeepAliveTimeout (%s): a Dequeue call that itself takes this long leaves that much less of the lease for the handler",
			c.DequeueTimeout, c.KeepAliveTimeout)
	}

	return nil
}

// leaseInfo tracks one in-flight task's lease for heartbeatLoop and the
// stale-report check in reportStatusLoop: the Attempt a heartbeat must
// carry (matching what Dequeue handed out), and when the lease was last
// renewed - by the original Dequeue, or by whichever heartbeat most
// recently succeeded.
type leaseInfo struct {
	attempt     int32
	lastRenewal time.Time
}

type taskCompletionStatus struct {
	taskId  string
	attempt int32
	err     error
}

// MoabConsumer pulls tasks off one queue and drives them through a
// HandlerFunc. NumPollers goroutines call Dequeue and hand tasks to a
// bounded buffer; NumWorkers goroutines drain that buffer and call the
// handler, reporting SUCCEEDED/FAILED via batched ReportStatus calls once
// each returns. Meanwhile a separate heartbeatLoop periodically renews the
// lease (via ReportStatus(IN_PROGRESS)) of every task this consumer
// currently holds - whether still queued in the buffer or already being
// handled - so long-running work never has its task reclaimed out from
// under it; Validate checks DequeueTimeout, ReportStatusTimeout, and
// ReportFlushInterval against KeepAliveTimeout, so none of them can eat too
// far into the lease they share.
type MoabConsumer struct {
	moabClient MoabApi
	queueName  string
	logger     *slog.Logger
	config     ConsumerConfig

	// leases tracks every task's lease from the moment Dequeue hands it out
	// until its final status is reported, keyed by task id - regardless of
	// whether it's still sitting in bufCh or already claimed by a worker.
	leases   map[string]*leaseInfo
	mu       sync.Mutex
	bufCh    chan *Task
	statusCh chan taskCompletionStatus
}

// Start runs the consumer until ctx is cancelled, then returns once shutdown
// finishes (or ShutdownTimeout elapses). It resets all leases/bufCh/statusCh
// state at the top, so calling it again after a previous call has returned -
// e.g. from a supervising retry loop - starts clean rather than resurfacing
// whatever that previous run left behind. That reset assumes the previous
// call has actually returned first: it must not run concurrently with, or
// while stragglers from, a still-active or ShutdownTimeout-abandoned prior
// call.
func (c *MoabConsumer) Start(ctx context.Context, h HandlerFunc) {
	c.mu.Lock()
	c.leases = make(map[string]*leaseInfo)
	c.mu.Unlock()
	c.bufCh = make(chan *Task, c.config.BufferSize)
	// Buffered to one slot per worker so a worker's send can never block
	// forever, even if it outlives ShutdownTimeout and the status reporter
	// has already stopped reading from it.
	c.statusCh = make(chan taskCompletionStatus, c.config.NumWorkers)

	var pollersWg, workersWg sync.WaitGroup

	for i := 0; i < c.config.NumPollers; i++ {
		pollersWg.Go(func() {
			c.pollLoop(ctx)
		})
	}

	for i := 0; i < c.config.NumWorkers; i++ {
		workersWg.Go(func() {
			c.workerLoop(ctx, h)
		})
	}

	drained := make(chan struct{})

	reporterDone := make(chan struct{})
	go func() {
		defer close(reporterDone)
		c.reportStatusLoop(drained)
	}()

	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		c.heartbeatLoop(drained)
	}()

	allStopped := make(chan struct{})
	go func() {
		pollersWg.Wait()
		workersWg.Wait()
		close(allStopped)
	}()

	// Wait for ctx to actually be cancelled before racing ShutdownTimeout
	// below: without this, the timer would start from the moment Start was
	// called rather than from shutdown actually beginning, so any run that
	// simply lasts longer than ShutdownTimeout would trip the "abandon" path
	// below on a perfectly healthy consumer - closing drained while
	// pollLoop/workerLoop keep right on running. statusCh is only sized to
	// absorb one still-in-flight send per worker past that point (see
	// below); many more than that per worker, from a consumer that never
	// actually stopped, would fill it and deadlock every worker forever.
	<-ctx.Done()

	// Pollers stop taking new work as soon as ctx is cancelled. Workers keep
	// running until ctx is cancelled AND they finish whatever handler call
	// is already in flight, unconditionally reporting its status before
	// exiting - so waiting for both here guarantees no worker can possibly
	// send on statusCh again. ShutdownTimeout bounds that wait: a handler
	// that ignores ctx cancellation and keeps running past it is abandoned
	// rather than left to hang Start forever (statusCh is sized so that
	// worker's eventual send, if it ever returns, never blocks it either
	// way). Only once we've stopped waiting is it safe to tell the status
	// reporter to flush whatever's left in its current sub-batch, and the
	// heartbeat loop to stop renewing anything still in flight, and stop -
	// an abandoned handler's task is meant to lapse and be reclaimed by the
	// server, not have its lease kept alive forever by a heartbeat that
	// outlived the point Start gave up on it.
	select {
	case <-allStopped:
	case <-time.After(c.config.ShutdownTimeout):
		c.logger.Warn("shutdown timeout exceeded, abandoning still-running handlers", "queue", c.queueName, "timeout", c.config.ShutdownTimeout)
	}

	close(drained)
	<-reporterDone
	<-heartbeatDone
}

// reportStatusResult is the outcome of one ReportStatus RPC attempt made on
// batch, delivered back to reportStatusLoop's goroutine over a channel
// rather than returned directly - see sendReportStatusBatch's doc comment
// for why the call itself must never run on that goroutine.
type reportStatusResult struct {
	batch []*ReportStatusRequestEntry
	resp  *ReportStatusResponse
	err   error
}

// sendReportStatusBatch performs one ReportStatus RPC for batch. Always
// called from a dedicated goroutine, never from reportStatusLoop's own
// goroutine directly: that goroutine is also the only reader of
// c.statusCh, and workers block sending on c.statusCh once it fills (sized
// to only one pending completion per worker) - so a slow or failing
// ReportStatus call sitting inline on that goroutine would stall every
// worker, and backpressure from a full bufCh would in turn stall every
// poller, for as long as the call takes to time out.
//
// Uses a context detached from Start's ctx (rather than a child of it) so a
// flush triggered by ctx being cancelled - the shutdown flush in
// reportStatusLoop - still gets a chance to go out instead of being
// cancelled before it starts.
func (c *MoabConsumer) sendReportStatusBatch(batch []*ReportStatusRequestEntry) reportStatusResult {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.ReportStatusTimeout)
	defer cancel()

	resp, err := c.moabClient.ReportStatus(ctx, &ReportStatusRequest{
		QueueName: c.queueName,
		Entries:   batch,
	})
	return reportStatusResult{batch: batch, resp: resp, err: err}
}

func (c *MoabConsumer) reportStatusLoop(drained <-chan struct{}) {
	entries := make([]*ReportStatusRequestEntry, 0, c.config.ReportBatchSize)
	sending := false
	resultCh := make(chan reportStatusResult)

	// send hands whatever's pending off to sendReportStatusBatch on a
	// background goroutine and returns immediately, so this loop keeps
	// draining c.statusCh (and heartbeatLoop keeps renewing leases
	// regardless) for as long as that RPC call is in flight. At most one
	// send is ever outstanding at a time - a second call while sending is
	// true is a no-op, since entries has nothing new to hand off until the
	// first one's result comes back through resultCh.
	send := func() {
		if sending || len(entries) == 0 {
			return
		}
		batch := entries
		entries = make([]*ReportStatusRequestEntry, 0, c.config.ReportBatchSize)
		sending = true
		go func() {
			resultCh <- c.sendReportStatusBatch(batch)
		}()
	}

	// mergeResult folds one RPC attempt's outcome back into entries, without
	// deciding whether to retry - callers that want the normal retry cadence
	// (every new arrival or tick, indefinitely) do that themselves; the
	// shutdown path deliberately doesn't, see reportStatusLoop's drained
	// case below.
	mergeResult := func(res reportStatusResult) {
		if res.err != nil {
			// Prepended, not appended: these are older than whatever's
			// accumulated in entries since this batch was taken, and
			// maxBufferedReportEntriesMultiplier's trim below assumes the
			// oldest entries sit at the front.
			c.logger.Error("failed to report task status, will retry", "queue", c.queueName, "error", res.err, "count", len(res.batch))
			entries = append(res.batch, entries...)

			if maxBuffered := maxBufferedReportEntriesMultiplier * int(c.config.ReportBatchSize); len(entries) > maxBuffered {
				dropped := len(entries) - maxBuffered
				c.logger.Error("dropping oldest buffered status reports to bound memory during a prolonged outage",
					"queue", c.queueName, "dropped", dropped)
				entries = entries[dropped:]
			}
			return
		}

		// The call succeeded, but that only means the server processed the
		// batch - each entry's own Result says whether its report actually
		// took effect. A non-OK one isn't something retrying would fix (the
		// task's already gone, or this attempt is already stale), so it's
		// still safe to drop; just make it visible.
		for _, e := range res.resp.Entries {
			if e.Result != ReportStatusResponseEntry_RESULT_OK {
				c.logger.Warn("status report was not applied", "queue", c.queueName, "task_id", e.TaskId, "result", e.Result)
			}
		}
	}

	// applyResult is mergeResult plus the normal-operation retry policy:
	// whatever's left in entries - freshly merged back in on failure, or
	// simply accumulated while this batch was in flight - gets sent again
	// immediately once it reaches ReportBatchSize, rather than waiting for
	// the next tick.
	applyResult := func(res reportStatusResult) {
		sending = false
		mergeResult(res)
		if len(entries) >= int(c.config.ReportBatchSize) {
			send()
		}
	}

	handle := func(status taskCompletionStatus) {
		var reportedStatus ReportStatusRequestEntry_Status
		if status.err != nil {
			reportedStatus = ReportStatusRequestEntry_STATUS_FAILED
		} else {
			reportedStatus = ReportStatusRequestEntry_STATUS_SUCCEEDED
		}
		entries = append(entries, &ReportStatusRequestEntry{
			TaskId:                    status.taskId,
			Attempt:                   status.attempt,
			Status:                    reportedStatus,
			KeepaliveTimeoutInSeconds: int64(c.config.KeepAliveTimeout / time.Second),
		})

		c.mu.Lock()
		lease, ok := c.leases[status.taskId]
		delete(c.leases, status.taskId)
		c.mu.Unlock()

		if ok {
			if elapsed := time.Since(lease.lastRenewal); elapsed > c.config.KeepAliveTimeout {
				c.logger.Warn("task handling exceeded keepalive timeout; status report may be rejected as stale",
					"queue", c.queueName, "task_id", status.taskId, "elapsed", elapsed, "keep_alive_timeout", c.config.KeepAliveTimeout)
			}
		}

		if len(entries) >= int(c.config.ReportBatchSize) {
			send()
		}
	}

	ticker := time.NewTicker(c.config.ReportFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-drained:
			// statusCh is buffered, so a worker's send can have already
			// completed - and be sitting unread - by the time drained
			// fires; drain it before the final flush so nothing already
			// sent gets lost.
			draining := true
			for draining {
				select {
				case status := <-c.statusCh:
					handle(status)
				default:
					draining = false
				}
			}

			c.logger.Debug("statusReporter: stopping, all workers drained", "queue", c.queueName)

			// Let any send already in flight land - merging its result
			// (including anything it puts back on failure) into entries -
			// before deciding what's left to flush. Not looped through
			// applyResult's own retry: at shutdown we get exactly one more
			// attempt at whatever remains, not an indefinite retry loop
			// against a server that may be the reason we're stuck.
			if sending {
				mergeResult(<-resultCh)
				sending = false
			}

			if len(entries) > 0 {
				res := c.sendReportStatusBatch(entries)
				if res.err != nil {
					c.logger.Error("failed to report task status during shutdown, giving up", "queue", c.queueName, "error", res.err, "count", len(entries))
				} else {
					for _, e := range res.resp.Entries {
						if e.Result != ReportStatusResponseEntry_RESULT_OK {
							c.logger.Warn("status report was not applied", "queue", c.queueName, "task_id", e.TaskId, "result", e.Result)
						}
					}
				}
			}
			return
		case <-ticker.C:
			send()
		case status := <-c.statusCh:
			handle(status)
		case res := <-resultCh:
			applyResult(res)
		}
	}
}

func (c *MoabConsumer) workerLoop(ctx context.Context, h HandlerFunc) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Debug("worker: stopping, context cancelled", "queue", c.queueName)
			return
		case task := <-c.bufCh:
			err := c.runHandler(ctx, h, task)

			// Safe to send unconditionally and without a ctx.Done() escape:
			// statusCh is buffered to hold one pending send per worker, so
			// this never blocks even if it outlives ShutdownTimeout and the
			// status reporter has already stopped reading.
			c.statusCh <- taskCompletionStatus{
				taskId:  task.Id,
				attempt: task.Attempts,
				err:     err,
			}
		}
	}
}

// heartbeatLoop periodically renews the lease of every task this consumer
// currently has in flight - whether it's still sitting in bufCh waiting for
// a worker, or already claimed by one - for as long as drained hasn't
// fired. It runs independently of any single worker: a task's lease needs
// renewing from the moment Dequeue hands it out, before any worker has even
// seen it, not just while a handler is actively running.
func (c *MoabConsumer) heartbeatLoop(drained <-chan struct{}) {
	interval := c.config.heartbeatInterval()

	// Validate requires KeepAliveTimeout > 0, but guards against the
	// integer division above ever landing on a non-positive tick: NewTicker
	// panics on one, and there's no sane duration this could fall back to
	// other than interval itself.
	sweepInterval := interval / heartbeatSweepDivisor
	if sweepInterval <= 0 {
		sweepInterval = interval
	}

	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-drained:
			return
		case <-ticker.C:
			c.sweepHeartbeats(interval)
		}
	}
}

// sweepHeartbeats renews the lease of every task in c.leases last renewed
// at least interval ago, via one or more batched ReportStatus(IN_PROGRESS)
// calls (chunked to ReportBatchSize entries each, the same cap
// reportStatusLoop uses for its own batches).
func (c *MoabConsumer) sweepHeartbeats(interval time.Duration) {
	now := time.Now()

	c.mu.Lock()
	due := make([]string, 0)
	for taskId, lease := range c.leases {
		if now.Sub(lease.lastRenewal) >= interval {
			due = append(due, taskId)
		}
	}
	c.mu.Unlock()

	for len(due) > 0 {
		n := min(len(due), int(c.config.ReportBatchSize))
		c.sendHeartbeatBatch(due[:n])
		due = due[n:]
	}
}

// sendHeartbeatBatch sends one ReportStatus call renewing every task id in
// taskIds, then - only for those that succeeded and are still tracked -
// advances its lease's lastRenewal to roughly now. Uses a context detached
// from Start's ctx (rather than a child of it), same as reportStatusLoop's
// flush: a heartbeat's entire purpose is to keep a lease alive for a task
// that's still legitimately in flight, so it must not be cut short by the
// same cancellation a slow handler elsewhere may be ignoring.
func (c *MoabConsumer) sendHeartbeatBatch(taskIds []string) {
	c.mu.Lock()
	entries := make([]*ReportStatusRequestEntry, 0, len(taskIds))
	for _, taskId := range taskIds {
		lease, ok := c.leases[taskId]
		if !ok {
			// Completed and removed between the scan and now.
			continue
		}
		entries = append(entries, &ReportStatusRequestEntry{
			TaskId:                    taskId,
			Attempt:                   lease.attempt,
			Status:                    ReportStatusRequestEntry_STATUS_IN_PROGRESS,
			KeepaliveTimeoutInSeconds: int64(c.config.KeepAliveTimeout / time.Second),
		})
	}
	c.mu.Unlock()

	if len(entries) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.ReportStatusTimeout)
	defer cancel()

	resp, err := c.moabClient.ReportStatus(ctx, &ReportStatusRequest{
		QueueName: c.queueName,
		Entries:   entries,
	})
	if err != nil {
		// Left overdue - the next sweep will retry them.
		c.logger.Error("failed to send keepalive heartbeat batch", "queue", c.queueName, "count", len(entries), "error", err)
		return
	}

	c.logger.Debug("sent keepalive heartbeat batch", "queue", c.queueName, "count", len(entries))

	now := time.Now()
	c.mu.Lock()
	for _, e := range resp.Entries {
		if e.Result == ReportStatusResponseEntry_RESULT_OK {
			if lease, ok := c.leases[e.TaskId]; ok {
				lease.lastRenewal = now
			}
			continue
		}

		// This consumer no longer holds a valid lease on this task - it's
		// already gone, or the attempt it heartbeated under is stale, so
		// the task has already been reclaimed (and possibly redelivered)
		// elsewhere. Stop heartbeating it: retrying would just get rejected
		// again, and a handler still running locally for it has already
		// lost the race - its eventual final report will be rejected too.
		delete(c.leases, e.TaskId)
		c.logger.Warn("heartbeat rejected, lease no longer held by this consumer",
			"queue", c.queueName, "task_id", e.TaskId, "result", e.Result)
	}
	c.mu.Unlock()
}

func (c *MoabConsumer) pollLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Debug("consumer: stopping, context cancelled", "queue", c.queueName)
			return
		default:
			c.logger.Debug("consumer: polling", "queue", c.queueName)

			ctx2, cancel := context.WithTimeout(ctx, c.config.DequeueTimeout)
			resp, err := c.moabClient.Dequeue(ctx2, &DequeueRequest{
				BatchSize:                 c.config.DequeueBatchSize,
				QueueName:                 c.queueName,
				KeepaliveTimeoutInSeconds: int64(c.config.KeepAliveTimeout / time.Second),
			})
			cancel()

			if err != nil {
				c.logger.Error("failed to dequeue tasks", "queue", c.queueName, "error", err)
				c.sleep(ctx, c.pollingIntervalWithJitter())
				continue
			}

			if len(resp.Tasks) > 0 {
				now := time.Now()

				c.mu.Lock()
				for i := range resp.Tasks {
					c.leases[resp.Tasks[i].Id] = &leaseInfo{attempt: resp.Tasks[i].Attempts, lastRenewal: now}
				}
				c.mu.Unlock()

				for i := range resp.Tasks {
					// Workers may have already exited on ctx.Done(), in which
					// case nothing will ever drain bufCh; an unconditional
					// send here would block this loop forever.
					select {
					case c.bufCh <- resp.Tasks[i]:
					case <-ctx.Done():
						return
					}
				}
			} else {
				c.logger.Debug("consumer: empty dequeue response, sleeping", "queue", c.queueName)
				c.sleep(ctx, c.pollingIntervalWithJitter())
				// TODO emit metric for empty response
			}
		}
	}
}

func (c *MoabConsumer) sleep(ctx context.Context, d time.Duration) {
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}

// pollingIntervalWithJitter adds up to 50% random jitter on top of
// PollingInterval. Without it, every poller backs off by the exact same
// fixed duration after an error or an empty Dequeue response, so - with
// NumPollers > 1, or many consumer instances sharing a queue - they'd all
// retry in lockstep instead of staggered.
func (c *MoabConsumer) pollingIntervalWithJitter() time.Duration {
	d := c.config.PollingInterval
	return d + rand.N(d/2+1)
}

// runHandler runs h, recovering a panic and turning it into an error so a
// single misbehaving task fails (and gets reported/retried) instead of
// taking down the whole consumer process.
func (c *MoabConsumer) runHandler(ctx context.Context, h HandlerFunc, task *Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("panic in task handler", "queue", c.queueName, "task_id", task.Id, "panic", r, "stack", string(debug.Stack()))
			err = fmt.Errorf("panic in task handler: %v", r)
		}
	}()

	return h(ctx, task)
}

func NewMoabConsumer(moabClient MoabApi, queueName string, logger *slog.Logger, config ConsumerConfig) (*MoabConsumer, error) {
	if logger == nil {
		logger = slog.Default()
	}
	config = config.withDefaults()

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ConsumerConfig: %w", err)
	}

	// leases/bufCh/statusCh are left nil here - Start creates them fresh on
	// every call (see its doc comment), so initializing them here too would
	// just be redundant.
	return &MoabConsumer{
		moabClient: moabClient,
		queueName:  queueName,
		logger:     logger,
		config:     config,
	}, nil
}
