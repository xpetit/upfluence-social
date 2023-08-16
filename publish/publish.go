package publish

import "sync"

// Publisher implements a pubsub messaging system.
type Publisher[Message any] struct {
	chans []chan Message
	mu    sync.RWMutex
}

// New returns a new Publisher that publishes messages received from the messages channel.
// Subcribers channels are closed as soon as the messages channel is closed.
func New[Message any](messages chan Message) *Publisher[Message] {
	p := Publisher[Message]{
		chans: make([]chan Message, 0, 256), // preallocate
	}

	go func() {
		// broadcast messages
		for message := range messages {
			p.mu.RLock()
			for _, c := range p.chans {
				if c != nil {
					c <- message
				}
			}
			p.mu.RUnlock()
		}

		// close all channels
		p.mu.Lock()
		for i, c := range p.chans {
			if c != nil {
				close(c)
				p.chans[i] = nil
			}
		}
		p.mu.Unlock()
	}()

	return &p
}

// Subscribe returns a channel that will forward all published messages, and a function to
// cancel the subscription and close the channel.
func (br *Publisher[Message]) Subscribe() (messages chan Message, cancel func()) {
	br.mu.Lock()
	defer br.mu.Unlock()

	messages = make(chan Message, 100)

	// find a channel spot or append a new one
	pos := -1
	for i, c := range br.chans {
		if c == nil {
			pos = i
			br.chans[pos] = messages
			break
		}
	}
	if pos == -1 {
		br.chans = append(br.chans, messages)
		pos = len(br.chans) - 1
	}

	var once sync.Once

	return messages, func() {
		once.Do(func() {
			br.mu.Lock()
			defer br.mu.Unlock()

			if c := br.chans[pos]; c != nil {
				close(c)
			}
			br.chans[pos] = nil

			if end := len(br.chans) - 1; pos == end {
				// right trim nil channels
				for ; end >= 0 && br.chans[end] == nil; end-- {
				}
				br.chans = br.chans[:end+1]
			}
		})
	}
}
