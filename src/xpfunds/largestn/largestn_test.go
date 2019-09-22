package largestn

import (
	"reflect"
	"testing"
)

type element struct {
	ret float64
	i   int
}

func TestLargestN(t *testing.T) {
	tests := []struct {
		name     string
		elements []element
		want     []int
	}{{
		"asc",
		[]element{
			{1, 1},
			{2, 2},
			{3, 3},
			{4, 4},
		},
		[]int{3, 4},
	}, {
		"desc",
		[]element{
			{4, 1},
			{3, 2},
			{2, 3},
			{1, 4},
		},
		[]int{2, 1},
	}}
	for _, test := range tests {
		l := NewLargestN(2)
		for _, e := range test.elements {
			l.Add(e.i, e.ret)
		}
		if got := l.Indexes; !reflect.DeepEqual(got, test.want) {
			t.Errorf("%v: got: %v, want: %v", test.name, got, test.want)
		}
	}
}
