package xpfunds

import (
	"math"
	"testing"
)

func TestFields(t *testing.T) {
	tests := []struct {
		field      string
		index      int
		expected01 float64
		expected12 float64
		expected02 float64
	}{{
		"return",
		0,
		1.1,
		0.9,
		0.99,
	}, {
		"median",
		1,
		1.1,
		0.9,
		1,
	}, {
		"stdDev",
		2,
		0,
		0,
		0.1,
	}, {
		"negativeMonthRatio",
		3,
		0,
		1,
		0.5,
	}, {
		"greatestFall",
		4,
		1,
		0.9,
		0.9,
	}, {
		"greatestFallLen",
		5,
		0,
		1,
		1,
	}}
	for _, test := range tests {
		f := NewFund([]float64{1.1, 0.9})
		if got, want := f.features[test.index][0][0], test.expected01; !eq(got, want) {
			t.Errorf("%v: got: %v, want: %v", test.field, got, want)
		}
		if got, want := f.features[test.index][1][0], test.expected12; !eq(got, want) {
			t.Errorf("%v: got: %v, want: %v", test.field, got, want)
		}
		if got, want := f.features[test.index][0][1], test.expected02; !eq(got, want) {
			t.Errorf("%v: got: %v, want: %v", test.field, got, want)
		}
	}
}

func TestWeighted(t *testing.T) {
	funds := []*Fund{NewFund([]float64{1.1, 0.9}), NewFund([]float64{1.1, 1.2})}
	SetRatio(funds)
	weights := []float64{1, 2}
	if got, want := funds[0].Weighted(weights, 0, 1), 3.0; !eq(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
	if got, want := funds[0].Weighted(weights, 1, 2), 2.25; !eq(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
	if got, want := funds[0].Weighted(weights, 0, 2), 2.489130434782609; !eq(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func eq(a, b float64) bool {
	return math.Abs(a-b) < 0.000001
}
