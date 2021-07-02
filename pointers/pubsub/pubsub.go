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
	eve      *pubsub.TopicEventHandler
	NewPeers chan *peer.ID
	Messages chan *[]byte
}

func (sub *Subscriber) Listen() error {
	for {
		msg, err := sub.sub.Next(sub.ctx)
		if err != nil {
			close(sub.Messages)
			return err
		}
		sub.Messages <- &msg.Data
	}
	return nil
}

func (sub *Subscriber) StopListen() {
	sub.sub.Cancel()
}

func (sub *Subscriber) Publish(message *[]byte) error {
	return sub.topic.Publish(sub.ctx, *message)
}

func (sub *Subscriber) Peers() []peer.ID {
	return sub.topic.ListPeers()
}

func (sub *Subscriber) ListenPeers() error {
	for {
		event, err := sub.eve.NextPeerEvent(sub.ctx)
		if err != nil {
			close(sub.NewPeers)
			return err
		}
		if event.Type == pubsub.PeerJoin {
			sub.NewPeers <- &event.Peer
		}
	}
}

func (sub *Subscriber) StopListenPeers() {
	sub.eve.Cancel()
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
	handler, err := topic.EventHandler()
	if err != nil {
		return nil, err
	}
	topic.Relay()
	suber := &Subscriber{
		ctx:      ps.ctx,
		topic:    topic,
		sub:      sub,
		eve:      handler,
		NewPeers: make(chan *peer.ID, bufSize),
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
