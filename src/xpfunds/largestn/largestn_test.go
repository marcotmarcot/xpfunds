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
		reverse  bool
		want     []string
	}{{
		"noReverseAsc",
		[]element{
			{1, "1"},
			{2, "2"},
			{3, "3"},
			{4, "4"},
		},
		false,
		[]string{"3", "4"},
	}, {
		"noReverseDesc",
		[]element{
			{4, "4"},
			{3, "3"},
			{2, "2"},
			{1, "1"},
		},
		false,
		[]string{"3", "4"},
	}, {
		"reverse",
		[]element{
			{1, "1"},
			{2, "2"},
			{3, "3"},
			{4, "4"},
		},
		true,
		[]string{"1", "2"},
	}, {
		"reverseDesc",
		[]element{
			{4, "4"},
			{3, "3"},
			{2, "2"},
			{1, "1"},
		},
		true,
		[]string{"1", "2"},
	}}
	for _, test := range tests {
		l := NewLargestN(2, test.reverse)
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
