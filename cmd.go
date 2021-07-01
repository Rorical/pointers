package main

import (
	"context"
	"fmt"
	"pointers/pointers/config"
	"pointers/pointers/pubsub"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p2p, err := pubsub.Libp2p(ctx, &config.Libp2pConf{
		BootstrapNodes: []string{
			"/ip4/127.0.0.1/tcp/24277/p2p/12D3KooWEqAh2WrvtuVDx3uVVFbqSzPJrPMBHrZMSBczGPK4JFtM",
		},
		ListenAddrs: []string{
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/tcp/0/quic",
		},
		MDnsName: "pointer",
	})
	if err != nil {
		panic(err)
	}
	ps, err := pubsub.NewPubSub(ctx, p2p)
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
