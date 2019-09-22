package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
	"xpfunds"
	"xpfunds/largestn"
)

var (
	minMonths   = 6
	maxNumFunds = 10
)

func main() {
	rand.Seed(time.Now().UnixNano())
	funds := xpfunds.ReadFunds()
	optimum := xpfunds.NewOptimum(funds)
	var strategies []strategy
	for numFunds := 1; numFunds <= maxNumFunds; numFunds++ {
		strategies = append(strategies, random{numFunds})
		for monthsToRead := 0; monthsToRead <= minMonths; monthsToRead++ {
			for ignoreWithoutMonths := monthsToRead; ignoreWithoutMonths <= minMonths; ignoreWithoutMonths++ {
				for _, reverse := range []bool{false, true} {
					for _, field := range optimum.Fields() {
						strategies = append(strategies, bestInPeriod{numFunds, monthsToRead, ignoreWithoutMonths, reverse, field})
					}
				}
			}
		}
	}

	c := make(chan string)
	for _, s := range strategies {
		go medianLoss(funds, optimum, s, c)
	}
	for range strategies {
		fmt.Printf(<-c)
	}
}

func medianLoss(funds []*xpfunds.Fund, optimum *xpfunds.Fund, s strategy, c chan string) {
	var losses []float64

	for time := optimum.Duration() - minMonths - 2; time >= 1; time-- {
		var active []*xpfunds.Fund
		for _, f := range funds {
			if f.Duration() >= time+1 {
				active = append(active, f)
			}
		}
		if len(active) < maxNumFunds+1 {
			continue
		}
		losses = append(losses, performance(active, s, time)/optimum.Field("return", 0, time))
	}
	c <- fmt.Sprintf("%v\t%v\n", s.name(), median(losses))
}

func median(s []float64) float64 {
	if len(s) == 0 {
		return -1
	}
	if len(s) == 1 {
		return s[0]
	}
	sort.Float64s(s)
	m := len(s) / 2
	if len(s)%2 == 0 {
		return s[m]
	}
	return (s[m] + s[m+1]) / 2
}

func performance(funds []*xpfunds.Fund, s strategy, time int) float64 {
	chosenFunds := s.choose(funds, time)
	total := 0.0
	for _, f := range chosenFunds {
		total += f.Field("return", 0, time)
	}
	return total / float64(len(chosenFunds))
}

type strategy interface {
	name() string

	choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund
}

type random struct {
	numFunds int
}

func (r random) name() string {
	return fmt.Sprintf("random(%v)", r.numFunds)
}

func (r random) choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund {
	rand.Shuffle(len(funds), func(i, j int) {
		funds[i], funds[j] = funds[j], funds[i]
	})
	return funds[:r.numFunds]
}

type bestInPeriod struct {
	numFunds            int
	monthsToRead        int
	ignoreWithoutMonths int
	reverse             bool
	field               string
}

func (b bestInPeriod) name() string {
	return fmt.Sprintf("bestInPeriod(%v,%v,%v,%v,%v)", b.numFunds, b.monthsToRead, b.ignoreWithoutMonths, b.reverse, b.field)
}

func (b bestInPeriod) choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund {
	l := largestn.NewLargestN(b.numFunds, b.reverse)
	for _, f := range funds {
		if f.Duration()-end < b.ignoreWithoutMonths {
			continue
		}
		start := end + b.monthsToRead
		if b.monthsToRead == 0 {
			start = f.Duration()
		}
		l.Add(f, f.Field(b.field, end, start))
	}
	return l.Funds
}
