package median

import (
	"testing"
)

func TestMedian(t *testing.T) {
	tests := []struct {
		s    []float64
		want float64
	}{{
		[]float64{1, 2, 3},
		2,
	}, {
		[]float64{1, 2, 3, 4},
		2.5,
	}}
	for _, test := range tests {
		if got := Median(test.s); got != test.want {
			t.Errorf("want: %v, got: %v", test.want, got)
		}
	}
}

func TestMedianFromSorted(t *testing.T) {
	tests := []struct {
		s    []float64
		want float64
	}{{
		[]float64{1, 2, 3},
		2,
	}, {
		[]float64{1, 2, 3, 4},
		2.5,
	}}
	for _, test := range tests {
		if got := MedianFromSorted(test.s); got != test.want {
			t.Errorf("want: %v, got: %v", test.want, got)
		}
	}
}
