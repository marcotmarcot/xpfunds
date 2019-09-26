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
			for ret := 0.15625; ret <= 0.15625; ret += step {
				n++
				go func(ret float64, monthsToRead, minMonths int) {
					for median := 0.28125; median <= 0.28125; median += step {
						for stdDev := 0.8125; stdDev <= 0.8125; stdDev += step {
							for nmr := -0.46875; nmr <= -0.46875; nmr += step {
								for gf := 0.0; gf <= -0.0; gf += step {
									for gfl := -0.28125; gfl <= -0.28125; gfl += step {
										weight := []float64{
											ret,
											median,
											stdDev,
											nmr,
											gf,
											gfl,
											-1,
											-0.78125,
										}
										s := simulate.NewWeighted(numFunds, maxMinMonths, weight)
										p := simulate.MedianPerformance(funds, maxDuration, maxMinMonths, numFunds, s)
										fmt.Printf("%v\t%v\n", s.Name(), p)
										if print {
											chosen := s.Choose(funds, 0)
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
