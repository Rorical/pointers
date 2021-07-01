package pubsub

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const bufSize = 128

type PubSub struct {
	ctx    context.Context
	p2p    *host.Host
	pubsub *pubsub.PubSub
}

type Subscriber struct {
	ctx      context.Context
	topic    *pubsub.Topic
	sub      *pubsub.Subscription
	Messages chan *[]byte
}

func (sub *Subscriber) Listen() {
	for {
		msg, err := sub.sub.Next(sub.ctx)
		if err != nil {
			close(sub.Messages)
			return
		}
		sub.Messages <- &msg.Data
	}
}

func (sub *Subscriber) Publish(message *[]byte) error {
	return sub.topic.Publish(sub.ctx, *message)
}

func (sub *Subscriber) Peers() []peer.ID {
	return sub.topic.ListPeers()
}

func (sub *Subscriber) ListenPeers() error {
	handler, err := sub.topic.EventHandler()
	if err != nil {
		return err
	}
	for {
		event, err := handler.NextPeerEvent(sub.ctx)
		if err != nil {
			close(sub.Messages)
			return err
		}
		if event.Type == pubsub.PeerJoin {

		}
	}
}

func NewPubSub(ctx context.Context, p2p *host.Host) (*PubSub, error) {
	pubsub, err := getPubsub(ctx, p2p)
	if err != nil {
		return nil, err
	}
	return &PubSub{
		ctx:    ctx,
		p2p:    p2p,
		pubsub: pubsub,
	}, nil
}

func (ps *PubSub) Subscribe(topicName string) (*Subscriber, error) {
	topic, err := ps.pubsub.Join(topicName)
	if err != nil {
		return nil, err
	}
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}
	suber := &Subscriber{
		ctx:      ps.ctx,
		topic:    topic,
		sub:      sub,
		Messages: make(chan *[]byte, bufSize),
	}
	return suber, nil
}

func (ps *PubSub) Publish(topicName string, message *[]byte) error {
	topic, err := ps.pubsub.Join(topicName)
	if err != nil {
		return err
	}
	return topic.Publish(ps.ctx, *message)
}
