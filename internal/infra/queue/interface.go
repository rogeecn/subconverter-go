package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/subconverter/subconverter-go/internal/app/converter"
	"github.com/subconverter/subconverter-go/internal/infra/config"
	"github.com/subconverter/subconverter-go/internal/pkg/logger"
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
	queue    chan *Job
	jobs     map[string]*Job
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

// RedisQueue implements Redis-based job queue
type RedisQueue struct {
	client *redis.Client
	prefix string
}

// NewRedisQueue creates a new Redis queue
func NewRedisQueue(cfg config.RedisConfig) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.Database,
	})
	
	return &RedisQueue{
		client: client,
		prefix: "subconverter:jobs:",
	}, nil
}

func (q *RedisQueue) Push(ctx context.Context, job *Job) error {
	if job.ID == "" {
		job.ID = generateJobID()
	}
	job.CreatedAt = time.Now()
	job.Status = "pending"
	
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	
	pipe := q.client.Pipeline()
	pipe.Set(ctx, q.prefix+job.ID, data, 24*time.Hour)
	pipe.LPush(ctx, q.prefix+"queue", job.ID)
	_, err = pipe.Exec(ctx)
	
	return err
}

func (q *RedisQueue) Pop(ctx context.Context) (*Job, error) {
	result, err := q.client.BRPop(ctx, 0, q.prefix+"queue").Result()
	if err != nil {
		return nil, err
	}
	
	if len(result) < 2 {
		return nil, nil
	}
	
	jobID := result[1]
	data, err := q.client.Get(ctx, q.prefix+jobID).Bytes()
	if err != nil {
		return nil, err
	}
	
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	
	job.Status = "processing"
	data, _ = json.Marshal(job)
	q.client.Set(ctx, q.prefix+jobID, data, 24*time.Hour)
	
	return &job, nil
}

func (q *RedisQueue) Complete(ctx context.Context, jobID string, result *converter.ConvertResponse) error {
	data, err := q.client.Get(ctx, q.prefix+jobID).Bytes()
	if err != nil {
		return err
	}
	
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return err
	}
	
	job.Status = "completed"
	job.Result = result
	data, _ = json.Marshal(job)
	
	return q.client.Set(ctx, q.prefix+jobID, data, 24*time.Hour).Err()
}

func (q *RedisQueue) Fail(ctx context.Context, jobID string, err error) error {
	data, err := q.client.Get(ctx, q.prefix+jobID).Bytes()
	if err != nil {
		return err
	}
	
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return err
	}
	
	job.Status = "failed"
	job.Error = err.Error()
	data, _ = json.Marshal(job)
	
	return q.client.Set(ctx, q.prefix+jobID, data, 24*time.Hour).Err()
}

func (q *RedisQueue) Get(ctx context.Context, jobID string) (*Job, error) {
	data, err := q.client.Get(ctx, q.prefix+jobID).Bytes()
	if err != nil {
		return nil, err
	}
	
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	
	return &job, nil
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
		"job_id": job.ID,
		"proxies": len(result.Proxies),
	}).Info("Job completed")
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}