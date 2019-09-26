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
	maxMonths   = 3
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
		left := 0.0
		leftStep := rand.NormFloat64()
		if newPoint[i]-leftStep >= -1 {
			newPoint[i] -= leftStep
			left = simulate.MedianPerformance(funds, maxDuration, maxMonths*2, numFunds, simulate.NewWeighted(maxMonths, newPoint))
			newPoint[i] += leftStep
		}
		right := 0.0
		rightStep := rand.NormFloat64()
		if newPoint[i]+rightStep <= 1 {
			newPoint[i] += rightStep
			right = simulate.MedianPerformance(funds, maxDuration, maxMonths*2, numFunds, simulate.NewWeighted(maxMonths, newPoint))
			newPoint[i] -= rightStep
		}
		// No change.
		if bestPerf > left && bestPerf > right {
			continue
		}
		if left > right {
			newPoint[i] -= leftStep
			bestPerf = left
			continue
		}
		newPoint[i] += rightStep
		bestPerf = right
	}
	return newPoint, bestPerf
}
