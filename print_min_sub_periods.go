package main

import (
	"fmt"
	"xpfunds"
)

func main() {
	for _, f := range xpfunds.ReadFunds() {
		fmt.Printf("%v\t%v\n", f.Name, f.MeanSubPeriodsReturn(0, len(f.Monthly)))
	}
}
