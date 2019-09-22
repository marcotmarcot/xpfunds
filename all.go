package main

import (
	"fmt"
	"math/rand"
	"time"
	"xpfunds"
	"xpfunds/simulate"
)

var (
	minMonths   = 6
	maxNumFunds = 10
)

func main() {
	rand.Seed(time.Now().UnixNano())
	funds := xpfunds.ReadFunds()
	maxDuration := 0
	for _, f := range funds {
		if f.Duration() > maxDuration {
			maxDuration = f.Duration()
		}
	}
	var strategies []simulate.Strategy
	for numFunds := 1; numFunds <= maxNumFunds; numFunds++ {
		strategies = append(strategies, random{numFunds})
		for monthsToRead := 0; monthsToRead <= minMonths; monthsToRead++ {
			for ignoreWithoutMonths := monthsToRead; ignoreWithoutMonths <= minMonths; ignoreWithoutMonths++ {
				for _, value := range []float64{-1, 1} {
					for _, field := range funds[0].Fields() {
						strategies = append(strategies, simulate.Weighted{numFunds, monthsToRead, ignoreWithoutMonths, map[string]float64{field: value}})
					}
				}
			}
		}
	}

	c := make(chan string)
	for _, s := range strategies {
		go func(s simulate.Strategy) {
			c <- fmt.Sprintf("%v\t%v\n", s.Name(), simulate.MedianPerformance(funds, maxDuration-minMonths, maxNumFunds, s))
		}(s)
	}
	for range strategies {
		fmt.Printf(<-c)
	}
}

type random struct {
	numFunds int
}

func (r random) Name() string {
	return fmt.Sprintf("random(%v)", r.numFunds)
}

func (r random) Choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund {
	rand.Shuffle(len(funds), func(i, j int) {
		funds[i], funds[j] = funds[j], funds[i]
	})
	return funds[:r.numFunds]
}
