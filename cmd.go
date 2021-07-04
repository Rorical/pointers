package main

import (
	"context"
	"fmt"
	"math/rand"
	"pointers/pointers/config"
	"pointers/pointers/protocol"
	"pointers/pointers/pubsub"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p2p, _, err := pubsub.Libp2p(ctx, &config.Libp2pConf{
		BootstrapNodes: []string{
			//"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
			//"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
			//"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
			//"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
			//"/ip4/127.0.0.1/tcp/1456/p2p/12D3KooWJXvsS2C1ykfVatAWGJBxgofCMRkUUChgDPWAw4uGnynM",
			"/ip4/127.0.0.1/tcp/26458/p2p/12D3KooW9rjP3f4zQJBU7QCLrbHzb5tAK9CY6FZ7xq4Fc925NWFx",
		},
		ListenAddrs: []string{
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/tcp/0/quic",
		},
		GroupName: "pointer",
	})
	if err != nil {
		panic(err)
	}

	ps, err := pubsub.NewPubSub(ctx, p2p)
	if err != nil {
		panic(err)
	}
	chann, err := ps.Subscribe("miaochannel")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			time.Sleep(20 * time.Second)
			byes := []byte("Publish:" + fmt.Sprint(rand.Intn(1000000)))
			chann.Publish(&byes)
		}

	}()

	fetcher := pubsub.NewFetcher(ctx, p2p)
	fetcher.Feed(func(key string) (*protocol.Operate, error) {
		val := "hello"
		return &protocol.Operate{
			Op:    protocol.Operations_Mut,
			Value: &val,
			Sign:  "MiaoMiaoMiao",
			Time:  398012380,
		}, nil
	})

	go func() {
		for {
			select {
			case message, ok := <-chann.Messages:
				if !ok {
					return
				}
				fmt.Println(string(*message))
			case nodeid, ok := <-chann.NewPeers:
				if !ok {
					return
				}
				fmt.Println("topic peer: ", nodeid)
				go func() {
					ops, err := fetcher.Fetch(nodeid, "testchan")
					if err != nil {
						panic(err)
					}
					fmt.Println(ops)
				}()
			}

		}
	}()

	go chann.ListenPeers()
	chann.Listen()

}
