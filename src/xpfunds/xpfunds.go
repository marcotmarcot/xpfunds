package xpfunds

import (
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"xpfunds/binarysearch"
	"xpfunds/check"
)

type Fund struct {
	Name string

	// The monthly return of the fund, starting from the last month.
	monthly []float64

	// The annualized return in a period, from end (inclusive) to number of
	// months after end (inclusive). That is, to get the period of months 4
	// months starting at 1 and ending at 4 is in ret[1][3].
	ret [][]float64

	// The median return in a period stored in the same way as 'ret'.
	median [][]float64

	stdDev [][]float64

	negativeMonthRatio [][]float64
}

func NewFund(n string, monthly []float64) *Fund {
	f := &Fund{
		Name:    n,
		monthly: monthly,
	}
	f.setFields()
	return f
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

	var monthly []float64
	for i := 5; i < len(fields); i++ {
		v, err := strconv.ParseFloat(strings.Replace(fields[i], ",", ".", 1), 64)
		check.Check(err)
		monthly = append(monthly, 1.0+v/100.0)
	}
	return NewFund(fields[0], monthly)
}

func (f *Fund) Duration() int {
	return len(f.ret)
}

func (f *Fund) setFields() {
	f.setRet()
	f.setMedian()
	f.setStdDev()
	f.setNegativeMonthRatio()
}

func (f *Fund) setRet() {
	f.ret = make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		f.ret[end] = make([]float64, len(f.monthly)-end)
		f.ret[end][0] = monthly
		for diff := 1; diff < len(f.monthly)-end; diff++ {
			f.ret[end][diff] = f.ret[end][diff-1] * f.monthly[end+diff]
		}
	}
}

// Return in the Period. end in the inclusive, start is exclusive.
func (f *Fund) Ret(end, start int) float64 {
	return f.ret[end][start-1-end]
}

func (f *Fund) setMedian() {
	f.median = make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		f.median[end] = make([]float64, len(f.monthly)-end)
		f.median[end][0] = monthly
		returns := []float64{monthly}
		for diff := 1; diff < len(f.monthly)-end; diff++ {
			returns = binarysearch.InsertInSorted(returns, f.monthly[end+diff])
			f.median[end][diff] = binarysearch.MedianFromSorted(returns)
		}
	}
}

// Returns the median return in period, similar to 'Ret'.
func (f *Fund) medianInPeriod(end, start int) float64 {
	return f.median[end][start-1-end]
}

func (f *Fund) setStdDev() {
	f.stdDev = make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		f.stdDev[end] = make([]float64, len(f.monthly)-end)
		f.stdDev[end][0] = 0
		total := monthly
		for diff := 1; diff < len(f.monthly)-end; diff++ {
			total += f.monthly[end+diff]
			count := float64(diff + 1)
			avg := total / count
			sumDiffs := 0.0
			for i := end; i <= end+diff; i++ {
				diff := f.monthly[i] - avg
				sumDiffs += diff * diff
			}
			f.stdDev[end][diff] = math.Sqrt(sumDiffs / count)
		}
	}
}

func (f *Fund) stdDevInPeriod(end, start int) float64 {
	return f.stdDev[end][start-1-end]
}

func (f *Fund) setNegativeMonthRatio() {
	f.negativeMonthRatio = make([][]float64, len(f.monthly))
	for end := range f.monthly {
		f.negativeMonthRatio[end] = make([]float64, len(f.monthly)-end)
		negative := 0
		nonNegative := 0
		for diff := 0; diff < len(f.monthly)-end; diff++ {
			if f.monthly[end+diff] < 1 {
				negative++
			} else {
				nonNegative++
			}
			f.negativeMonthRatio[end][diff] = float64(negative) / float64(negative+nonNegative)
		}
	}
}

func (f *Fund) negativeMonthRatioInPeriod(end, start int) float64 {
	return f.negativeMonthRatio[end][start-1-end]
}

func NewOptimum(funds []*Fund) *Fund {
	optimum := &Fund{}
	duration := maxDuration(funds)
	optimum.ret = make([][]float64, duration)
	for end := range optimum.ret {
		optimum.ret[end] = make([]float64, duration-end)
		for diff := 0; diff < duration-end; diff++ {
			for _, fund := range funds {
				if end+diff >= fund.Duration() {
					continue
				}
				if fund.ret[end][diff] > optimum.ret[end][diff] {
					optimum.ret[end][diff] = fund.ret[end][diff]
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

var Fields = map[string]func(f *Fund, start, end int) float64{
	"ret":                (*Fund).Ret,
	"median":             (*Fund).medianInPeriod,
	"stdDev":             (*Fund).stdDevInPeriod,
	"negativeMonthRatio": (*Fund).negativeMonthRatioInPeriod,
}
