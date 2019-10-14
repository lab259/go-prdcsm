package prdcsm_test

import (
	"time"

	. "github.com/lab259/go-prdcsm/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Producer Channel", func() {

	It("should stop in a goroutine", func(done Done) {
		var called safecounter
		producer := NewChannelProducer(50)
		pool := NewPool(PoolConfig{
			Workers:  5,
			Producer: producer,
			Consumer: func(data interface{}) {
				time.Sleep(time.Millisecond * 30)
				called.inc(data.(int))
			},
		})

		producer.Yield(10)
		producer.Yield(20)

		go func() {
			Expect(pool.Stop()).To(Succeed())
		}()

		producer.Yield(30)
		producer.Yield(40)

		Expect(pool.Start()).To(Succeed())

		Expect(called.count()).To(Equal(100))
		Expect(producer.GetCh()).To(BeClosed())

		Expect(called.count()).To(Equal(100))
		close(done)
	}, 5000)
})
