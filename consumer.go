package prdcsm

// Consumer is a blocking function that receive data and process it. This
// function can be called in parallel and should be well designed to be
// thread-safe.
//
// TODO Document the panic approach.
type Consumer func(data interface{})
