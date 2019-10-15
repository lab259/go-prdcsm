package prdcsm

import "sync"

// ChannelProducer implements using a channel as source of a producer. This is
// the most simplistic and, yet, powerful implementation.
type ChannelProducer struct {
	shutdown  chan struct{}
	Ch        chan interface{}
	stopMutex sync.Mutex
	stopped   bool
}

// NewChannelProducer returns a new ChannelProducer
func NewChannelProducer(cap int) *ChannelProducer {
	return &ChannelProducer{
		shutdown: make(chan struct{}),
		Ch:       make(chan interface{}, cap),
	}
}

// GetCh gets the next element of the channel.
func (producer *ChannelProducer) GetCh() <-chan interface{} {
	return producer.Ch
}

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
	close(producer.Ch)
	producer.stopped = true
}

// Cancel stops the producer
func (producer *ChannelProducer) Cancel() {
	producer.Stop()
	for range producer.Ch {
		// Flush all messages added in the channel.
	}
}
