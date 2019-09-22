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
	maxDuration := 0
	for _, f := range funds {
		if f.Duration() > maxDuration {
			maxDuration = f.Duration()
		}
	}
	var strategies []strategy
	for numFunds := 1; numFunds <= maxNumFunds; numFunds++ {
		strategies = append(strategies, random{numFunds})
		for monthsToRead := 0; monthsToRead <= minMonths; monthsToRead++ {
			for ignoreWithoutMonths := monthsToRead; ignoreWithoutMonths <= minMonths; ignoreWithoutMonths++ {
				for _, value := range []float64{-1, 1} {
					for _, field := range funds[0].Fields() {
						strategies = append(strategies, bestInPeriod{numFunds, monthsToRead, ignoreWithoutMonths, map[string]float64{field: value}})
					}
				}
			}
		}
	}

	c := make(chan string)
	for _, s := range strategies {
		go medianPerformance(funds, maxDuration, s, c)
	}
	for range strategies {
		fmt.Printf(<-c)
	}
}

func medianPerformance(funds []*xpfunds.Fund, maxDuration int, s strategy, c chan string) {
	var perfs []float64

	for time := maxDuration - minMonths - 2; time >= 1; time-- {
		var active []*xpfunds.Fund
		for _, f := range funds {
			if f.Duration() >= time+1 {
				active = append(active, f)
			}
		}
		if len(active) < maxNumFunds+1 {
			continue
		}
		perfs = append(perfs, performance(active, s, time))
	}
	c <- fmt.Sprintf("%v\t%v\n", s.name(), median(perfs))
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
		total += f.Weighted(map[string]float64{"return": 1}, 0, time)
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
	weight              map[string]float64
}

func (b bestInPeriod) name() string {
	var key string
	var value float64
	for k, v := range b.weight {
		key = k
		value = v
	}
	return fmt.Sprintf("bestInPeriod(%v,%v,%v,%v,%v)", b.numFunds, b.monthsToRead, b.ignoreWithoutMonths, key, value)
}

func (b bestInPeriod) choose(funds []*xpfunds.Fund, end int) []*xpfunds.Fund {
	l := largestn.NewLargestN(b.numFunds)
	for _, f := range funds {
		if f.Duration()-end < b.ignoreWithoutMonths {
			continue
		}
		start := end + b.monthsToRead
		if b.monthsToRead == 0 {
			start = f.Duration()
		}
		l.Add(f, f.Weighted(b.weight, end, start))
	}
	return l.Funds
}
