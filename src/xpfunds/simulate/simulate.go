package simulate

import (
	"fmt"
	"log"
	"math"
	"xpfunds"
	"xpfunds/largestn"
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
	numFunds            int
	monthsToRead        int
	ignoreWithoutMonths int
	weight              []float64
}

func NewWeighted(numFunds, maxMonths int, weight []float64) *Weighted {
	return &Weighted{
		numFunds,
		int(math.Round((weight[len(weight)-2] + 1) / 2 * float64(maxMonths))),
		int(math.Round((weight[len(weight)-1] + 1) / 2 * float64(maxMonths))),
		weight[:len(weight)-2],
	}
}

func (w *Weighted) Name() string {
	return fmt.Sprintf("Weighted(%v,%v,%v,%v)", w.numFunds, w.monthsToRead, w.ignoreWithoutMonths, w.weight)
}

func (w *Weighted) Choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund {
	l := largestn.NewLargestN(w.numFunds)
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
	if len(chosen) == 0 {
		for _, f := range funds {
			fmt.Println(f.Duration() - end)
		}
		log.Fatal("len(funds)=", len(funds), " w=", w.Name())
	}
	return chosen
}

func (w *Weighted) FeatureCount() int {
	return 2
}
