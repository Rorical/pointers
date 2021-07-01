package main

import (
	"context"
	"fmt"
	"pointers/pointers/pubsub"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ps, err := pubsub.NewPubSub(ctx)
	if err != nil {
		panic(err)
	}
	chann, err := ps.Subscribe("testchan")
	if err != nil {
		panic(err)
	}
	go func() {
		time.Sleep(2 * time.Second)
		byes := []byte("channel?")
		chann.Publish(&byes)
	}()
	go chann.Listen()
	for {
		message, ok := <-chann.Messages
		if !ok {
			return
		}
		fmt.Println(string(*message))
	}
}
