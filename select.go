package main

import (
	"fmt"
	"time"
	"xpfunds"
	"xpfunds/simulate"
)

var (
	funds       []*xpfunds.Fund
	maxDuration int
	maxMonths   = 6
	numFunds    = 10
)

func main() {
	funds = xpfunds.ReadFunds()
	for _, f := range funds {
		if f.Duration() > maxDuration {
			maxDuration = f.Duration()
		}
	}
	point := make([]float64, funds[0].FeatureCount()+(&simulate.Weighted{}).FeatureCount())
	step := 1.0
	bestPerf := 0.0
	for i := 0; true; i++ {
		start := time.Now()
		best, perf := bestInRegion(point, step)
		if perf-bestPerf > 0.000001 {
			bestPerf = perf
			point = nextPoint(point, best)
		} else {
			best = point
		}
		end := time.Now()
		fmt.Printf("%v\t%v\t%v\t%v\t%v\n", i, best, perf, end.Sub(start).String(), step)
		step /= 2
	}
}

func bestInRegion(point []float64, step float64) ([]float64, float64) {
	return runBestInRegion(nil, point, step, 2)
}

func runBestInRegion(picked, toPick []float64, step float64, parallel int) ([]float64, float64) {
	if len(toPick) == 0 {
		perf := simulate.MedianPerformance(funds, maxDuration, maxMonths*2, numFunds, simulate.NewWeighted(numFunds, maxMonths, picked))
		return picked, perf
	}
	bestPicked := make([]float64, len(picked)+len(toPick))
	best := -999999.99
	if parallel > 0 {
		c := make(chan *result)
		for d := toPick[0] - step; d <= toPick[0]+step; d += step {
			go func(d float64) {
				subBest, perf := runBestInRegion(append(picked, d), toPick[1:], step, parallel-1)
				c <- &result{subBest, perf}
			}(d)
		}
		for d := toPick[0] - step; d <= toPick[0]+step; d += step {
			r := <-c
			if r.perf > best {
				best = r.perf
				for i, p := range r.picked {
					bestPicked[i] = p
				}
			}
		}
		return bestPicked, best
	}
	for d := toPick[0] - step; d <= toPick[0]+step; d += step {
		subBest, perf := runBestInRegion(append(picked, d), toPick[1:], step, parallel-1)
		if perf > best {
			best = perf
			for i, p := range subBest {
				bestPicked[i] = p
			}
		}
	}
	return bestPicked, best
}

type result struct {
	picked []float64
	perf   float64
}

func nextPoint(orig []float64, best []float64) []float64 {
	new := make([]float64, len(orig))
	for i := range orig {
		new[i] = (orig[i] + best[i]) / 2
	}
	return new
}
