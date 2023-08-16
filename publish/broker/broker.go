// Package broker adds topic partitionning on top of pubsub
package broker

import (
	"sync"

	"github.com/xpetit/upfluence-social/publish"
)

type Broker[Topic comparable, Message any] struct {
	chans      map[Topic]chan Message
	publishers map[Topic]*publish.Publisher[Message]
	m          sync.Mutex
}

func New[Topic comparable, Message any]() *Broker[Topic, Message] {
	return &Broker[Topic, Message]{
		chans:      map[Topic]chan Message{},
		publishers: map[Topic]*publish.Publisher[Message]{},
	}
}

func (b *Broker[Topic, Message]) Publish(topic Topic, message Message) {
	b.m.Lock()
	c, ok := b.chans[topic]
	if !ok {
		c = make(chan Message, 100)
		b.chans[topic] = c
		b.publishers[topic] = publish.New(c)
	}
	b.m.Unlock()

	c <- message
}

func (b *Broker[Topic, Message]) Subscribe(topic Topic) (messages chan Message, cancel func()) {
	b.m.Lock()
	defer b.m.Unlock()

	publisher, ok := b.publishers[topic]
	if !ok {
		c := make(chan Message, 100)
		b.chans[topic] = c
		publisher = publish.New(c)
		b.publishers[topic] = publisher
	}
	return publisher.Subscribe()
}

// func (b *Broker[Topic, _]) closeTopic(topic Topic) {
// 	close(b.chans[topic]) // closing the publishing channel will close subscribers
// 	delete(b.chans, topic)
// 	delete(b.publishers, topic)
// }

// func (b *Broker[Topic, _]) CloseTopic(topic Topic) {
// 	b.m.Lock()
// 	defer b.m.Unlock()

// 	b.closeTopic(topic)
// }

// func (b *Broker[_, _]) Close() {
// 	b.m.Lock()
// 	defer b.m.Unlock()

// 	for topic := range b.chans {
// 		b.closeTopic(topic)
// 	}
// }
