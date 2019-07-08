package prdcsm

// ChannelProducer implements using a channel as source of a producer. This is
// the most simplistic and, yet, powerful implementation.
type ChannelProducer struct {
	Ch chan interface{}
}

// NewChannelProducer returns a new ChannelProducer
func NewChannelProducer(cap int) *ChannelProducer {
	return &ChannelProducer{
		Ch: make(chan interface{}, cap),
	}
}

// Produce gets the next element of the channel.
func (producer *ChannelProducer) Produce() interface{} {
	return <-producer.Ch
}

// Stop stops the producer closing the channel.
func (producer *ChannelProducer) Stop() {
	close(producer.Ch)
}
