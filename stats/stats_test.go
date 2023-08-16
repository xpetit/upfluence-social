package stats

import "testing"

func makeRange(n int) (vals []int) {
	for i := 0; i < n; i++ {
		vals = append(vals, i)
	}
	return
}

func TestGetPercentile(t *testing.T) {
	expect := func(nbVals int, p float64, want int) {
		t.Helper()
		got := GetPercentile(makeRange(nbVals), p)
		if got != want {
			t.Errorf("wrong p%.f: got:%d, want:%d", p*100, got, want)
		}
	}

	// TODO/FIXME are these value wrong? check https://en.wikipedia.org/wiki/Percentile#Calculation_methods
	expect(10, .01, -1)
	expect(10, .50, 4)
	expect(10, .90, 8)
	expect(10, .99, 9)

	expect(100, .01, 0)
	expect(100, .25, 24)
	expect(100, .50, 49)
	expect(100, .90, 89)
	expect(100, .99, 98)
}
