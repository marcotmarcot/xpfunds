package simulate

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"xpfunds"
)

func Main() {
	funds := xpfunds.ReadFunds()
	optimum := newOptimum(funds)
	strategies := []strategy{
		&random{},
		&fromStart{0, false},
		&fromStart{0, true},
		newMl()}

	// How many months to check data for. 0 for all history.
	for numMonths := 96; numMonths <= 98; numMonths += 1 {
		strategies = append(strategies, &fromStart{numMonths, false}, &fromStart{numMonths, true})
	}

	for _, s := range strategies {

		// How many funds to pick.
		for numFunds := 1; numFunds <= 12; numFunds += 11 {

				// Discard funds that don't have at least that many months.
				for minTime := 1; minTime <= 1; minTime += 1 {

				// future: Mean future return
				// loss: Mean (future return / best possible return in future)
				// min: Minimal future return
				future, loss, min := meanPerformance(funds, optimum, s, minTime, numFunds)
				fmt.Println(s.name(), numFunds, minTime, future, loss, min)
			}
		}
	}
}

func newOptimum(funds []*xpfunds.Fund) *xpfunds.Fund {
	optimum := &xpfunds.Fund{}
	duration := xpfunds.MaxDuration(funds)
	optimum.Period = make([][]float64, duration)
	for end := range optimum.Period {
		optimum.Period[end] = make([]float64, duration-end)
		for diff := 0; diff < duration-end; diff++ {
			for _, fund := range funds {
				if end + diff >= fund.Duration() {
					continue
				}
				if fund.Period[end][diff] > optimum.Period[end][diff] {
					optimum.Period[end][diff] = fund.Period[end][diff]
				}
			}
		}
	}
	fmt.Println(optimum.Annual(0, optimum.Duration() / 2))
	return optimum
}

func meanPerformance(funds []*xpfunds.Fund, optimum *xpfunds.Fund, s strategy, minTime, numFunds int) (future, loss, min float64) {
	var futures []float64
	var losses []float64

	// We start at 1 and end before the last element because we need to have at
	// least one month in the beginning to get the data and one month in the end
	// to check the performance. We divide the duration by 2 because the first
	// part was used for trainning.
	for time := 1; time < optimum.Duration() / 2 - 1; time++ {

		// Create a list with all indexes to funds.
		var allIndexes []int
		for i := 0; i < len(funds); i++ {
			allIndexes = append(allIndexes, i)
		}

		// Filter using minTime.
		var filtered []int
		for _, i := range allIndexes {
			if funds[i].Duration()-time > minTime {
				filtered = append(filtered, i)
			}
		}
		if len(filtered) <= 0 {
			continue
		}

		future := performance(funds, optimum, filtered, s, numFunds, time)
		futures = append(futures, future)
		losses = append(losses, future/optimum.Annual(0, time))
	}
	return mean(futures), mean(losses), minimum(futures)
}

func mean(s []float64) float64 {
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

func minimum(s []float64) float64 {
	if len(s) == 0 {
		return -1
	}
	min := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] < min {
			min = s[i]
		}
	}
	return min
}

func performance(funds []*xpfunds.Fund, optimum *xpfunds.Fund, indexes []int, s strategy, numFunds, time int) float64 {
	indexes = s.choose(funds, indexes, time)
	if len(indexes) > numFunds {
		indexes = indexes[:numFunds]
	}
	total_future := 0.0
	for _, i := range indexes {
		total_future += funds[i].Return(0, time)
	}
	return xpfunds.Annual(total_future/float64(len(indexes)), 0, time)
}

type strategy interface {
	name() string

	// Given a list of valid index, return them in the order they should be
	// picked. Only consider data in funds up to end (inclusive).
	choose(funds []*xpfunds.Fund, indexes []int, end int) []int
}

type random struct{}

func (r *random) name() string {
	return "Random"
}

func (r *random) choose(funds []*xpfunds.Fund, indexes []int, end int) []int {
	// Shuffles the indexes slice.
	chosen := make([]int, len(indexes))
	for i, index := range rand.Perm(len(indexes)) {
		chosen[i] = indexes[index]
	}
	return chosen
}

type fromStart struct {
	numMonths int
	reverse   bool
}

func (f *fromStart) name() string {
	var name string
	if f.reverse {
		name = "Worst"
	} else {
		name = "Best"
	}
	return name + strconv.Itoa(f.numMonths)
}

func (f *fromStart) choose(funds []*xpfunds.Fund, indexes []int, end int) []int {
	sort.Sort(byReturnFromStart{indexes, funds, end, f.numMonths, f.reverse})
	return indexes
}

type byReturnFromStart struct {
	indexes   []int
	funds     []*xpfunds.Fund
	end       int
	numMonths int
	reverse   bool
}

func (b byReturnFromStart) Len() int {
	return len(b.indexes)
}

func (b byReturnFromStart) Swap(i, j int) {
	b.indexes[i], b.indexes[j] = b.indexes[j], b.indexes[i]
}

func (b byReturnFromStart) Less(i, j int) bool {
	ri := returnFromStart(b.funds[b.indexes[i]], b.end, b.numMonths)
	rj := returnFromStart(b.funds[b.indexes[j]], b.end, b.numMonths)
	if b.reverse {
		return rj > ri
	}
	return ri > rj
}

func returnFromStart(f *xpfunds.Fund, end, numMonths int) float64 {
	start := f.Duration()
	if numMonths != 0 && end+numMonths < f.Duration() {
		start = end + numMonths
	}
	return f.Annual(end, start)
}

type ml struct {
	predicted map[fundTime]float64
}

func newMl() *ml {
	metadata_text, err := ioutil.ReadFile("test_metadata.tsv")
	xpfunds.Check(err)
	predictions_text, err := ioutil.ReadFile("test_predictions.tsv")
	xpfunds.Check(err)
	metadata_lines := strings.Split(string(metadata_text), "\n")
	predictions_lines := strings.Split(string(predictions_text), "\n")
	ml := &ml{make(map[fundTime]float64)}
	for i := range metadata_lines {
		fields := strings.Split(strings.Trim(metadata_lines[i], "\n"), "\t")
		if len(fields) < 2 {
			continue
		}
		time, err := strconv.Atoi(fields[1])
		xpfunds.Check(err)
		label, err := strconv.ParseFloat(predictions_lines[i], 64)
		xpfunds.Check(err)
		ml.predicted[fundTime{fields[0], time}] = label
	}
	return ml
}

type fundTime struct {
	name string
	time int
}

func (m *ml) name() string {
	return "Ml"
}

func (m *ml) choose(funds []*xpfunds.Fund, indexes []int, end int) []int {
	sort.Sort(byMl{indexes, funds, end, m.predicted})
	return indexes
}

type byMl struct {
	indexes   []int
	funds     []*xpfunds.Fund
	end       int
	predicted map[fundTime]float64
}

func (b byMl) Len() int {
	return len(b.indexes)
}

func (b byMl) Swap(i, j int) {
	b.indexes[i], b.indexes[j] = b.indexes[j], b.indexes[i]
}

func (b byMl) Less(i, j int) bool {
	ri := b.predicted[fundTime{b.funds[b.indexes[i]].Name, b.end}]
	rj := b.predicted[fundTime{b.funds[b.indexes[j]].Name, b.end}]
	return ri > rj
}
