package social

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/xpetit/upfluence-social/publish/broker"
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

type Count uint32

type DataPoint struct {
	Time  uint32
	Count Count
}

type Event struct {
	Counts   [len(Dimensions)]*Count
	UnixTime uint32
	ID       uint32
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

	done chan struct{}
	br   *broker.Broker[Dimension, DataPoint]
}

// OpenEventStream opens an event stream, allowing clients to listen to it.
// Parsing errors are logged using the standard logger.
func OpenEventStream(r io.Reader) *EventStream {
	stream := &EventStream{
		done: make(chan struct{}),
		br:   broker.New[Dimension, DataPoint](),
	}

	go func() {
		scanner := bufio.NewScanner(r)
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

			for i, count := range event.Counts {
				if count != nil {
					stream.br.Publish(Dimension(i), DataPoint{
						Time:  event.UnixTime,
						Count: *count,
					})
				}
			}
		}
		stream.Err = scanner.Err()
		close(stream.done)
	}()

	return stream
}

// parseEvent extracts an event from a JSON payload
func parseEvent(b []byte) (*Event, error) {
	payload, found := bytes.CutPrefix(b, []byte("data: "))
	if !found {
		return nil, errors.New("prefix not found")
	}
	eventWrapper := make(map[string]struct {
		ID        *uint32
		Timestamp *uint32
		Likes     *Count
		Comments  *Count
		Favorites *Count
		Retweets  *Count
	}, 1)
	if err := json.Unmarshal(payload, &eventWrapper); err != nil {
		return nil, err
	}
	if len(eventWrapper) != 1 {
		return nil, errors.New("only one rootkey expected")
	}
	for _, event := range eventWrapper {
		if event.Timestamp == nil {
			return nil, errors.New("missing timestamp")
		}
		if event.ID == nil {
			return nil, errors.New("missing ID")
		}
		return &Event{
			UnixTime: *event.Timestamp,
			ID:       *event.ID,
			Counts: [len(Dimensions)]*Count{
				event.Likes,
				event.Comments,
				event.Favorites,
				event.Retweets,
			},
		}, nil
	}
	panic("unreachable")
}

// ListenTo listens to the event stream for a given dimension and duration.
// It returns a channel passing data points.
// The channel closes as soon as the duration has elapsed, or an error occurred.
// The caller is expected to check *EventStream.Err for errors after the channel closed.
func (stream *EventStream) ListenTo(dimension Dimension, duration time.Duration) chan DataPoint {
	messages, cancel := stream.br.Subscribe(dimension)

	go func() {
		select {
		case <-stream.done:
		case <-time.After(duration):
		}
		cancel()
	}()

	return messages
}
