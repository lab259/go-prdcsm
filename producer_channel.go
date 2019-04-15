package prdcsm

// ChannelProducer implements using a channel as source of a producer. This is
// the most simplistic and, yet, powerful implementation.
type ChannelProducer struct {
	Ch chan interface{}
}

// Produce gets the next element of the channel.
func (producer *ChannelProducer) Produce() interface{} {
	return <-producer.Ch
}

func (producer *ChannelProducer) Stop() {
	close(producer.Ch)
}
