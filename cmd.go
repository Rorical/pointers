package main

import (
	"context"
	"fmt"
	"math/rand"
	"pointers/pointers/config"
	"pointers/pointers/pubsub"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p2p, _, err := pubsub.Libp2p(ctx, &config.Libp2pConf{
		BootstrapNodes: []string{
			"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
			"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
			"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
			"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
			"/ip4/127.0.0.1/tcp/24277/p2p/12D3KooWEqAh2WrvtuVDx3uVVFbqSzPJrPMBHrZMSBczGPK4JFtM",
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

	time.Sleep(20 * time.Second)

	ps, err := pubsub.NewPubSub(ctx, p2p)
	if err != nil {
		panic(err)
	}
	chann, err := ps.Subscribe("testchan")
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
	go chann.Listen()
	go chann.ListenPeers()
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
		}

	}

}
