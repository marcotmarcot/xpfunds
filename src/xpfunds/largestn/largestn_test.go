package largestn

import (
	"reflect"
	"testing"
	"xpfunds"
)

type element struct {
	ret float64
	f   string
}

func TestLargestN(t *testing.T) {
	tests := []struct {
		name     string
		elements []element
		want     []string
	}{{
		"asc",
		[]element{
			{1, "1"},
			{2, "2"},
			{3, "3"},
			{4, "4"},
		},
		[]string{"3", "4"},
	}, {
		"desc",
		[]element{
			{4, "4"},
			{3, "3"},
			{2, "2"},
			{1, "1"},
		},
		[]string{"3", "4"},
	}}
	for _, test := range tests {
		l := NewLargestN(2)
		for _, e := range test.elements {
			l.Add(xpfunds.NewFund(e.f, nil), e.ret)
		}
		var want []*xpfunds.Fund
		for _, n := range test.want {
			want = append(want, xpfunds.NewFund(n, nil))
		}
		if got := l.Funds; !reflect.DeepEqual(got, want) {
			t.Errorf("%v: got: %v, want: %v", test.name, got, want)
		}
	}
}
