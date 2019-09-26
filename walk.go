package main

import (
	"fmt"
	"math/rand"
	"time"
	"xpfunds"
	"xpfunds/simulate"
)

var (
	funds       []*xpfunds.Fund
	maxDuration int
	maxMonths   = 60
	numFunds    = 1
)

func main() {
	rand.Seed(time.Now().UnixNano())
	funds = xpfunds.ReadFunds()
	for _, f := range funds {
		if f.Duration() > maxDuration {
			maxDuration = f.Duration()
		}
	}
	point := make([]float64, (funds[0].FeatureCount()+(&simulate.Weighted{}).FeatureCount())*numFunds)
	for i := range point {
		point[i] = rand.Float64()*2 - 1
	}
	step := 1.0
	for i := 0; true; i++ {
		start := time.Now()
		best, perf := bestInRegion(point)
		end := time.Now()
		fmt.Printf("%v\t%v\t%v\t%v\t%v\n", i, perf, end.Sub(start).String(), best, step)
		point = best
	}
}

func bestInRegion(point []float64) ([]float64, float64) {
	newPoint := make([]float64, len(point))
	for i, p := range point {
		newPoint[i] = p
	}
	bestPerf := simulate.MedianPerformance(funds, maxDuration, maxMonths*2, numFunds, simulate.NewWeighted(maxMonths, newPoint))
	for i := 0; i < len(newPoint); i++ {
		step := rand.Float64()*2 - 1
		if newPoint[i]+step <= -1 || newPoint[i]+step >= 1 {
			continue
		}
		newPoint[i] += step
		perf := simulate.MedianPerformance(funds, maxDuration, maxMonths*2, numFunds, simulate.NewWeighted(maxMonths, newPoint))
		if perf > bestPerf {
			bestPerf = perf
			continue
		}
		newPoint[i] -= step
	}
	return newPoint, bestPerf
}
