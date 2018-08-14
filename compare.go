package main

import (
	"fmt"
	"xpfunds"
)

func main() {
	cdis := xpfunds.ReadLines("cdi.tsv")
	funds := xpfunds.ReadFunds()
	duration := xpfunds.MaxDuration(funds)

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
		ratio_total += monthly_total / float64(monthly_count) / cdis[time]
	}
	fmt.Println(ratio_total / float64(duration))
}
