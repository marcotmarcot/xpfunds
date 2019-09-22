package largestn

import (
	"xpfunds/binarysearch"
)

type LargestN struct {
	Indexes []int
	n       int
	returns []float64
}

func NewLargestN(n int) *LargestN {
	return &LargestN{nil, n, nil}
}

func (l *LargestN) Add(index int, r float64) {
	i := binarysearch.UpperBound(l.returns, r)
	if len(l.returns) == l.n && i == 0 {
		return
	}
	start := 1
	if len(l.returns) < l.n {
		start = 0
	}
	l.returns = append(l.returns[start:i], append([]float64{r}, l.returns[i:]...)...)
	l.Indexes = append(l.Indexes[start:i], append([]int{index}, l.Indexes[i:]...)...)
}
