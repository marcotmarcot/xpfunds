package simulate

import (
	"fmt"
	"xpfunds"
	"xpfunds/median"
)

func MedianPerformance(funds []*xpfunds.Fund, maxDuration, minMonths, maxNumFunds int, s Strategy) float64 {
	var perfs []float64

	for time := maxDuration - 1; time >= 1; time-- {
		var active []*xpfunds.Fund
		withMinMonths := 0
		for _, f := range funds {
			if f.Duration() >= time+1 {
				active = append(active, f)
			}
			if f.Duration() >= time+minMonths {
				withMinMonths++
			}
		}
		if len(active) < maxNumFunds+1 || withMinMonths == 0 {
			continue
		}
		perfs = append(perfs, performance(active, s, time))
	}
	return median.Median(perfs)
}

func performance(funds []*xpfunds.Fund, s Strategy, time int) float64 {
	chosenFunds := s.Choose(funds, time)
	total := 0.0
	for _, f := range chosenFunds {
		total += f.Return(0, time)
	}
	return total / float64(len(chosenFunds))
}

type Strategy interface {
	Name() string

	Choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund
}

type Weighted struct {
	monthsToRead        int
	ignoreWithoutMonths int
	weight              []float64
}

func NewWeighted(maxMonths int, weight []float64) *Weighted {
	return &Weighted{
		0,      // int(math.Round((weight[len(weight)-2] + 1) / 2 * float64(maxMonths))),
		6,      // int(math.Round((weight[len(weight)-1] + 1) / 2 * float64(maxMonths))),
		weight, // weight[:len(weight)-2],
	}
}

func (w *Weighted) Name() string {
	return fmt.Sprintf("Weighted(%v,%v,%v)", w.monthsToRead, w.ignoreWithoutMonths, w.weight)
}

func (w *Weighted) Choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund {
	featureCount := funds[0].FeatureCount()
	numFunds := len(w.weight) / featureCount
	chosen := make(map[*xpfunds.Fund]bool)
	for i := 0; i < numFunds; i++ {
		var bestFund *xpfunds.Fund
		bestValue := -999999.99
		for _, f := range funds {
			if chosen[f] || f.Duration()-end < w.monthsToRead+w.ignoreWithoutMonths {
				continue
			}
			start := end + w.monthsToRead
			if w.monthsToRead == 0 {
				start = f.Duration()
			}
			value := f.Weighted(w.weight[i*featureCount:(i+1)*featureCount], end, start)
			if value > bestValue {
				bestValue = value
				bestFund = f
			}
		}
		chosen[bestFund] = true
	}
	ret := make([]*xpfunds.Fund, numFunds)
	i := 0
	for f := range chosen {
		ret[i] = f
		i++
	}
	return ret
}

func (w *Weighted) FeatureCount() int {
	return 0
}
