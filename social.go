package social

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Dimensions are the social values we generate the statistics upon.
// Treat it like a iota const enumeration: do not change order, append new values at the end.
var Dimensions = [...]string{
	"likes",
	"comments",
	"favorites",
	"retweets",
}

type Dimension uint8

func (d Dimension) String() string {
	return Dimensions[d]
}

type Event struct {
	UnixTime uint32
	ID       uint32
	Counts   [len(Dimensions)]uint32
}

func (e Event) String() string {
	var sb strings.Builder
	date := time.Unix(int64(e.UnixTime), 0).UTC().Format(time.DateTime)
	fmt.Fprintf(&sb, "%s ID:%d,", date, e.ID)
	for i, count := range e.Counts {
		fmt.Fprintf(&sb, "%s:%d,", Dimension(i), count)
	}
	s := sb.String()
	return s[:len(s)-1] // remove trailing comma
}

type EventStream struct {
	Err error // unrecoverable error, if any

	done  chan struct{}
	m     sync.RWMutex
	chans []chan *Event
}

// OpenEventStream opens an event stream, allowing clients to listen to it.
// Parsing errors are logged using the standard logger.
func OpenEventStream(address string) (*EventStream, error) {
	resp, err := http.Get(address)
	if err != nil {
		return nil, err
	}
	stream := &EventStream{done: make(chan struct{})}

	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			b := scanner.Bytes()
			if len(b) == 0 {
				continue // skip empty lines
			}

			event, err := parseEvent(b)
			if err != nil {
				log.Println(err) // skip malformatted event
				continue
			}

			// dispatch the event to the listeners
			stream.m.RLock()
			for _, c := range stream.chans {
				if c != nil {
					c <- event
				}
			}
			stream.m.RUnlock()
		}
		stream.m.Lock()
		stream.Err = scanner.Err()
		stream.m.Unlock()
		close(stream.done)
	}()

	return stream, nil
}

// parseEvent extracts an event from a JSON payload
func parseEvent(b []byte) (*Event, error) {
	payload, found := bytes.CutPrefix(b, []byte("data: "))
	if !found {
		return nil, errors.New("prefix not found")
	}

	var eventRaw map[string]map[string]json.RawMessage
	if err := json.Unmarshal(payload, &eventRaw); err != nil {
		return nil, err
	}

	for rootKey, post := range eventRaw {
		_ = rootKey // TODO: validate this? ("pin", "instagram_media", "tiktok_video", etc)

		unixTime, err := strconv.ParseUint(string(post["timestamp"]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp: %s", err)
		}

		id, err := strconv.ParseUint(string(post["id"]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid ID: %s", err)
		}

		event := Event{
			UnixTime: uint32(unixTime),
			ID:       uint32(id),
		}
		for dimensionIdx, dimension := range Dimensions {
			countRaw, found := post[dimension]
			if !found {
				continue
			}
			count, err := strconv.ParseUint(string(countRaw), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid count: %s", err)
			}
			event.Counts[dimensionIdx] = uint32(count)
		}
		return &event, nil
	}

	return nil, errors.New("empty event")
}

// ListenFor listens to the event stream for d duration.
// The channel closes as soon as the duration has elapsed, or an error occurred.
// The caller is expected to check *EventStream.Err for errors after the channel closed.
func (stream *EventStream) ListenFor(duration time.Duration) chan *Event {
	dst := make(chan *Event, 100)

	go func() {
		// find (or create) a channel spot
		stream.m.Lock()
		idx := -1
		for i, c := range stream.chans {
			if c == nil {
				idx = i
				stream.chans[idx] = dst
				break
			}
		}
		if idx == -1 {
			stream.chans = append(stream.chans, dst)
			idx = len(stream.chans) - 1
		}
		stream.m.Unlock()

		select {
		case <-time.After(duration):
		case <-stream.done:
		}

		// remove channel
		stream.m.Lock()
		close(dst)
		stream.chans[idx] = nil
		stream.m.Unlock()
	}()

	return dst
}
