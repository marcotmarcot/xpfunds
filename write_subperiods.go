package main

import (
	"fmt"
	"os"
	"path"
	"xpfunds"
)

func main() {
	xpfunds.Check(os.Mkdir("subperiods", 0775))
	cdi := xpfunds.FundFromFile("cdi.tsv")
	for _, f := range xpfunds.ReadFunds("") {
		meanSubPeriods := make([][]float64, len(f.Monthly))
		for end := range f.Monthly {
			meanSubPeriods[end] = make([]float64, len(f.Monthly)-end)
			var subPeriods []float64
			for start := end + 1; start <= len(f.Monthly); start++ {
				for endSubPeriod := end; endSubPeriod < start; endSubPeriod++ {
					subPeriods = append(subPeriods, f.Annual(endSubPeriod, start))
				}
				meanSubPeriods[end][start-1-end] = xpfunds.Mean(subPeriods) / cdi.Annual(end, start)
			}
		}
		data, err := os.Create(path.Join("subperiods", f.Name+".tsv"))
		xpfunds.Check(err)
		for _, endPeriod := range meanSubPeriods {
			for _, mean := range endPeriod {
				fmt.Fprintf(data, "%v\t", mean)
			}
			fmt.Fprintf(data, "\n")
		}
	}
}
