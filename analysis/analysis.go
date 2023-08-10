package analysis

import (
	"math"
	"slices"
	"time"

	social "github.com/xpetit/upfluence-social"
)

func getPercentile(vals []uint32, percentile float64) uint32 {
	if len(vals) == 0 {
		return 0
	}

	f := percentile*float64(len(vals)) - 1
	i := int(f)
	if f == float64(i) {
		return vals[i]
	}

	lower := vals[i]
	upper := vals[i+1]

	delta := float64(upper - lower)
	ratio := f - float64(i)
	return lower + uint32(math.Round(delta*ratio))
}

type Statistics struct {
	TotalPosts       int `json:"total_posts"`
	MinimumTimestamp int `json:"minimum_timestamp"`
	MaximumTimestamp int `json:"maximum_timestamp"`
	P50              int `json:"p50"`
	P90              int `json:"p90"`
	P99              int `json:"p99"`
}

func compute(events []social.Event, dimension social.Dimension) *Statistics {
	var counts []uint32
	var minUnix, maxUnix uint32

	for _, event := range events {
		if minUnix == 0 || event.UnixTime < minUnix {
			minUnix = event.UnixTime
		}
		if maxUnix == 0 || event.UnixTime > maxUnix {
			maxUnix = event.UnixTime
		}
		counts = append(counts, event.Counts[dimension])
	}

	slices.Sort(counts)

	return &Statistics{
		TotalPosts:       len(counts),
		MinimumTimestamp: int(minUnix),
		MaximumTimestamp: int(maxUnix),
		P50:              int(getPercentile(counts, .5)),
		P90:              int(getPercentile(counts, .9)),
		P99:              int(getPercentile(counts, .99)),
	}
}

func Gather(stream *social.EventStream, duration time.Duration, dimension social.Dimension) *Statistics {
	var events []social.Event

	for event := range stream.ListenFor(duration) {
		events = append(events, event)
	}

	return compute(events, dimension)
}
