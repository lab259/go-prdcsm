# prdcsm [![CircleCI](https://circleci.com/gh/lab259/go-prdcsm.svg?style=shield)](https://circleci.com/gh/lab259/go-prdcsm) [![Go Report Card](https://goreportcard.com/badge/github.com/lab259/go-prdcsm)](https://goreportcard.com/report/github.com/lab259/go-prdcsm) [![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=shield)](http://godoc.org/github.com/lab259/go-prdcsm) [![Release](https://img.shields.io/github/release/lab259/go-prdcsm.svg?style=shield)](https://github.com/lab259/go-prdcsm/releases/latest)

prdcsm implements a wrapper for working with producers and consumers in
Go.

This is a simple implementation that aims to standardize the behaviors
consumer and producer applications.

## Instalation

```bash
$ go get github.com/lab259/go-prdcsm/v3
```

## Getting started

`prdcsm` needs 3 things:
1. A producer

A producer is a structure that will generate the data that will be processed by
the consumers. It will keep the information about how much data it can allocate
and is responsible for transferring the information through all consumers that
might be running (the default implementation makes use of channels). A pool will
have only one producer (at least for now).

2. A consumer

A consumer is a function that will be responsible for receiving the information
from a producer and process it. Usually, the pool will have many consumers.

3. The pool

The pool works like a manager that starts the consumers and keeps forwarding for
them all work done by the producer.

---

In the next example you will be able to see all three main components working
together.

```go
package main

import (
    "github.com/lab259/go-prdcsm/v3"
    "time"
    "math/rand"
    "bufio"
    "os"
    "fmt"
)

func main() {
    fmt.Println("Hit <enter> to start. Hit <Ctrl+C> to terminate.")
    reader := bufio.NewReader(os.Stdin)
    reader.ReadString('\n')

    producer := prdcsm.NewChannelProducer(50)

    i := 0
    pool := prdcsm.NewPool(prdcsm.PoolConfig{
        Workers: 4,
        Consumer: func(data interface{}) {
            fmt.Println(data, " ")
            i++
            time.Sleep(time.Millisecond * 110)  // Simulate a "heavy" consumer proccess.
        },
        Producer: producer,
    })


    go func() {
        for {
            producer.Yield(rand.Int())
            time.Sleep(time.Millisecond * 10) // Simulate a "heavy" producer proccess.
        }
    }()

    pool.Start() // Starts running the workers
}
```

## License

MIT License

Copyright (c) 2018 lab259

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
