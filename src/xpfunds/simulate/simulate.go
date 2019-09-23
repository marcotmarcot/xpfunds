package simulate

import (
	"fmt"
	"math"
	"xpfunds"
	"xpfunds/largestn"
	"xpfunds/median"
)

func MedianPerformance(funds []*xpfunds.Fund, maxDuration, numFunds int, s Strategy) float64 {
	var perfs []float64

	for time := maxDuration - 1; time >= 1; time-- {
		var active []*xpfunds.Fund
		for _, f := range funds {
			if f.Duration() >= time+1 {
				active = append(active, f)
			}
		}
		if len(active) < numFunds+1 {
			continue
		}
		perf, ok := performance(active, numFunds, s, time)
		if ok {
			perfs = append(perfs, perf)
		}
	}
	return median.Median(perfs)
}

func performance(funds []*xpfunds.Fund, numFunds int, s Strategy, time int) (float64, bool) {
	chosenFunds := s.Choose(funds, numFunds, time)
	if len(chosenFunds) < numFunds {
		return 0, false
	}
	total := 0.0
	for _, f := range chosenFunds {
		total += f.Return(0, time)
	}
	return total / float64(len(chosenFunds)), true
}

type Strategy interface {
	Name() string

	Choose(funds []*xpfunds.Fund, numFunds, end int) []*xpfunds.Fund
}

type Weighted struct {
	monthsToRead        int
	ignoreWithoutMonths int
	weight              []float64
}

func NewWeighted(maxMonths int, weight []float64) *Weighted {
	return &Weighted{
		int(math.Round((weight[len(weight)-2] + 1) / 2 * float64(maxMonths))),
		int(math.Round((weight[len(weight)-1] + 1) / 2 * float64(maxMonths))),
		weight[:len(weight)-2],
	}
}

func (w *Weighted) Name() string {
	return fmt.Sprintf("Weighted(%v,%v,%v)", w.monthsToRead, w.ignoreWithoutMonths, w.weight)
}

func (w *Weighted) Choose(funds []*xpfunds.Fund, numFunds, end int) []*xpfunds.Fund {
	l := largestn.NewLargestN(numFunds)
	for i, f := range funds {
		if f.Duration()-end < w.monthsToRead+w.ignoreWithoutMonths {
			continue
		}
		start := end + w.monthsToRead
		if w.monthsToRead == 0 {
			start = f.Duration()
		}
		l.Add(i, f.Weighted(w.weight, end, start))
	}
	chosen := make([]*xpfunds.Fund, len(l.Indexes))
	for i, index := range l.Indexes {
		chosen[i] = funds[index]
	}
	return chosen
}

func (w *Weighted) FeatureCount() int {
	return 2
}
