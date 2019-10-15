package prdcsm_test

import (
	"sync"
	"time"

	. "github.com/lab259/go-prdcsm/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type safecounter struct {
	sync.RWMutex
	c int
}

func (i *safecounter) inc(j ...int) {
	i.Lock()
	if len(j) > 0 {
		i.c += j[0]
	} else {
		i.c += 1
	}
	i.Unlock()
}

func (i *safecounter) count() int {
	i.RLock()
	defer i.RUnlock()
	return i.c
}

var _ = Describe("Pool", func() {
	It("should run with multiple workers", func(done Done) {
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  4,
			Producer: producer,
			Consumer: func(data interface{}) {
				called.inc(data.(int))
			},
		})

		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40
		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40
		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40

		go func() {
			time.Sleep(time.Millisecond * 10)
			Expect(pool.Stop()).To(Succeed())
		}()

		pool.Start()

		Expect(called.count()).To(Equal(300))
		close(done)
	})

	It("should run with multiple workers with a 30ms task to run", func(done Done) {
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  4,
			Producer: producer,
			Consumer: func(data interface{}) {
				time.Sleep(time.Millisecond * 30)
				called.inc(data.(int))
			},
		})

		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40

		go func() {
			time.Sleep(time.Millisecond * 10)
			Expect(pool.Stop()).To(Succeed())
		}()

		pool.Start()

		Expect(called.count()).To(Equal(100))
		Expect(producer.Ch).To(BeClosed())

		Expect(called.count()).To(Equal(100))
		close(done)
	})

	It("should start and stop without any consumer data", func(done Done) {
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  4,
			Producer: producer,
			Consumer: func(data interface{}) {
				called.inc(data.(int))
			},
		})

		go pool.Start()

		Expect(pool.Stop()).To(Succeed())

		Expect(called.count()).To(Equal(0))
		close(done)
	})

	It("should ignore nils", func(done Done) {
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  4,
			Producer: producer,
			Consumer: func(data interface{}) {
				called.inc(data.(int))
			},
		})

		producer.Ch <- 10
		producer.Ch <- nil
		producer.Ch <- 20
		producer.Ch <- nil
		producer.Ch <- 30
		producer.Ch <- nil
		producer.Ch <- 40
		producer.Ch <- nil

		go func() {
			time.Sleep(time.Millisecond * 10)
			Expect(pool.Stop()).To(Succeed())
		}()

		pool.Start()

		Expect(called.count()).To(Equal(100))
		close(done)
	})

	It("should stop on EOF", func(done Done) {
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  4,
			Producer: producer,
			Consumer: func(data interface{}) {
				called.inc(data.(int))
			},
		})

		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40
		producer.Ch <- EOF

		pool.Start()

		Expect(called.count()).To(Equal(100))
		Expect(producer.Ch).To(BeClosed())
		close(done)
	})

	It("should stop on EOF with pending messages to be reached", func(done Done) {
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  1,
			Producer: producer,
			Consumer: func(data interface{}) {
				called.inc(data.(int))
			},
		})

		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- EOF
		producer.Ch <- 30
		producer.Ch <- 40

		pool.Start()

		Expect(called.count()).To(Equal(30))
		Expect(producer.Ch).To(HaveLen(2))
		n, ok := <-producer.Ch
		Expect(n).To(Equal(30))
		Expect(ok).To(BeTrue())
		n, ok = <-producer.Ch
		Expect(n).To(Equal(40))
		Expect(ok).To(BeTrue())
		Expect(producer.Ch).To(BeClosed())
		close(done)
	})

	It("should cancel producer discarding enqueued messages", func(done Done) {
		// # Summary (Not all steps are executed in order in the code. But the logic is this.)
		// 1. Creates a pool with 1 worker
		// 2. Enqueue 4 entries
		// 3. Makes the consumer to process the 1st entry
		// 4. Cancel the pool
		// 5. Check if the only 1st entry were processed.

		// This channel will control the Consumer controlling whe it should be
		// unlocked.
		consumerChForWaiting := make(chan bool)

		// 1. Creates a pool with 1 worker
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  1,
			Producer: producer,
			Consumer: func(data interface{}) {
				// This enables the test to control how much time a consumer
				// will spend on each entry.
				<-consumerChForWaiting

				called.inc(data.(int))
			},
		})

		// 2. Enqueue 4 entries
		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40

		go func() {
			defer GinkgoRecover()

			// Waits a pool to starts and the worker gets the 1st entry.
			time.Sleep(time.Millisecond * 10)

			// 4. Cancel the pool
			// Yes, it is being cancelled before. It happens because the worker
			// is already running and just waiting for reading the
			// `consumerChForWaiting`.
			producer.Cancel()

			// 3. Makes the consumer to process the 1st entry
			// Since we already have the consumer running (just waiting for
			// reading from the channel). This write will makes it proceed.
			consumerChForWaiting <- true // Unlocks the consumer
		}()
		pool.Start()

		Expect(called.count()).To(Equal(10))
		Expect(producer.Ch).To(BeClosed())
		close(done)
	})

	It("should interrupt a pool canceling its execution", func(done Done) {
		// # Summary (Not all steps are executed in order in the code. But the logic is this.)
		// 1. Creates a pool with 1 worker
		// 2. Enqueue 4 entries
		// 3. Process 2 entries
		// 4. Cancel the pool
		// 5. Check if the only 1st and 2nd entry were processed.

		// This channel will control the Consumer controlling whe it should be
		// unlocked.
		consumerChForWaiting := make(chan bool)

		// 1. Creates a pool with 1 worker
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  1,
			Producer: producer,
			Consumer: func(data interface{}) {
				called.inc(data.(int))

				// This enables the test to control how much time a consumer
				// will spend on each entry.
				<-consumerChForWaiting
			},
		})

		// 2. Enqueue 4 entries
		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40

		go func() {
			// 3. Process 2 entries (Part 1)
			// Signal the consumer to proceed.
			consumerChForWaiting <- true

			// Waits a little for the `select` (inside the pool implementation)
			// to receive the next entry of the producer. Without the sleep, the
			// cancel will close the `shutdown` channel from the pool making
			// the process of the 2nd entry impossible.
			time.Sleep(time.Millisecond * 10)

			// 4. Cancel the pool
			pool.Cancel()

			// 3. Process 2 entries (Part 2)
			// Signal the consumer to proceed again. This will make the worker shutdown.
			consumerChForWaiting <- true
		}()

		pool.Start()

		// 5. Check if the only 1st and 2nd entry were processed.
		Expect(called.count()).To(Equal(30))
		Expect(producer.Ch).To(BeClosed())
		close(done)
	})

	It("should interrupt a pool canceling its execution", func(done Done) {
		// # Summary (Not all steps are executed in order in the code. But the logic is this.)
		// 1. Creates a pool with 1 worker
		// 2. Enqueue 4 entries
		// 3. Process the 1st entry
		// 4. Cancel the pool
		// 5. Check if the none was processed.

		// This channel will control the Consumer controlling whe it should be
		// unlocked.
		consumerChForWaiting := make(chan bool)

		// 1. Creates a pool with 1 worker
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  1,
			Producer: producer,
			Consumer: func(data interface{}) {
				<-consumerChForWaiting

				called.inc(data.(int))
			},
		})

		go func() {
			// 2. Enqueue 4 entries
			producer.Ch <- 10
			producer.Ch <- 20
			producer.Ch <- 30
			producer.Ch <- 40

			// Waits a little before asking the customer to proceed.
			time.Sleep(time.Millisecond * 10)

			// 3. Process the 1st entry
			consumerChForWaiting <- true

			// 4. Cancel the pool
			pool.Cancel()
		}()

		pool.Start()

		// 5. Check if the none was processed.
		Expect(called.count()).To(Equal(10))
		Expect(producer.Ch).To(BeClosed())
		close(done)
	})

	It("should wait the worker to finish after canceling", func(done Done) {
		// # Summary (Not all steps are executed in order in the code. But the logic is this.)
		// 1. Creates a pool with 1 worker
		// 2. Enqueue 4 entries
		// 3. Process 1 entry and starts processing the 2nd
		// 4. Cancel the pool
		// 5. Wait Consumers to finish.
		// 6. Tells the Consumer to finish processing the 2nd entry.
		// 7. Check if the 2nd entry was finished.

		// This channel will control the Consumer controlling whe it should be
		// unlocked.
		consumerChForWaiting := make(chan bool)

		// 1. Creates a pool with 1 worker
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  1,
			Producer: producer,
			Consumer: func(data interface{}) {
				// This enables the test to control how much time a consumer
				// will spend on each entry.
				<-consumerChForWaiting

				called.inc(data.(int))
			},
		})

		// 2. Enqueue 4 entries
		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40

		go pool.Start()

		// 3. Process 1 entry and starts processing the 2nd
		consumerChForWaiting <- true

		// 4. Cancel the pool
		pool.Cancel()

		go func() {
			// 6. Tells the Consumer to finish processing the 2nd entry.
			time.Sleep(time.Millisecond * 50)

			// Signal the Consumer to process the entry.
			consumerChForWaiting <- true
		}()

		// 5. Wait Consumers to finish.
		pool.Wait()

		// 7. Check if the 2nd entry was finished.
		Expect(called.count()).To(Equal(30))
		Expect(producer.Ch).To(BeClosed())
		close(done)
	})

})
