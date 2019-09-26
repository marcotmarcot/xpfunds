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
	maxMonths   = 3
	numFunds    = 10
)

func main() {
	funds = xpfunds.ReadFunds()
	for _, f := range funds {
		if f.Duration() > maxDuration {
			maxDuration = f.Duration()
		}
	}
	point := make([]float64, (funds[0].FeatureCount()+(&simulate.Weighted{}).FeatureCount())*numFunds)
	fmt.Println(point, simulate.MedianPerformance(funds, maxDuration, maxMonths*2, numFunds, simulate.NewWeighted(maxMonths, point)))
	step := 1.0
	for i := 0; true; i++ {
		start := time.Now()
		best, perf := bestInRegion(point, step)
		end := time.Now()
		fmt.Printf("%v\t%v\t%v\t%v\t%v\n", i, best, perf, end.Sub(start).String(), step)
		point = nextPoint(point, best)
		step /= 2
	}
}

func bestInRegion(point []float64, step float64) ([]float64, float64) {
	newPoint := make([]float64, len(point))
	for i, p := range point {
		newPoint[i] = p
	}
	bestPerf := simulate.MedianPerformance(funds, maxDuration, maxMonths*2, numFunds, simulate.NewWeighted(maxMonths, newPoint))
	for i := 0; i < len(newPoint); i++ {
		newPoint[i] -= step
		left := simulate.MedianPerformance(funds, maxDuration, maxMonths*2, numFunds, simulate.NewWeighted(maxMonths, newPoint))
		newPoint[i] += step * 2
		right := simulate.MedianPerformance(funds, maxDuration, maxMonths*2, numFunds, simulate.NewWeighted(maxMonths, newPoint))
		// No change.
		if bestPerf > left && bestPerf > right {
			newPoint[i] -= step
			continue
		}
		if right >= left {
			bestPerf = right
			// NewPoint is already at right.
			continue
		}
		bestPerf = left
		newPoint[i] -= step * 2
	}
	return newPoint, bestPerf
}

func nextPoint(orig []float64, best []float64) []float64 {
	new := make([]float64, len(orig))
	for i := range orig {
		new[i] = (orig[i] + best[i]) / 2
	}
	return new
}
