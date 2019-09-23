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
	step          = 1.0
	print         = true
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
			for ret := 0.675; ret <= 0.7; ret += step {
				n++
				go func(ret float64, monthsToRead, minMonths int) {
					for median := 0.775; median <= 0.825; median += step {
						for stdDev := 0.95; stdDev <= 0.975; stdDev += step {
							for nmr := -0.575; nmr <= -0.575; nmr += step {
								for gf := -0.425; gf <= -0.375; gf += step {
									for gfl := -0.625; gfl <= -0.625; gfl += step {
										weight := []float64{
											ret,
											median,
											stdDev,
											nmr,
											gf,
											gfl,
											-1,
											1,
										}
										s := simulate.NewWeighted(maxMinMonths, weight)
										p := simulate.MedianPerformance(funds, maxDuration-maxMinMonths, numFunds, s)
										fmt.Printf("%v\t%v\n", s.Name(), p)
										if print {
											chosen := s.Choose(funds, 20, 0)
											for _, f := range chosen {
												fmt.Println(f.Print())
											}
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
