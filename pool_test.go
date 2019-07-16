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

		go pool.Start()

		producer.Ch <- 10
		producer.Ch <- 20
		producer.Ch <- 30
		producer.Ch <- 40
		producer.Ch <- EOF

		time.Sleep(50 * time.Millisecond)

		Expect(called.count()).To(Equal(100))
		Expect(producer.Ch).To(BeClosed())
		Expect(pool.Start()).To(Equal(ErrPoolCancelled))
		close(done)
	})
})
