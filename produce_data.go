// This program reads the raw data from XP, get.tsv, and generates 3 files for
// train and for test. The files for train are:
//
// 1.  train_data.tsv: A matrix (r x f) with the features used to train the ML.
// 2.  train_labels.tsv: A vector (r) with the labels used to test the ML, one
//     label for each line in train_data.tsv.
// 3.  train_metadata.tsv: A matrix (r x m) containing, for each line in
//     train_labels.tsv:
//     a. The name of the fund.
//     b. The time of the data.
//
// The files for test are analogous.

package main

import (
	"flag"
	"fmt"
	// "github.com/montanaflynn/stats"
	"os"
	"xpfunds"
)

var testMonths = flag.Int("test_months", 1,
	"How many months to use for test. The rest of the months will be used for"+
		" trainning. The last months are reserved for testing.")

func main() {
	flag.Parse()
	funds := xpfunds.ReadFunds()
	duration := xpfunds.MaxDuration(funds)
	cdi := xpfunds.FundFromFile("cdi.tsv")
	ipca := xpfunds.FundFromFile("ipca.tsv")
	writeFiles(funds, cdi, ipca, *testMonths, duration, "train")
	writeFiles(funds, cdi, ipca, 0, *testMonths, "test")
}

// Goes through all time periods between end and start (both exclusive) and
// product the data and labels for this period.
func writeFiles(funds []*xpfunds.Fund, cdi, ipca *Fund, end, start int, name string) {
	data, err := os.Create(name + "_data.tsv")
	xpfunds.Check(err)
	labels, err := os.Create(name + "_labels.tsv")
	xpfunds.Check(err)
	metadata, err := os.Create(name + "_metadata.tsv")
	xpfunds.Check(err)
	for _, f := range funds {
		for time := end + 1; time < start+1; time++ {
			if time >= f.Duration() {
				break
			}
			// The annualized return from the beginning of the fund until now.
			fmt.Fprintf(data, "\t%v", f.Annual(time, f.Duration()))

			// Standard deviation.
			// std, err := stats.StandardDeviation(f.Monthly[time:f.Duration()])
			// xpfunds.Check(err)
			// fmt.Fprintf(data, "\t%v", std)

			// The CDI from the month.
			// fmt.Fprintf(data, "\t%v", cdi.Monthly[time])

			// The IPCA from the month.
			// fmt.Fprintf(data, "\t%v", ipca.Monthly[time])

			// The return from the last month
			// fmt.Fprintf(data, "\t%v", f.Period[time][0])

			// An indicator of which fund this data came from.
			// for _, f2 := range funds {
			// 	if f.Name == f2.Name {
			// 		fmt.Fprintf(data, "\t1")
			// 	} else {
			// 		fmt.Fprintf(data, "\t0")
			// 	}
			// }

			// The number of months from the beginning of the fund until now.
			// fmt.Fprintf(data, "\t%v", f.Duration() - time)

			fmt.Fprintf(data, "\n")
			fmt.Fprintf(labels, "%v\n", f.Annual(end, time))
			fmt.Fprintf(metadata, "%v\t%v\n", f.Name, time)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
