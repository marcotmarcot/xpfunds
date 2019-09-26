package simulate

import (
	"fmt"
	"log"
	"math"
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
			if f.Duration() >= time+minMonths+1 {
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
	maxMonths int
	weight    []float64
}

func NewWeighted(maxMonths int, weight []float64) *Weighted {
	return &Weighted{
		maxMonths,
		weight,
	}
}

func (w *Weighted) Name() string {
	return fmt.Sprintf("Weighted(%v,%v)", w.maxMonths, w.weight)
}

func (w *Weighted) Choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund {
	fundFeatureCount := funds[0].FeatureCount()
	featureCount := fundFeatureCount + w.FeatureCount()
	numFunds := len(w.weight) / featureCount
	chosen := make(map[*xpfunds.Fund]bool)
	for i := 0; i < numFunds; i++ {
		var bestFund *xpfunds.Fund
		bestValue := -999999.99
		for _, f := range funds {
			monthsToReadWeight := w.weight[i*featureCount+fundFeatureCount]
			monthsToRead := int(math.Round((monthsToReadWeight + 1) / 2 * float64(w.maxMonths)))
			ignoreWithoutMonthsWeight := w.weight[i*featureCount+fundFeatureCount+1]
			ignoreWithoutMonths := int(math.Round((ignoreWithoutMonthsWeight + 1) / 2 * float64(w.maxMonths)))
			if chosen[f] || f.Duration()-end < monthsToRead+ignoreWithoutMonths {
				continue
			}
			start := end + monthsToRead
			if monthsToRead == 0 {
				start = f.Duration()
			}
			value := f.Weighted(w.weight[i*featureCount:i*featureCount+fundFeatureCount], end, start)
			if value > bestValue {
				bestValue = value
				bestFund = f
			}
		}
		if bestFund == nil {
			break
		}
		chosen[bestFund] = true
	}
	if len(chosen) == 0 {
		for _, f := range funds {
			log.Print(f.Duration() - end)
		}
		log.Fatal("len(funds)=", len(funds), " w=", w.Name())
	}
	ret := make([]*xpfunds.Fund, len(chosen))
	i := 0
	for f := range chosen {
		ret[i] = f
		i++
	}
	return ret
}

func (w *Weighted) FeatureCount() int {
	return 2
}
