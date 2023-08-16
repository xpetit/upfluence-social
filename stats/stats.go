package stats

import "math"

type Integer interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int8 | ~int16 | ~int32 | ~int64 |
		int | ~uint | ~uintptr
}

// GetPercentile returns the percentile for a [0.0, 1.0] value
// It assumes a sorted list of values
func GetPercentile[Int Integer](vals []Int, percentile float64) Int {
	n := len(vals)
	switch n {
	case 0:
		return 0
	case 1:
		return vals[0]
	}

	f := percentile*float64(n) - 1
	i := int(f)
	if f == float64(i) {
		return vals[i]
	}

	lower := vals[i]
	upper := vals[i+1]

	delta := float64(upper - lower)
	ratio := f - float64(i)
	return lower + Int(math.Round(delta*ratio))
}
