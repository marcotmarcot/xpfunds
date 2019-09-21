package xpfunds

import (
	"io/ioutil"
	"strconv"
	"strings"
	"xpfunds/check"
)

type Fund struct {
	Name string

	// The monthly return of the fund, starting from the last month.
	monthly []float64

	// The annualized return in a period, from end (inclusive) to number of
	// months after end (inclusive). That is, to get the period of months 4
	// months starting at 1 and ending at 4 is in period[1][3].
	period [][]float64

	// The median return in a period stored in the same way as 'period'.
	median [][]float64
}

func ReadFunds() []*Fund {
	text, err := ioutil.ReadFile("get.tsv")
	check.Check(err)
	var funds []*Fund
	for _, line := range strings.Split(string(text), "\n") {
		f := fundFromLine(line)
		if f == nil {
			continue
		}
		funds = append(funds, f)
	}
	return funds
}

func fundFromLine(line string) *Fund {
	fields := strings.Split(strings.Trim(line, "\n"), "\t")
	if len(fields) < 6 {
		return nil
	}

	f := &Fund{}
	f.Name = fields[0]

	for i := 5; i < len(fields); i++ {
		v, err := strconv.ParseFloat(strings.Replace(fields[i], ",", ".", 1), 64)
		check.Check(err)
		f.monthly = append(f.monthly, 1.0+v/100.0)
	}
	f.setPeriod()
	f.setMedian()
	return f
}

// Return in the Period. end in the inclusive, start is exclusive.
func (f *Fund) Return(end, start int) float64 {
	return f.period[end][start-1-end]
}

// Returns the median return in period, similar to 'Return'.
func (f *Fund) MedianReturn(end, start int) float64 {
	return f.median[end][start-1-end]
}

func (f *Fund) Duration() int {
	return len(f.period)
}

func (f *Fund) setPeriod() {
	f.period = make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		f.period[end] = make([]float64, len(f.monthly)-end)
		f.period[end][0] = monthly
		for diff := 1; diff < len(f.monthly)-end; diff++ {
			f.period[end][diff] = f.period[end][diff-1] * f.monthly[end+diff]
		}
	}
}

func (f *Fund) setMedian() {
	f.median = make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		f.median[end] = make([]float64, len(f.monthly)-end)
		f.median[end][0] = monthly
		returns := []float64{monthly}
		for diff := 1; diff < len(f.monthly)-end; diff++ {
			returns = insert(returns, f.monthly[end+diff])
			f.median[end][diff] = medianFromSorted(returns)
		}
	}
}

func NewOptimum(funds []*Fund) *Fund {
	optimum := &Fund{}
	duration := maxDuration(funds)
	optimum.period = make([][]float64, duration)
	for end := range optimum.period {
		optimum.period[end] = make([]float64, duration-end)
		for diff := 0; diff < duration-end; diff++ {
			for _, fund := range funds {
				if end+diff >= fund.Duration() {
					continue
				}
				if fund.period[end][diff] > optimum.period[end][diff] {
					optimum.period[end][diff] = fund.period[end][diff]
				}
			}
		}
	}
	return optimum
}

func maxDuration(funds []*Fund) int {
	duration := 0
	for _, f := range funds {
		if f.Duration() > duration {
			duration = f.Duration()
		}
	}
	return duration
}

func insert(s []float64, e float64) []float64 {
	i := binarySearch(s, 0, len(s), e)
	return append(s[:i], append([]float64{e}, s[i:]...)...)
}

func binarySearch(s []float64, begin, end int, e float64) int {
	if end-begin == 0 {
		return begin
	}
	pivot := (begin + end) / 2
	if s[pivot] == e {
		return pivot
	} else if s[pivot] < e {
		return binarySearch(s, pivot+1, end, e)
	}
	return binarySearch(s, begin, pivot, e)
}

func medianFromSorted(s []float64) float64 {
	if len(s)%2 == 0 {
		return s[len(s)/2]
	}
	return (s[len(s)/2] + s[len(s)/2+1]) / 2
}
