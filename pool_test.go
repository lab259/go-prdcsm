package prdcsm_test

import (
	"sync"
	"time"

	. "github.com/lab259/go-prdcsm/v2"
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
	It("should run with 4 workers", func(done Done) {
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

		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40

		time.Sleep(50 * time.Millisecond)
		Expect(pool.Stop()).To(Succeed())

		Expect(called.count()).To(Equal(100))
		close(done)
	})

	It("should run with 4 workers waiting to be finished", func(done Done) {
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

		go pool.Start()

		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40

		time.Sleep(time.Millisecond * 10)

		Expect(pool.Stop()).To(Succeed())
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

		go pool.Start()

		producer.Ch <- 10
		producer.Ch <- nil
		producer.Ch <- 20
		producer.Ch <- nil
		producer.Ch <- 30
		producer.Ch <- nil
		producer.Ch <- 40
		producer.Ch <- nil

		time.Sleep(50 * time.Millisecond)
		Expect(pool.Stop()).To(Succeed())

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

		go pool.Start()

		time.Sleep(50 * time.Millisecond)

		Expect(called.count()).To(Equal(100))
		Expect(producer.Ch).To(BeClosed())
		Expect(pool.Start()).To(Equal(ErrPoolCancelled))
		close(done)
	})

	It("should stop on EOF with pending messages to be reached", func(done Done) {
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
		producer.Ch <- EOF
		producer.Ch <- 30
		producer.Ch <- 40

		go pool.Start()

		time.Sleep(50 * time.Millisecond)

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
		// 4 messages, each one longing 10 milliseconds.
		// Give times to run 1 message and cancel the producer.
		// It should check how many consumers run and if all messages where
		// flushed from the producer channel.
		var condM sync.Mutex
		cond := sync.NewCond(&condM)
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  1,
			Producer: producer,
			Consumer: func(data interface{}) {
				called.inc(data.(int))

				cond.L.Lock()
				defer cond.L.Unlock()

				// This `.Wait` makes the consumer wait for a signal. So, after
				// canceling the producer we should be able to signal and
				// guarantee only one consumer is called.
				cond.Wait()
			},
		})

		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40

		go pool.Start()

		// Wait some time to ensure only 1 consumer is being called.
		time.Sleep(time.Millisecond * 25)
		go func() {
			defer GinkgoRecover()

			// Waits more a little to ensure everything was cancelled.
			time.Sleep(time.Millisecond * 25)
			cond.Signal() // Unlocks the consumer
		}()
		producer.Cancel()

		Expect(called.count()).To(Equal(10))
		Expect(producer.Ch).To(BeClosed())
		close(done)
	})
})
