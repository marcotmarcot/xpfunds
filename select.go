package main

import (
	"fmt"
	"math"
	"time"
	"xpfunds"
	"xpfunds/simulate"
)

var (
	funds       []*xpfunds.Fund
	maxDuration int
	maxMonths   = 0
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
	for i := 0; true; i++ {
		start := time.Now()
		best, perf := bestInRegion(point, step)
		end := time.Now()
		fmt.Printf("%v\t%v\t%v\t%v\n", i, best, perf, end.Sub(start).String())
		step /= 2
		if !samePoint(point, best) {
			point = nextPoint(point, best)
		}
	}
}

func bestInRegion(point []float64, step float64) ([]float64, float64) {
	return runBestInRegion(nil, point, step, 0)
}

func runBestInRegion(picked, toPick []float64, step float64, parallel int) ([]float64, float64) {
	if len(toPick) == 0 {
		perf := simulate.MedianPerformance(funds, maxDuration, numFunds, simulate.NewWeighted(maxMonths, picked))
		fmt.Println(picked, perf)
		return picked, perf
	}
	var bestPicked []float64
	best := -999999.99
	i := len(toPick) - 1
	if parallel > 0 {
		c := make(chan *result)
		for d := toPick[i] - step; d <= toPick[i]+step; d += step {
			go func(d float64) {
				picked, perf := runBestInRegion(append(picked, d), toPick[:i], step, parallel-1)
				c <- &result{picked, perf}
			}(d)
		}
		for d := toPick[i] - step; d <= toPick[i]+step; d += step {
			r := <-c
			if r.perf > best {
				best = r.perf
				bestPicked = r.picked
			}
		}
		return bestPicked, best
	}
	for d := toPick[i] - step; d <= toPick[i]+step; d += step {
		picked, perf := runBestInRegion(append(picked, d), toPick[:i], step, parallel-1)
		if perf > best {
			best = perf
			bestPicked = picked
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

func samePoint(a, b []float64) bool {
	for i := range a {
		if math.Abs(a[i]-b[i]) > 0.000001 {
			return false
		}
	}
	return true
}
