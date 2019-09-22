package xpfunds

import (
	"math"
	"testing"
)

func TestFields(t *testing.T) {
	tests := []struct {
		field      string
		expected01 float64
		expected12 float64
		expected02 float64
	}{{
		"return",
		1.1,
		0.9,
		0.99,
	}, {
		"median",
		1.1,
		0.9,
		1,
	}, {
		"stdDev",
		0,
		0,
		0.1,
	}, {
		"negativeMonthRatio",
		0,
		1,
		0.5,
	}}
	for _, test := range tests {
		f := NewFund("", []float64{1.1, 0.9})
		if got, want := f.Field(test.field, 0, 1), test.expected01; !eq(got, want) {
			t.Errorf("%v: got: %v, want: %v", test.field, got, want)
		}
		if got, want := f.Field(test.field, 1, 2), test.expected12; !eq(got, want) {
			t.Errorf("%v: got: %v, want: %v", test.field, got, want)
		}
		if got, want := f.Field(test.field, 0, 2), test.expected02; !eq(got, want) {
			t.Errorf("%v: got: %v, want: %v", test.field, got, want)
		}
	}
}

func TestOptimum(t *testing.T) {
	o := NewOptimum([]*Fund{NewFund("", []float64{1.1, 1.2}), NewFund("", []float64{1.2, 1.1})})
	if got, want := o.Field("return", 0, 1), 1.2; !eq(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
	if got, want := o.Field("return", 1, 2), 1.2; !eq(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
	if got, want := o.Field("return", 0, 2), 1.32; !eq(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func eq(a, b float64) bool {
	return math.Abs(a-b) < 0.000001
}
