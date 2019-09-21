package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
	"xpfunds"
)

var minMonths = 12

func main() {
	rand.Seed(time.Now().UnixNano())
	funds := xpfunds.ReadFunds()
	optimum := xpfunds.NewOptimum(funds)
	strategies := []strategy{random{}}
	for monthsToRead := 0; monthsToRead <= minMonths; monthsToRead++ {
		for ignoreWithoutMonths := monthsToRead + 1; ignoreWithoutMonths <= minMonths; ignoreWithoutMonths++ {
			for _, reverse := range []bool{false, true} {
				strategies = append(strategies,
					bestInPeriod{monthsToRead, ignoreWithoutMonths, reverse, "return", (*xpfunds.Fund).Return},
					bestInPeriod{monthsToRead, ignoreWithoutMonths, reverse, "median", (*xpfunds.Fund).MedianReturn})
			}
		}
	}

	for _, s := range strategies {
		fmt.Printf("%v\t%v\n", s.name(), medianLoss(funds, optimum, s))
	}
}

func medianLoss(funds []*xpfunds.Fund, optimum *xpfunds.Fund, s strategy) float64 {
	var losses []float64

	for time := optimum.Duration() - minMonths - 2; time >= 1; time-- {
		var active []*xpfunds.Fund
		for _, f := range funds {
			if f.Duration() >= time+1 {
				active = append(active, f)
			}
		}
		if len(active) < 2 {
			continue
		}
		losses = append(losses, performance(active, s, time)/optimum.Return(0, time))
	}
	return median(losses)
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
	best := s.choose(funds, time)
	return best.Return(0, time)
}

type strategy interface {
	name() string

	// Given a list of valid index, return them in the order they should be
	// picked. Only consider data in funds up to end (inclusive).
	choose(funds []*xpfunds.Fund, end int) *xpfunds.Fund
}

type random struct{}

func (r random) name() string {
	return "random"
}

func (r random) choose(funds []*xpfunds.Fund, end int) *xpfunds.Fund {
	return funds[rand.Int()%len(funds)]
}

type bestInPeriod struct {
	monthsToRead        int
	ignoreWithoutMonths int
	reverse             bool
	retName             string
	ret                 func(f *xpfunds.Fund, end, start int) float64
}

func (b bestInPeriod) name() string {
	return fmt.Sprintf("bestInPeriod(%v,%v,%v,%v)", b.monthsToRead, b.ignoreWithoutMonths, b.reverse, b.retName)
}

func (b bestInPeriod) choose(funds []*xpfunds.Fund, end int) *xpfunds.Fund {
	bestReturn := -999999.99
	if b.reverse {
		bestReturn = 999999.99
	}
	var bestFund *xpfunds.Fund
	for _, f := range funds {
		if f.Duration()-end < b.ignoreWithoutMonths {
			continue
		}
		start := end + b.monthsToRead
		if b.monthsToRead == 0 {
			start = f.Duration()
		}
		r := b.ret(f, end, start)
		if (!b.reverse && r > bestReturn) || (b.reverse && r < bestReturn) {
			bestReturn = r
			bestFund = f
		}
	}
	return bestFund
}
