package main

import (
	"flag"
	"fmt"
	"xpfunds"
)

var months = flag.Int("months", -1,
	"How many months back in time to read. -1 for reading all time")

func main() {
	flag.Parse()
	cdi := xpfunds.FundFromFile("cdi.tsv")
	optimum := xpfunds.NewOptimum(xpfunds.ReadFunds())
	var duration int
	if *months == -1 {
		duration = len(optimum.Period)
	} else {
		duration = *months
	}

	ratio_total := 0.0
	for time := 0; time < duration; time++ {
		ratio_total += optimum.Period[time][0] / xpfunds.Annual(cdi.Monthly[time], 0, 1)
	}
	fmt.Println(ratio_total / float64(duration))
}
