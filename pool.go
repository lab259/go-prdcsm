package prdcsm

import (
	"sync"
)

// `Pool` is the management structure for initializing the working pool.
//
// For the proper management of the workers, a channel as big as the desired
// workers count is created and populated. Hence, when a consumer is needed, but
// none is available, the pool will wait until one be available (<3 channels!).
//
// For receiving the data for processing, a `Producer` interface is used. Check
// its documentation for more information.
type Pool struct {
	Consumer      Consumer
	Producer      Producer
	consumerPool  chan Consumer
	channelClosed bool
	running       bool
	waitGroup     sync.WaitGroup
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
	pool.consumerPool = make(chan Consumer, count)
	defer func() {
		if !pool.channelClosed {
			close(pool.consumerPool)
		}
	}()

	pool.waitGroup.Add(1)
	defer pool.waitGroup.Done()

	// Initialize the worker pool
	for i := 0; i < count; i++ {
		pool.consumerPool <- pool.runWorker
	}

	pool.running = true

	// While the pool is running ...
	for pool.running {
		// Grab data from the producer
		data := pool.Producer.Produce()

		if data == nil { // Nothing to be done.
			continue
		}

		// Grabs a worker from the channel pool
		worker, more := <-pool.consumerPool
		if !more { // No more workers to process.
			break
		}
		go worker(data)
	}
}

func (pool *Pool) runWorker(data interface{}) {
	pool.waitGroup.Add(1)
	defer func() {
		pool.waitGroup.Done()
		if pool.running {
			// Returns the "worker" to the pool
			pool.consumerPool <- pool.runWorker
		}
	}()
	pool.Consumer(data)
}

// Waits the pool to stop
func (pool *Pool) Wait() {
	pool.waitGroup.Wait()
}

// Stop gracefully finalize the pool and waits all workers to be done.
//
// BE AWARE: The running workers will not be stopped. The stop will finalize the
// pool and WAIT the workers stop by themselves.
func (pool *Pool) Stop() {
	pool.running = false
	pool.channelClosed = true
	close(pool.consumerPool)
	pool.Producer.Stop()
	pool.Wait()
}
