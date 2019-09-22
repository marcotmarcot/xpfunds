package binarysearch

import (
	"reflect"
	"testing"
)

func TestInsertInSorted(t *testing.T) {
	tests := []struct {
		s    []float64
		e    float64
		want []float64
	}{{
		nil,
		1,
		[]float64{1},
	}, {
		[]float64{1},
		0,
		[]float64{0, 1},
	}, {
		[]float64{1},
		2,
		[]float64{1, 2},
	}, {
		[]float64{1, 3},
		2,
		[]float64{1, 2, 3},
	}}
	for _, test := range tests {
		if got := InsertInSorted(test.s, test.e); !reflect.DeepEqual(got, test.want) {
			t.Errorf("want: %v, got: %v", test.want, got)
		}
	}
}

func TestUpperBound(t *testing.T) {
	tests := []struct {
		s    []float64
		e    float64
		want int
	}{{
		nil,
		1,
		0,
	}, {
		[]float64{1},
		0,
		0,
	}, {
		[]float64{1},
		2,
		1,
	}, {
		[]float64{1, 3},
		2,
		1,
	}}
	for _, test := range tests {
		if got := UpperBound(test.s, test.e); got != test.want {
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
