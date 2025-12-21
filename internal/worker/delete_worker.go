package worker

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// DeleteRequest describes a user's URLs to delete.
type DeleteRequest struct {
	UserID string
	URLIDs []string
}

// DeleteService deletes user URLs in the underlying storage.
type DeleteService interface {
	DeleteUserURLs(userID string, urlIDs []string) error
}

// DeleteWorkerPool batches and processes asynchronous delete requests.
type DeleteWorkerPool struct {
	service      DeleteService
	requestChan  chan DeleteRequest
	batchSize    int
	batchTimeout time.Duration
	workerCount  int
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	shutdownOnce sync.Once
}

// Config configures the worker pool behavior.
type Config struct {
	WorkerCount  int           // Количество воркеров
	BufferSize   int           // Размер буфера канала
	BatchSize    int           // Максимальный размер батча
	BatchTimeout time.Duration // Таймаут для накопления батча
}

// DefaultConfig returns sane defaults for the worker pool.
func DefaultConfig() Config {
	return Config{
		WorkerCount:  5,
		BufferSize:   100,
		BatchSize:    10,
		BatchTimeout: 5 * time.Second,
	}
}

// NewDeleteWorkerPool creates a new delete worker pool with the given config.
func NewDeleteWorkerPool(service DeleteService, config Config) *DeleteWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &DeleteWorkerPool{
		service:      service,
		requestChan:  make(chan DeleteRequest, config.BufferSize),
		batchSize:    config.BatchSize,
		batchTimeout: config.BatchTimeout,
		workerCount:  config.WorkerCount,
		ctx:          ctx,
		cancel:       cancel,
	}

	return pool
}

// Start launches worker goroutines to process delete requests.
func (p *DeleteWorkerPool) Start() {
	log.Info().
		Int("workers", p.workerCount).
		Int("batchSize", p.batchSize).
		Dur("batchTimeout", p.batchTimeout).
		Msg("Starting delete worker pool")

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *DeleteWorkerPool) worker(id int) {
	defer p.wg.Done()

	log.Debug().Int("workerID", id).Msg("Worker started")

	batch := make(map[string][]string) // userID -> []urlIDs
	totalURLs := 0
	var timer *time.Timer
	var timerC <-chan time.Time

	processBatch := func() {
		if len(batch) == 0 {
			return
		}

		log.Debug().
			Int("workerID", id).
			Int("users", len(batch)).
			Msg("Processing batch")

		for userID, urlIDs := range batch {
			if err := p.service.DeleteUserURLs(userID, urlIDs); err != nil {
				log.Error().
					Err(err).
					Int("workerID", id).
					Str("userID", userID).
					Int("urlCount", len(urlIDs)).
					Msg("Failed to delete user URLs")
			} else {
				log.Debug().
					Int("workerID", id).
					Str("userID", userID).
					Int("urlCount", len(urlIDs)).
					Msg("Successfully deleted user URLs")
			}
		}

		for k := range batch {
			delete(batch, k)
		}
		totalURLs = 0
	}

	startOrResetTimer := func() {
		if timer == nil {
			timer = time.NewTimer(p.batchTimeout)
			timerC = timer.C
			return
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(p.batchTimeout)
		timerC = timer.C
	}

	stopTimer := func() {
		if timer == nil {
			return
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timerC = nil
	}

	for {
		select {
		case <-p.ctx.Done():
			log.Debug().Int("workerID", id).Msg("Worker shutting down")
			processBatch()
			stopTimer()
			return

		case req, ok := <-p.requestChan:
			if !ok {
				// Канал закрыт - обрабатываем оставшиеся запросы и выходим
				log.Debug().Int("workerID", id).Msg("Request channel closed, processing remaining batch")
				processBatch()
				stopTimer()
				return
			}

			batchWasEmpty := len(batch) == 0
			batch[req.UserID] = append(batch[req.UserID], req.URLIDs...)
			totalURLs += len(req.URLIDs)

			if totalURLs >= p.batchSize {
				processBatch()
				if len(batch) == 0 {
					stopTimer()
				} else {
					startOrResetTimer()
				}
			} else if batchWasEmpty {
				startOrResetTimer()
			}

		case <-timerC:
			processBatch()
			stopTimer()
		}
	}
}

// Submit queues a delete request for processing.
func (p *DeleteWorkerPool) Submit(userID string, urlIDs []string) error {
	select {
	case <-p.ctx.Done():
		return context.Canceled
	case p.requestChan <- DeleteRequest{UserID: userID, URLIDs: urlIDs}:
		log.Debug().
			Str("userID", userID).
			Int("urlCount", len(urlIDs)).
			Msg("Delete request submitted")
		return nil
	default:
		log.Warn().
			Str("userID", userID).
			Int("urlCount", len(urlIDs)).
			Msg("Request channel is full, blocking")

		select {
		case <-p.ctx.Done():
			return context.Canceled
		case p.requestChan <- DeleteRequest{UserID: userID, URLIDs: urlIDs}:
			return nil
		}
	}
}

// Shutdown gracefully stops workers, waiting up to the provided timeout.
func (p *DeleteWorkerPool) Shutdown(timeout time.Duration) error {
	var shutdownErr error

	p.shutdownOnce.Do(func() {
		log.Info().Msg("Shutting down delete worker pool")

		close(p.requestChan)

		done := make(chan struct{})
		go func() {
			p.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			log.Info().Msg("Delete worker pool shut down gracefully")
		case <-time.After(timeout):
			log.Warn().Msg("Delete worker pool shutdown timeout, forcing shutdown")
			p.cancel()
			<-done
			shutdownErr = context.DeadlineExceeded
		}
	})

	return shutdownErr
}

// Stats returns current pool statistics.
func (p *DeleteWorkerPool) Stats() PoolStats {
	return PoolStats{
		QueueSize:   len(p.requestChan),
		QueueCap:    cap(p.requestChan),
		WorkerCount: p.workerCount,
	}
}

// PoolStats contains worker pool metrics.
type PoolStats struct {
	QueueSize   int
	QueueCap    int
	WorkerCount int
}
