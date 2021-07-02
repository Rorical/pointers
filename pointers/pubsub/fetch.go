package pubsub

import (
	"context"
	"fmt"
	"time"

	"pointers/pointers/protocol"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-msgio/protoio"
	"google.golang.org/protobuf/proto"
)

const FetchProtoID = protocol.ID("/libp2p/fetch/0.0.1")

type Fetcher struct {
	ctx context.Context
	p2p *host.Host
}

func NewFetcher(ctx context.Context, p2p *host.Host) (*Fetcher, error) {
	return &Fetcher{
		ctx: ctx,
		p2p: p2p,
	}, nil
}

func (f *Fetcher) Fetch(pid *peer.ID, key string) {
	ctx, cancel := context.WithTimeout(f.ctx, time.Second*10)
	defer cancel()
	s, err := p.host.NewStream(ctx, pid, FetchProtoID)

	ask := &protocol.Ask{
		Key: key,
	}
	sendBuff(f.ctx, s, ask)
}

func sendBuff(ctx context.Context, s network.Stream, msg proto.Message) error {
	done := make(chan error, 1)
	go func() {
		wc := protoio.NewDelimitedWriter(s)

		if err := wc.WriteMsg(msg); err != nil {
			done <- err
			return
		}

		done <- nil
	}()

	var retErr error
	select {
	case retErr = <-done:
	case <-ctx.Done():
		retErr = ctx.Err()
	}

	if retErr != nil {
		fmt.Println("error writing response to %s: %s", s.Conn().RemotePeer(), retErr)
	}
	return retErr
}

func readBuff(ctx context.Context, s network.Stream, msg proto.Message) error {
	done := make(chan error, 1)
	go func() {
		r := protoio.NewDelimitedReader(s, 1<<20)
		if err := r.ReadMsg(msg); err != nil {
			done <- err
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		s.Reset()
		return ctx.Err()
	}
}
