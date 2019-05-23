package simulate

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sort"
	"strconv"
	"strings"
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
	for numFunds := 5; numFunds <= 5; numFunds++ {
		strategies = append(strategies,
			// &random{numFunds},
			// &minAndDays{numFunds},
			// newMl(numFunds),
		)

		// How many months to check data for. 0 for all history.
		for numMonths := 0; numMonths <= 48; numMonths += 1 {
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
		for minTime := 1; minTime <= 12; minTime += 1 {

			// future: Mean future return
			// loss: Mean (future return / best possible return in future)
			// cdi_ratio: Mean (future return / cdi)
			// min: Minimal future return
			future, loss, cdi_ratio, min := meanPerformance(funds, cdi, optimum, s, minTime)
			fmt.Println(s.name(), minTime, future, loss, cdi_ratio, min)
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
			if funds[i].Duration()-time > minTime {
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
	return xpfunds.Annual(total_future / float64(len(indexes)), 0, time)
}

type strategy interface {
	name() string

	// Given a list of valid index, return them in the order they should be
	// picked. Only consider data in funds up to end (inclusive).
	choose(funds []*xpfunds.Fund, indexes []int, end int) []int
}

type random struct {
	numFunds int
}

func (r *random) name() string {
	return "Random" + strconv.Itoa(r.numFunds)
}

func (r *random) choose(funds []*xpfunds.Fund, indexes []int, end int) []int {
	// Shuffles the indexes slice.
	chosen := make([]int, len(indexes))
	for i, index := range rand.Perm(len(indexes)) {
		chosen[i] = indexes[index]
	}
	if len(chosen) > r.numFunds {
		chosen = chosen[:r.numFunds]
	}
	return chosen
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
	numFunds  int
	predicted map[fundTime]float64
}

func newMl(numFunds int) *ml {
	metadata_text, err := ioutil.ReadFile("test_metadata.tsv")
	xpfunds.Check(err)
	predictions_text, err := ioutil.ReadFile("test_predictions.tsv")
	xpfunds.Check(err)
	metadata_lines := strings.Split(string(metadata_text), "\n")
	predictions_lines := strings.Split(string(predictions_text), "\n")
	ml := &ml{numFunds, make(map[fundTime]float64)}
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
	return "Ml" + strconv.Itoa(m.numFunds)
}

func (m *ml) choose(funds []*xpfunds.Fund, indexes []int, end int) []int {
	sort.Sort(byMl{indexes, funds, end, m.predicted})
	if len(indexes) > m.numFunds {
		indexes = indexes[:m.numFunds]
	}
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

type minAndDays struct {
	numFunds int
}

func (m *minAndDays) name() string {
	return "MinAndDays" + strconv.Itoa(m.numFunds)
}

func (m *minAndDays) choose(funds []*xpfunds.Fund, indexes []int, end int) []int {
	sort.Sort(byMinAndDays{indexes, funds})
	if len(indexes) > m.numFunds {
		indexes = indexes[:m.numFunds]
	}
	return indexes
}

type byMinAndDays struct {
	indexes []int
	funds   []*xpfunds.Fund
}

func (b byMinAndDays) Len() int {
	return len(b.indexes)
}

func (b byMinAndDays) Swap(i, j int) {
	b.indexes[i], b.indexes[j] = b.indexes[j], b.indexes[i]
}

func (b byMinAndDays) Less(i, j int) bool {
	mi := b.funds[b.indexes[i]].Min
	mj := b.funds[b.indexes[j]].Min
	if mi == mj {
		di := b.funds[b.indexes[i]].Days
		dj := b.funds[b.indexes[j]].Days
		return di < dj
	}
	return mi < mj
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
	start := f.Duration()
	if numMonths != 0 && end+numMonths < f.Duration() {
		start = end + numMonths
	}
	return f.MeanSubPeriodsReturn(end, start)
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
	start := f.Duration()
	if numMonths != 0 && end+numMonths < f.Duration() {
		start = end + numMonths
	}
	return f.MedianReturn(end, start)
}
