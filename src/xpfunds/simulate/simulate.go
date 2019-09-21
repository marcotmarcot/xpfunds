package simulate

import (
	"flag"
	"fmt"
	"sort"
	"strconv"
	"xpfunds"
)

var start_time = flag.Int("start_time", -1,
	"The beginning of the evaluation period in months from the last month.")

var end_time = flag.Int("end_time", -1,
	"The end of the evaluation period in months from the last month.")

var index = flag.String("index", "",
	"An index to subtract the gains from.")

func Main() {
	flag.Parse()
	funds := xpfunds.ReadFunds(*index)
	optimum := xpfunds.NewOptimum(funds)
	cdi := xpfunds.FundFromFile("cdi.tsv")
	var strategies []strategy
	for numFunds := 1; numFunds <= 10; numFunds++ {
		// How many months to check data for
		for numMonths := 1; numMonths <= 240; numMonths += 12 {
			strategies = append(strategies,
				&fromStart{numFunds, numMonths, false},
				&fromStart{numFunds, numMonths, true},
				&meanSubPeriods{numFunds, numMonths, false},
				&meanSubPeriods{numFunds, numMonths, true},
				&median{numFunds, numMonths, false},
				&median{numFunds, numMonths, true})
		}
	}

	for _, s := range strategies {

		// Discard funds that don't have at least that many months.
		for minTime := 1; minTime <= 240; minTime += 12 {

			// future: Mean future return
			// loss: Mean (future return / best possible return in future)
			// cdi_ratio: Mean (future return / cdi)
			// min: Minimal future return
			future, loss, cdi_ratio, min := meanPerformance(funds, cdi, optimum, s, minTime)
			fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", s.name(), minTime, future, loss, cdi_ratio, min)
		}
	}
}

func meanPerformance(funds []*xpfunds.Fund, cdi, optimum *xpfunds.Fund, s strategy, minTime int) (future, loss, cdi_ratio, min float64) {
	var futures []float64
	var losses []float64
	var cdi_ratios []float64

	start := *start_time
	if start == -1 {
		// We start at 1 and end before the last element because we need to have
		// at least one month in the beginning to get the data and one month in
		// the end to check the performance. We divide the duration by 2 because
		// the first part was used for trainning.
		start = optimum.Duration() - 1
	}
	end := *end_time
	if end == -1 {
		end = 1
	}
	for time := end; time <= start; time++ {

		// Create a list with all indexes to funds.
		var allIndexes []int
		for i := 0; i < len(funds); i++ {
			allIndexes = append(allIndexes, i)
		}

		// Filter using minTime.
		var filtered []int
		for _, i := range allIndexes {
			if funds[i].Duration()-time > minTime && funds[i].Duration()-time > s.months() {
				filtered = append(filtered, i)
			}
		}
		if len(filtered) <= 0 {
			continue
		}

		future := performance(funds, optimum, filtered, s, time)
		futures = append(futures, future)
		losses = append(losses, future/optimum.Annual(0, time))
		cdi_ratios = append(cdi_ratios, future/cdi.Annual(0, time))
	}
	return xpfunds.Mean(futures), xpfunds.Mean(losses), xpfunds.Mean(cdi_ratios), minimum(futures)
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

func performance(funds []*xpfunds.Fund, optimum *xpfunds.Fund, indexes []int, s strategy, time int) float64 {
	indexes = s.choose(funds, indexes, time)
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

	months() int
}

type fromStart struct {
	numFunds  int
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
	return name + strconv.Itoa(f.numFunds) + "," + strconv.Itoa(f.numMonths)
}

func (f *fromStart) choose(funds []*xpfunds.Fund, indexes []int, end int) []int {
	sort.Sort(byReturnFromStart{indexes, funds, end, f.numMonths, f.reverse})
	if len(indexes) > f.numFunds {
		indexes = indexes[:f.numFunds]
	}
	return indexes
}

func (f *fromStart) months() int {
	return f.numMonths
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
	if end+numMonths >= f.Duration() {
		return -9999
	}
	return f.Annual(end, end + numMonths)
}

type meanSubPeriods struct {
	numFunds  int
	numMonths int
	reverse   bool
}

func (m *meanSubPeriods) name() string {
	var name string
	if m.reverse {
		name = "WorstMSP"
	} else {
		name = "BestMSP"
	}
	return name + strconv.Itoa(m.numFunds) + "," + strconv.Itoa(m.numMonths)
}

func (m *meanSubPeriods) choose(funds []*xpfunds.Fund, indexes []int, end int) []int {
	sort.Sort(byMeanSubPeriods{indexes, funds, end, m.numMonths, m.reverse})
	if len(indexes) > m.numFunds {
		indexes = indexes[:m.numFunds]
	}
	return indexes
}

func (m *meanSubPeriods) months() int {
	return m.numMonths
}

type byMeanSubPeriods struct {
	indexes   []int
	funds     []*xpfunds.Fund
	end       int
	numMonths int
	reverse   bool
}

func (b byMeanSubPeriods) Len() int {
	return len(b.indexes)
}

func (b byMeanSubPeriods) Swap(i, j int) {
	b.indexes[i], b.indexes[j] = b.indexes[j], b.indexes[i]
}

func (b byMeanSubPeriods) Less(i, j int) bool {
	ri := meanSubPeriodsReturn(b.funds[b.indexes[i]], b.end, b.numMonths)
	rj := meanSubPeriodsReturn(b.funds[b.indexes[j]], b.end, b.numMonths)
	if b.reverse {
		return rj > ri
	}
	return ri > rj
}

func meanSubPeriodsReturn(f *xpfunds.Fund, end, numMonths int) float64 {
	if end+numMonths >= f.Duration() {
		return -9999
	}
	return f.MeanSubPeriodsReturn(end, end+numMonths)
}

type median struct {
	numFunds  int
	numMonths int
	reverse   bool
}

func (m *median) name() string {
	var name string
	if m.reverse {
		name = "WorstMedian"
	} else {
		name = "BestMedian"
	}
	return name + strconv.Itoa(m.numFunds) + "," + strconv.Itoa(m.numMonths)
}

func (m *median) choose(funds []*xpfunds.Fund, indexes []int, end int) []int {
	sort.Sort(byMedian{indexes, funds, end, m.numMonths, m.reverse})
	if len(indexes) > m.numFunds {
		indexes = indexes[:m.numFunds]
	}
	return indexes
}

func (m *median) months() int {
	return m.numMonths
}

type byMedian struct {
	indexes   []int
	funds     []*xpfunds.Fund
	end       int
	numMonths int
	reverse   bool
}

func (b byMedian) Len() int {
	return len(b.indexes)
}

func (b byMedian) Swap(i, j int) {
	b.indexes[i], b.indexes[j] = b.indexes[j], b.indexes[i]
}

func (b byMedian) Less(i, j int) bool {
	ri := medianReturn(b.funds[b.indexes[i]], b.end, b.numMonths)
	rj := medianReturn(b.funds[b.indexes[j]], b.end, b.numMonths)
	if b.reverse {
		return rj > ri
	}
	return ri > rj
}

func medianReturn(f *xpfunds.Fund, end, numMonths int) float64 {
	if end+numMonths >= f.Duration() {
		return -9999
	}
	return f.MedianReturn(end, end+numMonths)
}
