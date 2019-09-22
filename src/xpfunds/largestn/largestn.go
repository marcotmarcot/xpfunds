package largestn

import (
	"xpfunds"
	"xpfunds/binarysearch"
)

type LargestN struct {
	Funds   []*xpfunds.Fund
	n       int
	reverse bool
	returns []float64
}

func NewLargestN(n int, reverse bool) *LargestN {
	return &LargestN{nil, n, reverse, nil}
}

func (l *LargestN) Add(f *xpfunds.Fund, r float64) {
	i := binarysearch.UpperBound(l.returns, r)
	if !l.reverse {
		l.noReverseAdd(f, r, i)
		return
	}
	l.reverseAdd(f, r, i)
}

func (l *LargestN) noReverseAdd(f *xpfunds.Fund, r float64, i int) {
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

func (l *LargestN) reverseAdd(f *xpfunds.Fund, r float64, i int) {
	if len(l.returns) == l.n && i == l.n {
		return
	}
	end := l.n - 1
	if len(l.returns) < l.n {
		end = len(l.returns)
	}
	l.returns = append(l.returns[:i], append([]float64{r}, l.returns[i:end]...)...)
	l.Funds = append(l.Funds[:i], append([]*xpfunds.Fund{f}, l.Funds[i:end]...)...)
}
