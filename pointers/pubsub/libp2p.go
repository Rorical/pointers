package pubsub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pointers/pointers/config"

	libp2p "github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	routing "github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"
)

func Libp2p(ctx context.Context, config *config.Libp2pConf) (*host.Host, *dht.IpfsDHT, error) {
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519,
		-1,
	)
	if err != nil {
		return nil, nil, err
	}
	var idht *dht.IpfsDHT
	p2p, err := libp2p.New(
		ctx,
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(config.ListenAddrs...),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.DefaultTransports,
		libp2p.ConnectionManager(connmgr.NewConnManager(
			100,
			400,
			time.Minute,
		)),
		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h)
			return idht, err
		}),
		libp2p.EnableAutoRelay(),
		libp2p.EnableNATService(),
	)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("\n[*] Listening at %v/p2p/%s\n", p2p.Addrs()[0], p2p.ID().Pretty())

	p2p = rhost.Wrap(p2p, idht)

	peerChan := initMDNS(ctx, p2p, config.GroupName)
	go func() {
		for {
			peer := <-peerChan
			fmt.Println("Found peer:", peer, ", connecting")
			if err := p2p.Connect(ctx, peer); err != nil {
				fmt.Println("Connection failed:", err)
			}
		}
	}()

	err = idht.Bootstrap(ctx)
	if err != nil {
		return nil, nil, err
	}
	routingDiscovery := discovery.NewRoutingDiscovery(idht)
	discovery.Advertise(ctx, routingDiscovery, config.GroupName)
	peerChan2, err := routingDiscovery.FindPeers(ctx, config.GroupName)
	if err != nil {
		return nil, nil, err
	}
	go func() {
		for {
			peer := <-peerChan2
			if peer.ID != "" && peer.ID != p2p.ID() {
				fmt.Println("Found peer:", peer.ID, ", connecting")
				if err := p2p.Connect(ctx, peer); err != nil {
					fmt.Println("Connection failed:", err)
				}
			}
		}
	}()

	var wg sync.WaitGroup
	for _, peerAddr := range config.BootstrapNodes {
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

	return &p2p, idht, nil
}

func getPubsub(ctx context.Context, p2p *host.Host) (*pubsub.PubSub, error) {
	pubsub, err := pubsub.NewGossipSub(ctx, *p2p)
	if err != nil {
		return nil, err
	}
	return pubsub, nil
}
