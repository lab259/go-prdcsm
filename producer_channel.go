package prdcsm

import "sync"

// ChannelProducer implements using a channel as source of a producer. This is
// the most simplistic and, yet, powerful implementation.
type ChannelProducer struct {
	shutdown  chan struct{}
	ch        chan interface{}
	stopMutex sync.RWMutex
	stopped   bool
}

// NewChannelProducer returns a new ChannelProducer
func NewChannelProducer(cap int) *ChannelProducer {
	return &ChannelProducer{
		shutdown: make(chan struct{}),
		ch:       make(chan interface{}, cap),
	}
}

// Produce store data on the channel.
func (producer *ChannelProducer) Produce(data interface{}) {
	// producer.stopMutex.RLock()
	producer.ch <- data
	// producer.stopMutex.RUnlock()
}

// GetCh gets the next element of the channel.
func (producer *ChannelProducer) GetCh() <-chan interface{} {
	return producer.ch
}

// GetShutdown gets the shutdown element of the channel.
func (producer *ChannelProducer) GetShutdown() <-chan struct{} {
	return producer.shutdown
}

// Stop stops the producer closing the channel but keep all added to the
// channel.
func (producer *ChannelProducer) Stop() {
	producer.stopMutex.Lock()
	defer producer.stopMutex.Unlock()

	if producer.stopped {
		return
	}
	close(producer.ch)
	producer.stopped = true
}

// Cancel stops the producer
func (producer *ChannelProducer) Cancel() {
	producer.Stop()
	for range producer.ch {
		// Flush all messages added in the channel.
	}
}
