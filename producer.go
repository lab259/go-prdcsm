package prdcsm

// Producer abstracts the behavior or receiving data to be processed.
//
// `Producer` will return `nil` if there is no work to be done.
type Producer interface {
	// GetCh returns a channel that will receive all produced messages.
	GetCh() <-chan interface{}

	// GetShutdown return a channel that will be closed when the producer is
	// cancelled.
	GetShutdown() <-chan struct{}

	// Stop should cancel the process of producing new items keeping the ones
	// already produced.
	//
	// It should close the `GetCh` channel. But, not clear it.
	Stop()

	// Cancel should cancel the process of producing new items discarding the
	// the messages already produced.
	Cancel()
}
