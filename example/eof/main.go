package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/lab259/go-prdcsm/v3"
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
			time.Sleep(time.Millisecond * 110)
		},
		Producer: producer,
	})

	go func() {
		s := time.Now()
		for {
			if time.Since(s) > time.Second {
				producer.Produce(prdcsm.EOF)
				return
			}
			producer.Produce(rand.Int())
			time.Sleep(time.Millisecond * 10)
		}
	}()

	pool.Start()
}
