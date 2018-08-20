package main

import (
	"../.."
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

	producer := &prdcsm.ChannelProducer{
		Ch: make(chan interface{}, 50),
	}

	i := 0
	pool := prdcsm.Pool{
		Consumer: func(data interface{}) {
			fmt.Println(data, " ")
			i++
			time.Sleep(time.Millisecond * 110)
		},
		Producer: producer,
	}

	go func() {
		for {
			producer.Ch <- rand.Int()
			time.Sleep(time.Millisecond * 10)
		}
	}()

	pool.Run(4)
}