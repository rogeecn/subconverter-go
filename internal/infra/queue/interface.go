package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/rogeecn/subconverter-go/internal/app/converter"
	"github.com/rogeecn/subconverter-go/internal/pkg/logger"
)

// Job represents a conversion job
type Job struct {
	ID        string                     `json:"id"`
	Type      string                     `json:"type"`
	Request   converter.ConvertRequest   `json:"request"`
	CreatedAt time.Time                  `json:"created_at"`
	Status    string                     `json:"status"`
	Result    *converter.ConvertResponse `json:"result,omitempty"`
	Error     string                     `json:"error,omitempty"`
}

// Queue defines the interface for job queue operations
type Queue interface {
	Push(ctx context.Context, job *Job) error
	Pop(ctx context.Context) (*Job, error)
	Complete(ctx context.Context, jobID string, result *converter.ConvertResponse) error
	Fail(ctx context.Context, jobID string, err error) error
	Get(ctx context.Context, jobID string) (*Job, error)
}

// MemoryQueue implements in-memory job queue
type MemoryQueue struct {
	queue chan *Job
	jobs  map[string]*Job
}

// NewMemoryQueue creates a new in-memory queue
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		queue: make(chan *Job, 1000),
		jobs:  make(map[string]*Job),
	}
}

func (q *MemoryQueue) Push(ctx context.Context, job *Job) error {
	job.ID = generateJobID()
	job.CreatedAt = time.Now()
	job.Status = "pending"

	q.jobs[job.ID] = job

	select {
	case q.queue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *MemoryQueue) Pop(ctx context.Context) (*Job, error) {
	select {
	case job := <-q.queue:
		job.Status = "processing"
		return job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (q *MemoryQueue) Complete(ctx context.Context, jobID string, result *converter.ConvertResponse) error {
	if job, exists := q.jobs[jobID]; exists {
		job.Status = "completed"
		job.Result = result
	}
	return nil
}

func (q *MemoryQueue) Fail(ctx context.Context, jobID string, err error) error {
	if job, exists := q.jobs[jobID]; exists {
		job.Status = "failed"
		job.Error = err.Error()
	}
	return nil
}

func (q *MemoryQueue) Get(ctx context.Context, jobID string) (*Job, error) {
	if job, exists := q.jobs[jobID]; exists {
		return job, nil
	}
	return nil, nil
}

// Worker processes jobs from the queue
type Worker struct {
	queue   Queue
	service *converter.Service
	log     logger.Logger
}

// NewWorker creates a new worker
func NewWorker(queue Queue, service *converter.Service, log logger.Logger) *Worker {
	return &Worker{
		queue:   queue,
		service: service,
		log:     log,
	}
}

// Start starts the worker
func (w *Worker) Start(ctx context.Context, numWorkers int) error {
	w.log.WithField("workers", numWorkers).Info("Starting worker pool")

	for i := 0; i < numWorkers; i++ {
		go w.worker(ctx, i)
	}

	<-ctx.Done()
	w.log.Info("Worker pool shutting down")

	return nil
}

func (w *Worker) worker(ctx context.Context, id int) {
	w.log.WithField("worker_id", id).Info("Worker started")

	for {
		select {
		case <-ctx.Done():
			w.log.WithField("worker_id", id).Info("Worker stopped")
			return
		default:
			job, err := w.queue.Pop(ctx)
			if err != nil {
				if err != context.Canceled {
					w.log.WithError(err).Error("Failed to get job from queue")
				}
				continue
			}

			if job == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			w.processJob(ctx, job)
		}
	}
}

func (w *Worker) processJob(ctx context.Context, job *Job) {
	w.log.WithFields(map[string]interface{}{
		"job_id": job.ID,
		"type":   job.Type,
	}).Info("Processing job")

	result, err := w.service.Convert(ctx, &job.Request)
	if err != nil {
		w.log.WithError(err).Error("Job failed")
		w.queue.Fail(ctx, job.ID, err)
		return
	}

	if err := w.queue.Complete(ctx, job.ID, result); err != nil {
		w.log.WithError(err).Error("Failed to complete job")
		return
	}

	w.log.WithFields(map[string]interface{}{
		"job_id":  job.ID,
		"proxies": len(result.Proxies),
	}).Info("Job completed")
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
