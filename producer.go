package prdcsm

// Producer abstracts the behavior or receiving data to be processed.
//
// `Producer` will return `nil` if there is no work to be done.
type Producer interface {
	// Produce returns  the data to be processed. It shall return `nil` if there
	// is no.
	Produce() interface{}

	// Stop should cancel the process of producing new items keeping the ones
	// already produced.
	Stop()

	// Cancel should cancel the process of producing new items discarding the
	// the messages already produced.
	Cancel()
}
