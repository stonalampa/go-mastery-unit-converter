package convert_test

import (
	"errors"
	"math"
	"testing"

	"github.com/stonalampa/go-mastery-unit-converter/internal/convert"
)

var categories = []convert.Category{convert.Length, convert.Weight, convert.Temperature}

// FuzzConvertRoundTrip checks that converting a value from Unit A to Unit B
// and back returns the original value (within float tolerance), and that
// rejected values are rejected with one of the domain's errors.
func FuzzConvertRoundTrip(f *testing.F) {
	// Seed corpus: every Unit pair of every Category with a few typical and
	// edge values. Plain `go test` (and CI) runs these; `go test -fuzz`
	// explores further.
	for ci, c := range categories {
		for from := range c.Units {
			for to := range c.Units {
				for _, v := range []float64{0, 1, 0.5, 100, -40, -273.15, -459.67, 1e-9, 1e9, 1e300} {
					f.Add(v, uint8(ci), uint8(from), uint8(to))
				}
			}
		}
	}
	f.Fuzz(func(t *testing.T, value float64, cat, from, to uint8) {
		c := categories[int(cat)%len(categories)]
		a := c.Units[int(from)%len(c.Units)]
		b := c.Units[int(to)%len(c.Units)]

		out, err := c.Convert(value, a.ID, b.ID)
		if err != nil {
			if !errors.Is(err, c.BelowZero) && !errors.Is(err, convert.ErrNotFinite) {
				t.Fatalf("Convert(%v, %s, %s): unexpected error %v", value, a.ID, b.ID, err)
			}
			return
		}
		back, err := c.Convert(out, b.ID, a.ID)
		if err != nil {
			t.Fatalf("round trip %v %s -> %v %s failed on the way back: %v", value, a.ID, out, b.ID, err)
		}
		// Relative tolerance; offset Conversions (temperature) also cancel
		// numbers around 273 or 460, which costs absolute precision.
		tol := 1e-9 * math.Abs(value)
		if a.Offset != 0 || b.Offset != 0 {
			tol += 1e-9 * 1000
		}
		if diff := math.Abs(back - value); diff > tol {
			t.Errorf("%v %s -> %v %s -> %v %s; off by %v (> %v)", value, a.ID, out, b.ID, back, a.ID, diff, tol)
		}
	})
}
