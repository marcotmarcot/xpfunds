package main

import (
	"fmt"
	"xpfunds"
	"xpfunds/largestn"
	"xpfunds/simulate"
	"xpfunds/weight"
)

var (
	monthsToRead = 0
	numFunds     = 10
	generations  = 100
	size         = 30
)

func main() {
	funds := xpfunds.ReadFunds()
	maxDuration := 0
	for _, f := range funds {
		if f.Duration() > maxDuration {
			maxDuration = f.Duration()
		}
	}
	weights := weight.NewWeights(funds[0].Fields(), size)
	for generation := 0; generation < generations; generation++ {
		l := largestn.NewLargestN(2)
		c := make(chan bool)
		for i, weight := range weights {
			go func() {
				s := &simulate.Weighted{numFunds, monthsToRead, monthsToRead, weight}
				p := simulate.MedianPerformance(funds, maxDuration-monthsToRead, numFunds, s)
				fmt.Printf("%v\t%v\n", s.Name(), p)
				l.Add(i, p)
				c <- true
			}()
		}
		for range weights {
			<-c
		}
		weights = weight.Reproduce(weights[l.Indexes[0]], weights[l.Indexes[1]], size)
	}
}
