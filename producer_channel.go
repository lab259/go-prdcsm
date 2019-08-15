package prdcsm

import (
	"context"
	"sync"
)

// ChannelProducer implements using a channel as source of a producer. This is
// the most simplistic and, yet, powerful implementation.
type ChannelProducer struct {
	ctx        context.Context
	ctxCancel  context.CancelFunc
	closeMutex sync.Mutex
	stopped    bool
	Ch         chan interface{}
}

// NewChannelProducer returns a new ChannelProducer
func NewChannelProducer(cap int) *ChannelProducer {
	ctx, cancel := context.WithCancel(context.Background())
	return &ChannelProducer{
		ctx:       ctx,
		ctxCancel: cancel,
		Ch:        make(chan interface{}, cap),
	}
}

// Produce gets the next element of the channel.
func (producer *ChannelProducer) Produce() interface{} {
	select {
	case <-producer.ctx.Done():
		// If the ChannelProducer is cancelled.
		return EOF
	case got, ok := <-producer.Ch:
		// The channel was closed...
		if !ok {
			return EOF
		}
		return got // If everything went fine.
	}
}

// Stop stops the producer closing the channel but keep all added to the
// channel.
func (producer *ChannelProducer) Stop() {
	producer.closeMutex.Lock()
	defer producer.closeMutex.Unlock()
	if !producer.stopped {
		close(producer.Ch)
		producer.stopped = true
	}
}

// Cancel stops the producer
func (producer *ChannelProducer) Cancel() {
	producer.ctxCancel()
	producer.Stop()
	for range producer.Ch {
		// Flush all messages added in the channel.
	}
}
