package metrics

import (
	"slices"

	social "github.com/xpetit/upfluence-social"
	stats "github.com/xpetit/upfluence-social/stats"
)

type Statistics struct {
	TotalPosts       int `json:"total_posts"`
	MinimumTimestamp int `json:"minimum_timestamp"`
	MaximumTimestamp int `json:"maximum_timestamp"`
	P50              int `json:"p50"`
	P90              int `json:"p90"`
	P99              int `json:"p99"`
}

// Collect collects data points and returns statistics as soon as the channel is closed
func Collect(points chan social.DataPoint) *Statistics {
	// Aggregate data, using a map instead of a slice to lower memory usage (assuming that
	// the values are not randomly and uniformly distributed)
	var minTime, maxTime uint32
	counter := map[social.Count]int{}
	for point := range points {
		if minTime == 0 || point.Time < minTime {
			minTime = point.Time
		}
		if maxTime == 0 || point.Time > maxTime {
			maxTime = point.Time
		}
		counter[point.Count]++
	}

	// compute stats (TODO: could weighted percentiles avoid unpacking all the values?)
	var counts []social.Count
	var total int
	for count, occurrences := range counter {
		for i := 0; i < occurrences; i++ {
			counts = append(counts, count)
		}
		total += occurrences
	}
	slices.Sort(counts)

	return &Statistics{
		TotalPosts:       total,
		MinimumTimestamp: int(minTime),
		MaximumTimestamp: int(maxTime),
		P50:              int(stats.GetPercentile(counts, .5)),
		P90:              int(stats.GetPercentile(counts, .9)),
		P99:              int(stats.GetPercentile(counts, .99)),
	}
}
