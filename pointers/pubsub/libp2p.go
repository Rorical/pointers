package pubsub

import (
	"context"
	"fmt"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	routing "github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	ma "github.com/multiformats/go-multiaddr"
)

func getLibp2p(ctx context.Context) (*host.Host, error) {
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519,
		-1,
	)
	if err != nil {
		return nil, err
	}
	p2p, err := libp2p.New(
		ctx,
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic",
		),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(secio.ID, secio.New),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.DefaultTransports,
		libp2p.ConnectionManager(connmgr.NewConnManager(
			100,
			400,
			time.Minute,
		)),
		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err := dht.New(ctx, h)
			return idht, err
		}),
		libp2p.EnableAutoRelay(),
		libp2p.EnableNATService(),
	)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n[*] Your Multiaddress Is: %v/p2p/%s\n", p2p.Addrs()[0], p2p.ID().Pretty())

	var wg sync.WaitGroup
	for _, peerAddr := range []string{
		"/ip4/127.0.0.1/tcp/24277/p2p/12D3KooWEqAh2WrvtuVDx3uVVFbqSzPJrPMBHrZMSBczGPK4JFtM",
	} {
		addr, _ := ma.NewMultiaddr(peerAddr)
		peerinfo, _ := peer.AddrInfoFromP2pAddr(addr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p2p.Connect(ctx, *peerinfo); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Connection established with node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	peerChan := initMDNS(ctx, p2p, "pointers")
	go func() {
		for {
			peer := <-peerChan
			fmt.Println("Found peer:", peer, ", connecting")
			if err := p2p.Connect(ctx, peer); err != nil {
				fmt.Println("Connection failed:", err)
			}
		}
	}()
	return &p2p, nil
}

func getPubsub(ctx context.Context, p2p *host.Host) (*pubsub.PubSub, error) {
	pubsub, err := pubsub.NewGossipSub(ctx, *p2p)
	if err != nil {
		return nil, err
	}
	return pubsub, nil
}
