package largestn

import (
	"xpfunds"
	"xpfunds/binarysearch"
)

type LargestN struct {
	Funds   []*xpfunds.Fund
	n       int
	returns []float64
}

func NewLargestN(n int) *LargestN {
	return &LargestN{nil, n, nil}
}

func (l *LargestN) Add(f *xpfunds.Fund, r float64) {
	i := binarysearch.UpperBound(l.returns, r)
	if len(l.returns) == l.n && i == 0 {
		return
	}
	start := 1
	if len(l.returns) < l.n {
		start = 0
	}
	l.returns = append(l.returns[start:i], append([]float64{r}, l.returns[i:]...)...)
	l.Funds = append(l.Funds[start:i], append([]*xpfunds.Fund{f}, l.Funds[i:]...)...)
}
