package metrics_test

import (
	"testing"
	"time"

	social "github.com/xpetit/upfluence-social"
	"github.com/xpetit/upfluence-social/metrics"
)

func TestCollect(t *testing.T) {
	// use a fixed date for data points
	date, err := time.Parse(time.DateOnly, "2023-08-10")
	if err != nil {
		t.Fatal(err)
	}
	unix := uint32(date.Unix())
	minUnix := unix - 1
	maxUnix := unix + 1

	assertStatsEqual := func(nb int, want metrics.Statistics) {
		t.Helper()

		// generate data points
		var points []social.DataPoint
		for i := uint32(1); i <= uint32(nb); i++ {
			points = append(points, social.DataPoint{
				Time:  unix,
				Count: social.Count(i),
			})
		}
		// inject in the middle the minimum & maximum timestamps
		mid := len(points) / 2
		if len(points) > 0 {
			points[mid].Time = maxUnix
		}
		if len(points) > 1 {
			points[mid+1].Time = minUnix
		}

		c := make(chan social.DataPoint)
		go func() {
			for _, point := range points {
				c <- point
			}
			close(c)
		}()
		got := metrics.Collect(c)

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

	assertStatsEqual(100, metrics.Statistics{
		TotalPosts:       100,
		MinimumTimestamp: int(minUnix),
		MaximumTimestamp: int(maxUnix),
		P50:              50,
		P90:              90,
		P99:              99,
	})

	assertStatsEqual(10, metrics.Statistics{
		TotalPosts:       10,
		MinimumTimestamp: int(minUnix),
		MaximumTimestamp: int(maxUnix),
		P50:              5,
		P90:              9,
		P99:              10,
	})

	assertStatsEqual(1, metrics.Statistics{
		TotalPosts:       1,
		MinimumTimestamp: int(maxUnix),
		MaximumTimestamp: int(maxUnix),
		P50:              1,
		P90:              1,
		P99:              1,
	})

	assertStatsEqual(0, metrics.Statistics{})
}
