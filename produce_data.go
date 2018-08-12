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
	"fmt"
	"os"
	"xpfunds"
)

func main() {
	funds := xpfunds.ReadFunds()
	duration := xpfunds.MaxDuration(funds)
	writeFiles(funds, duration/2, duration, "train")
	writeFiles(funds, 0, duration/2+1, "test")
}

// Goes through all time periods between end and start (both exclusive) and
// product the data and labels for this period.
func writeFiles(funds []*xpfunds.Fund, end, start int, name string) {
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
			fmt.Fprintf(data, "%v", f.Annual(time, f.Duration()))
			// The number of months from the beginning of the fund until now.
			// fmt.Fprintf(data, "\t%v", f.Duration() - time)
			fmt.Fprintf(data, "\n")
			fmt.Fprintf(labels, "%v\n", f.Annual(end, time))
			fmt.Fprintf(metadata, "%v\t%v\n", f.Name, time)
		}
	}
}
