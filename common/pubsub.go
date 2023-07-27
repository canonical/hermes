package common

import (
	zmq "github.com/zeromq/goczmq"
)

const (
	IPCPath = "ipc:///tmp/hermespubsub"
)

type PubSubType int

const (
	Pub PubSubType = iota
	Sub
)

type PubSub struct {
	topic     string
	channeler *zmq.Channeler
}

func NewPubSub(pubsubType PubSubType, topic string) (*PubSub, error) {
	var channeler *zmq.Channeler

	if pubsubType == Pub {
		channeler = zmq.NewPubChanneler(IPCPath)
	} else {
		channeler = zmq.NewSubChanneler(IPCPath, topic)
	}

	return &PubSub{
		topic:     topic,
		channeler: channeler,
	}, nil
}

func (pubsub *PubSub) Send(bytes []byte) {
	pubsub.channeler.SendChan <- [][]byte{[]byte(pubsub.topic), bytes}
}

func (pubsub *PubSub) Recv() []byte {
	resp := <-pubsub.channeler.RecvChan
	return resp[1]
}

func (pubsub *PubSub) Release() {
	pubsub.channeler.Destroy()
}
