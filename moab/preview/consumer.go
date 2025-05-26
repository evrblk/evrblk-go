package moab

import (
	"context"
	"sync"
	"time"
)

// HandlerFunc is used to define the Handler that is run on for each task
type HandlerFuncPreview func(task *Task) error

// HandleTask wraps a function for handling sqs messages
func (f HandlerFuncPreview) HandleTask(task *Task) error {
	return f(task)
}

// Handler interface
type HandlerPreview interface {
	HandleTask(task *Task) error
}

type taskCompletionStatus struct {
	taskId string
	err    error
}

type MoabPreviewConsumer struct {
	moabClient MoabPreviewApi
	queueName  string

	inflightTasks map[string]*Task
	mu            sync.Mutex
	bufCh         chan *Task
	statusCh      chan taskCompletionStatus
	numWorkers    int
}

func (c *MoabPreviewConsumer) Start(ctx context.Context, h HandlerPreview) {
	go func(ctx context.Context) {
		entries := make([]*ReportStatusRequestEntry, 0)
		for {
			select {
			case <-ctx.Done():
				//log.Println("statusReporter: Stopping polling because a context kill signal was sent")
				return
			case status := <-c.statusCh:
				var reportedStatus ReportStatusRequestEntry_Status
				if status.err != nil {
					reportedStatus = ReportStatusRequestEntry_STATUS_FAILED
				} else {
					reportedStatus = ReportStatusRequestEntry_STATUS_SUCCEEDED
				}
				entries = append(entries, &ReportStatusRequestEntry{
					TaskId: status.taskId,
					Status: reportedStatus,
				})

				c.mu.Lock()
				delete(c.inflightTasks, status.taskId)
				c.mu.Unlock()

				if len(entries) >= 10 {
					req := &ReportStatusRequest{
						QueueName: c.queueName,
						Entries:   entries,
					}

					ctx2, cancel := context.WithTimeout(ctx, time.Millisecond*time.Duration(5000))
					defer cancel()

					_, err := c.moabClient.ReportStatus(ctx2, req)
					if err != nil {
						//log.Println(err)
					}

					entries = make([]*ReportStatusRequestEntry, 0)
				}
			}
		}
	}(ctx)

	for i := 0; i < c.numWorkers; i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					//log.Println("worker: Stopping polling because a context kill signal was sent")
					return
				case task := <-c.bufCh:
					err := h.HandleTask(task)
					// TODO catch panics

					c.statusCh <- taskCompletionStatus{
						taskId: task.Id,
						err:    err,
					}
				}
			}
		}(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			//log.Println("consumer: Stopping polling because a context kill signal was sent")
			return
		default:
			//log.Println("consumer: Polling")

			ctx2, cancel := context.WithTimeout(ctx, time.Millisecond*time.Duration(5000))
			defer cancel()

			resp, err := c.moabClient.Dequeue(ctx2, &DequeueRequest{
				BatchSize: 10,
				QueueName: c.queueName,
			})

			if err != nil {
				//log.Println(err)
				time.Sleep(time.Second) // TODO sleep
				continue
			}

			if len(resp.Tasks) > 0 {
				c.mu.Lock()
				for i := range resp.Tasks {
					c.inflightTasks[resp.Tasks[i].Id] = resp.Tasks[i]
				}
				c.mu.Unlock()
				for i := range resp.Tasks {
					c.bufCh <- resp.Tasks[i]
				}
			} else {
				//log.Println("Empty. Sleeping...")
				time.Sleep(time.Second) // TODO configure sleep
				// TODO emit metric for empty response
			}
		}
	}
}

func NewMoabPreviewConsumer(moabClient MoabPreviewApi, queueName string) *MoabPreviewConsumer {
	return &MoabPreviewConsumer{
		moabClient:    moabClient,
		queueName:     queueName,
		inflightTasks: make(map[string]*Task),
		bufCh:         make(chan *Task, 32*16),
		statusCh:      make(chan taskCompletionStatus),
		numWorkers:    32,
	}
}
