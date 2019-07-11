package prdcsm

import (
	"context"
	"sync"
)

// EOF represents the end of the process. If, by any means, a producer returns
// it to the `Pool`, it will be finished.
var EOF = &struct{}{}

// Pool is the management structure for initializing the working pool.
//
// For the proper management of the workers, a channel as big as the desired
// workers count is created and populated. Hence, when a consumer is needed, but
// none is available, the pool will wait until one be available (<3 channels!).
//
// For receiving the data for processing, a `Producer` interface is used. Check
// its documentation for more information.
type Pool struct {
	Consumer         Consumer
	Producer         Producer
	consumerPool     chan Consumer
	waitGroupWorkers sync.WaitGroup
	ctx              context.Context
	cancel           context.CancelFunc
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
func (pool *Pool) Run(count int) {
	pool.ctx, pool.cancel = context.WithCancel(context.Background())

	pool.consumerPool = make(chan Consumer, count)
	defer func() {
		// Ensure context is cancelled
		pool.cancel()

		// Stop the Producer
		pool.Producer.Stop()

		// Close the consumerPool
		close(pool.consumerPool)
	}()

	// Initialize the worker pool
	for i := 0; i < count; i++ {
		pool.consumerPool <- pool.runWorker
	}

	for {
		select {
		case <-pool.ctx.Done():
			return
		default:
			// Grab data from the producer
			data := pool.Producer.Produce()

			if data == nil { // Nothing to be done.
				continue
			}

			// Check the end of the
			if data == EOF {
				return
			}

			// Grabs a worker from the channel pool
			worker, more := <-pool.consumerPool
			if !more { // No more workers to process.
				return
			}
			pool.waitGroupWorkers.Add(1)
			go worker(data)
		}
	}
}

func (pool *Pool) runWorker(data interface{}) {
	defer func() {
		pool.waitGroupWorkers.Done()
		if pool.ctx.Err() == nil {
			// Returns the "worker" to the pool
			pool.consumerPool <- pool.runWorker
		}
	}()
	pool.Consumer(data)
}

// Wait the pool to stop
func (pool *Pool) Wait() {
	pool.waitGroupWorkers.Wait()
}

// Stop gracefully finalize the pool and waits all workers to be done.
//
// BE AWARE: The running workers will not be stopped. The stop will finalize the
// pool and WAIT the workers stop by themselves.
func (pool *Pool) Stop() {
	if pool.ctx != nil && pool.ctx.Err() == nil {
		pool.cancel()
	}
	pool.Wait()
}
