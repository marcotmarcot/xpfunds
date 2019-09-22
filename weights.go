package main

import (
	"fmt"
	"xpfunds"
	"xpfunds/simulate"
)

var (
	monthsToRead  = 0
	numFunds      = 10
	maxMinMonths  = 2
	stepMinMonths = 1
)

func main() {
	funds := xpfunds.ReadFunds()
	maxDuration := 0
	for _, f := range funds {
		if f.Duration() > maxDuration {
			maxDuration = f.Duration()
		}
	}
	c := make(chan bool)
	n := 0
	for monthsToRead := 0; monthsToRead <= 0; monthsToRead += 1 {
		for minMonths := 2; minMonths <= 2; minMonths += 1 {
			for ret := 0.675; ret <= 0.675; ret += 0.025 {
				n++
				go func(ret float64, monthsToRead, minMonths int) {
					for median := 0.8; median <= 0.8; median += 0.05 {
						for stdDev := 0.95; stdDev <= 0.95; stdDev += 0.025 {
							for nmr := -0.6; nmr <= -0.6; nmr += 0.025 {
								for gf := -0.4; gf <= -0.4; gf += 0.025 {
									for gfl := -0.65; gfl <= -0.65; gfl += 0.025 {
										weight := map[string]float64{
											"return":             ret,
											"median":             median,
											"stdDev":             stdDev,
											"negativeMonthRatio": nmr,
											"greatestFall":       gf,
											"greatestFallLen":    gfl,
										}
										s := &simulate.Weighted{numFunds, monthsToRead, minMonths, weight}
										p := simulate.MedianPerformance(funds, maxDuration-maxMinMonths, numFunds, s)
										fmt.Printf("%v\t%v\n", s.Name(), p)
										chosen := s.Choose(funds, 0)
										for _, f := range chosen {
											fmt.Println(f.Name)
										}
									}
								}
							}
						}
					}
					c <- true
				}(ret, monthsToRead, minMonths)
			}
		}
	}
	for i := 0; i < n; i++ {
		<-c
	}
}
