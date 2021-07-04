package pubsub

import (
	"context"
	"fmt"
	"time"

	protos "pointers/pointers/protocol"

	"google.golang.org/protobuf/proto"

	"github.com/Rorical/go-msgio/protoio"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Fetcher struct {
	ctx context.Context
	p2p *host.Host
}

type Response func(key string) (*protos.Operate, error)

func NewFetcher(ctx context.Context, p2p *host.Host) *Fetcher {
	fetcher := &Fetcher{
		ctx: ctx,
		p2p: p2p,
	}
	return fetcher
}

func (f *Fetcher) Feed(res Response) {
	(*f.p2p).SetStreamHandler(protos.FetchProtocol, func(s network.Stream) {
		f.receive(s, res)
	})
}

func (f *Fetcher) Fetch(pid *peer.ID, key string) (*protos.Operate, error) {
	ctx, cancel := context.WithTimeout(f.ctx, time.Second*10)
	defer cancel()
	s, err := (*f.p2p).NewStream(ctx, *pid, protos.FetchProtocol)
	if err != nil {
		return nil, err
	}
	ask := &protos.Ask{
		Key: key,
	}

	if err := sendBuff(f.ctx, s, ask); err != nil {
		s.Reset()
		return nil, err
	}

	if err := s.CloseWrite(); err != nil {
		s.Reset()
		return nil, err
	}

	res := &protos.Answer{}

	if err := readBuff(f.ctx, s, res); err != nil {
		_ = s.Reset()
		return nil, err
	}

	switch res.Status {
	case protos.Status_Ok:
		return res.Op, nil
	case protos.Status_NotExist:
		return nil, nil
	default:
		return nil, protos.UnknowProtocolError
	}
}

func (f *Fetcher) receive(s network.Stream, getResponse Response) {
	defer s.Close()

	ask := &protos.Ask{}
	if err := readBuff(f.ctx, s, ask); err != nil {
		s.Reset()
		return
	}

	res, err := getResponse(ask.Key)

	var response *protos.Answer
	if err == nil {
		response = &protos.Answer{
			Status: protos.Status_Ok,
			Op:     res,
		}
	} else {
		response = &protos.Answer{
			Status: protos.Status_NotExist,
		}

	}
	if err := sendBuff(f.ctx, s, response); err != nil {
		s.Reset()
		return
	}
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
		fmt.Printf("error writing response to %s: %s", s.Conn().RemotePeer(), retErr)
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
