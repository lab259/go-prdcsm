package prdcsm

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrPoolCancelled means the Pool was cancelled and must be restarted.
	ErrPoolCancelled = errors.New("pool is already cancelled")
)

// EOF represents the end of the process. If, by any means, a producer returns
// it to the `Pool`, it will be finished.
var EOF = &struct{}{}

// Pool abstracts the behavior of a pool.
type Pool interface {
	// Start starts consuming the Producer.
	Start() error
	// Stop gracefully finalize the pool and waits all workers to be done.
	Stop() error
	// Restart gracefully finalize the current pool and waits all workers to be done. Then
	// restarts the pool.
	Restart() error
	// Wait the pool workers to stop
	Wait()
}

type pool struct {
	config           PoolConfig
	consumerPool     chan Consumer
	waitGroupWorkers sync.WaitGroup
	ctx              context.Context
	cancel           context.CancelFunc
}

// PoolConfig specify the needs to create a new Pool.
type PoolConfig struct {
	Consumer Consumer
	Producer Producer
	Workers  int
}

// NewPool returns the management structure for initializing the working pool.
//
// For the proper management of the workers, a channel as big as the desired
// workers count is created and populated. Hence, when a consumer is needed, but
// none is available, the pool will wait until one be available (<3 channels!).
//
// For receiving the data for processing, a `Producer` interface is used. Check
// its documentation for more information.
func NewPool(config PoolConfig) Pool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := pool{
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		consumerPool: make(chan Consumer, config.Workers),
	}

	// Initialize the worker pool
	for i := 0; i < pool.config.Workers; i++ {
		pool.consumerPool <- pool.runWorker
	}

	return &pool
}

// Run starts the worker pool process.
//
// First, initialize and populates a channel with workers.
//
// Second, read data from the `Data` channel;
//
// Third, grabs an available worker and delivers the data starting a new
// goroutine. The new goroutine is a does not call the worker directly. Instead,
// the private `runWorker` is called to add the `waitGroup` and return the
// worker to the channel.
func (pool *pool) Start() error {
	if pool.ctx.Err() == context.Canceled {
		return ErrPoolCancelled
	}

	defer func() {
		// Ensure context is cancelled
		pool.cancel()

		// Stop the Producer
		pool.config.Producer.Stop()
	}()

	for {
		select {
		case <-pool.ctx.Done():
			return nil
		default:
			// Grab data from the producer
			data := pool.config.Producer.Produce()

			if data == nil { // Nothing to be done.
				break
			}

			// Check the end of the
			if data == EOF {
				return nil
			}

			// Grabs a worker from the channel pool
			worker, more := <-pool.consumerPool
			if !more { // No more workers to process.
				return nil
			}

			pool.waitGroupWorkers.Add(1)
			go worker(data)
		}
	}
}

func (pool *pool) runWorker(data interface{}) {
	defer func() {
		pool.waitGroupWorkers.Done()
		pool.consumerPool <- pool.runWorker
	}()
	pool.config.Consumer(data)
}

// Wait the pool to stop
func (pool *pool) Wait() {
	pool.waitGroupWorkers.Wait()
}

// Stop gracefully finalize the pool and waits all workers to be done.
//
// BE AWARE: The running workers will not be stopped. The stop will finalize the
// pool and WAIT the workers stop by themselves.
func (pool *pool) Stop() error {
	pool.cancel()
	pool.Wait()
	return nil
}

// Restart gracefully finalize the current pool and waits all workers to be done. Then
// restarts the pool.
func (pool *pool) Restart() error {
	if err := pool.Stop(); err != nil {
		return err
	}

	pool.ctx, pool.cancel = context.WithCancel(context.Background())
	return pool.Start()
}
