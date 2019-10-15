package prdcsm

import (
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
	// Stop gracefully finalize the pool and waits all workers to be done. All
	// not processed data will be processed before leaving.
	Stop() error
	// Cancel, as the Stop function, gracefully finalizes the pool and waits all
	// workers to be done. But, all not processed produced data will be thrown
	// away.
	Cancel() error
	// Restart gracefully finalize the current pool and waits all workers to be done. Then
	// restarts the pool.
	Restart() error
	// Wait the pool workers to stop
	Wait()
}

type pool struct {
	config                   PoolConfig
	waitGroupWorkersForStart sync.WaitGroup
	waitGroupWorkers         sync.WaitGroup
	shutdown                 chan struct{}
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
	pool := pool{
		config:   config,
		shutdown: make(chan struct{}, 0),
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
func (p *pool) Start() error {
	p.waitGroupWorkersForStart.Add(p.config.Workers)
	p.waitGroupWorkers.Add(p.config.Workers)
	for i := 0; i < p.config.Workers; i++ {
		go p.runWorker()
	}
	p.waitGroupWorkersForStart.Wait()
	return nil
}

func (p *pool) runWorker() {
	defer func() {
		p.waitGroupWorkersForStart.Done()
		p.waitGroupWorkers.Done()
	}()

	producerCh := p.config.Producer.GetCh()
	producerChShutdown := p.config.Producer.GetShutdown()

	for { // Keep the runWorker running...
		select {
		case <-p.shutdown:
			// The pool was cancelled.
			return
		case <-producerChShutdown:
			// The producer was cancelled. All produced data is ignored.
			return
		case data, ok := <-producerCh:
			if !ok {
				// When the producer channel is closed, the worker will halt.
				return
			}

			// nil data should be ignored
			if data == nil {
				break // goes to the beginning of the for loop.
			}

			if data == EOF { // EOF means that nothing else should be processed.
				// Stop will leave the enqueued data in the channel. At least
				// it should. Depends on the Producer implementation.
				p.config.Producer.Stop()
				return
			}

			p.config.Consumer(data)
		}
	}
}

// Wait the pool to stop
func (p *pool) Wait() {
	p.waitGroupWorkers.Wait()
}

// Stop gracefully finalize the pool and waits all workers to be done. All data
// produced still in the queue and
//
// BE AWARE: The running workers will not be stopped. The stop will finalize the
// pool and WAIT the workers stop by themselves.
func (p *pool) Stop() error {
	p.config.Producer.Stop()
	return nil
}

// Cancel gracefully finalize the pool and waits all workers to be done. But, all
//
// BE AWARE: The running workers will not be stopped. The stop will finalize the
// pool and WAIT the workers stop by themselves.
func (p *pool) Cancel() error {
	close(p.shutdown) // Cancel the execution of all workers.
	p.config.Producer.Cancel()
	return nil
}

// Restart gracefully finalize the current pool and waits all workers to be done. Then
// restarts the pool.
func (p *pool) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	return p.Start()
}
