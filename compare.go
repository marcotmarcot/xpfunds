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
	funds := xpfunds.ReadFunds()
	var duration int
	if *months == -1 {
		duration = xpfunds.MaxDuration(funds)
	} else {
		duration = *months
	}

	ratio_total := 0.0
	for time := 0; time < duration; time++ {
		monthly_total := 0.0
		monthly_count := 0
		for _, f := range funds {
			if time >= f.Duration() {
				continue
			}
			monthly_total += f.Monthly[time]
			monthly_count++
		}
		ratio_total += monthly_total / float64(monthly_count) / cdi.Monthly[time]
	}
	fmt.Println(ratio_total / float64(duration))
}
