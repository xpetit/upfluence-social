package analysis

import (
	"testing"
	"time"

	social "github.com/xpetit/upfluence-social"
)

func TestCompute(t *testing.T) {
	// use a fixed date for events
	date, err := time.Parse(time.DateOnly, "2023-08-10")
	if err != nil {
		t.Fatal(err)
	}
	unix := uint32(date.Unix())
	minUnix := unix - 1
	maxUnix := unix + 1

	assertStatsEqual := func(nb int, want Statistics) {
		t.Helper()

		// generate events
		var events []*social.Event
		for i := uint32(1); i <= uint32(nb); i++ {
			events = append(events, &social.Event{
				UnixTime: unix,
				ID:       i,
				Counts:   [len(social.Dimensions)]uint32{i},
			})
		}
		// inject in the middle the minimum & maximum timestamps
		mid := len(events) / 2
		if len(events) > 0 {
			events[mid].UnixTime = maxUnix
		}
		if len(events) > 1 {
			events[mid+1].UnixTime = minUnix
		}

		const likes = social.Dimension(0)
		got := compute(events, likes)

		assertEqual := func(name string, got, want int) {
			t.Helper()
			if got != want {
				t.Errorf("wrong %s: got:%d, want:%d", name, got, want)
			}
		}
		assertEqual("total posts", got.TotalPosts, want.TotalPosts)
		assertEqual("minimum timestamp", got.MinimumTimestamp, want.MinimumTimestamp)
		assertEqual("maximum timestamp", got.MaximumTimestamp, want.MaximumTimestamp)
		assertEqual("p50", got.P50, want.P50)
		assertEqual("p90", got.P90, want.P90)
		assertEqual("p99", got.P99, want.P99)
	}

	assertStatsEqual(100, Statistics{
		TotalPosts:       100,
		MinimumTimestamp: int(minUnix),
		MaximumTimestamp: int(maxUnix),
		P50:              50,
		P90:              90,
		P99:              99,
	})

	assertStatsEqual(10, Statistics{
		TotalPosts:       10,
		MinimumTimestamp: int(minUnix),
		MaximumTimestamp: int(maxUnix),
		P50:              5,
		P90:              9,
		P99:              10,
	})

	assertStatsEqual(1, Statistics{
		TotalPosts:       1,
		MinimumTimestamp: int(maxUnix),
		MaximumTimestamp: int(maxUnix),
		P50:              1,
		P90:              1,
		P99:              1,
	})

	assertStatsEqual(0, Statistics{})
}
