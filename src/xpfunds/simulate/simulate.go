package simulate

import (
	"fmt"
	"sort"
	"xpfunds"
	"xpfunds/largestn"
)

func MedianPerformance(funds []*xpfunds.Fund, maxDuration, maxNumFunds int, s Strategy) float64 {
	var perfs []float64

	for time := maxDuration - 2; time >= 1; time-- {
		var active []*xpfunds.Fund
		for _, f := range funds {
			if f.Duration() >= time+1 {
				active = append(active, f)
			}
		}
		if len(active) < maxNumFunds+1 {
			continue
		}
		perfs = append(perfs, performance(active, s, time))
	}
	return median(perfs)
}

func median(s []float64) float64 {
	if len(s) == 0 {
		return -1
	}
	if len(s) == 1 {
		return s[0]
	}
	sort.Float64s(s)
	m := len(s) / 2
	if len(s)%2 == 0 {
		return s[m]
	}
	return (s[m] + s[m+1]) / 2
}

func performance(funds []*xpfunds.Fund, s Strategy, time int) float64 {
	chosenFunds := s.Choose(funds, time)
	total := 0.0
	for _, f := range chosenFunds {
		total += f.Weighted(map[string]float64{"return": 1}, 0, time)
	}
	return total / float64(len(chosenFunds))
}

type Strategy interface {
	Name() string

	Choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund
}

type Weighted struct {
	NumFunds            int
	MonthsToRead        int
	IgnoreWithoutMonths int
	Weight              map[string]float64
}

func (w Weighted) Name() string {
	return fmt.Sprintf("Weighted(%v,%v,%v,%v)", w.NumFunds, w.MonthsToRead, w.IgnoreWithoutMonths, w.Weight)
}

func (w Weighted) Choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund {
	l := largestn.NewLargestN(w.NumFunds)
	for i, f := range funds {
		if f.Duration()-end < w.IgnoreWithoutMonths {
			continue
		}
		start := end + w.MonthsToRead
		if w.MonthsToRead == 0 {
			start = f.Duration()
		}
		l.Add(i, f.Weighted(w.Weight, end, start))
	}
	chosen := make([]*xpfunds.Fund, len(l.Indexes))
	for i, index := range l.Indexes {
		chosen[i] = funds[index]
	}
	return chosen
}
