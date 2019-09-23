package xpfunds

import (
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"xpfunds/binarysearch"
	"xpfunds/check"
)

type Fund struct {
	name string

	active string

	min string

	// The monthly return of the fund, starting from the last month.
	monthly []float64

	fields map[string][][]float64

	ratio map[string][][]float64
}

func NewFund(n string, monthly []float64) *Fund {
	f := &Fund{
		name:    n,
		monthly: monthly,
		fields:  make(map[string][][]float64),
		ratio:   make(map[string][][]float64),
	}
	f.setFields()
	return f
}

func (f *Fund) setFields() {
	f.setReturn()
	f.setMedian()
	f.setStdDev()
	f.setNegativeMonthRatio()
	f.setGreatestFall()
}

func (f *Fund) setReturn() {
	f.fields["return"] = make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		f.fields["return"][end] = make([]float64, len(f.monthly)-end)
		f.fields["return"][end][0] = monthly
		for diff := 1; diff < len(f.monthly)-end; diff++ {
			f.fields["return"][end][diff] = f.fields["return"][end][diff-1] * f.monthly[end+diff]
		}
	}
}

func (f *Fund) setMedian() {
	f.fields["median"] = make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		f.fields["median"][end] = make([]float64, len(f.monthly)-end)
		f.fields["median"][end][0] = monthly
		returns := []float64{monthly}
		for diff := 1; diff < len(f.monthly)-end; diff++ {
			returns = binarysearch.InsertInSorted(returns, f.monthly[end+diff])
			f.fields["median"][end][diff] = binarysearch.MedianFromSorted(returns)
		}
	}
}

func (f *Fund) setStdDev() {
	f.fields["stdDev"] = make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		f.fields["stdDev"][end] = make([]float64, len(f.monthly)-end)
		f.fields["stdDev"][end][0] = 0
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
			f.fields["stdDev"][end][diff] = math.Sqrt(sumDiffs / count)
		}
	}
}

func (f *Fund) setNegativeMonthRatio() {
	f.fields["negativeMonthRatio"] = make([][]float64, len(f.monthly))
	for end := range f.monthly {
		f.fields["negativeMonthRatio"][end] = make([]float64, len(f.monthly)-end)
		negative := 0
		nonNegative := 0
		for diff := 0; diff < len(f.monthly)-end; diff++ {
			if f.monthly[end+diff] < 1 {
				negative++
			} else {
				nonNegative++
			}
			f.fields["negativeMonthRatio"][end][diff] = float64(negative) / float64(negative+nonNegative)
		}
	}
}

func (f *Fund) setGreatestFall() {
	f.fields["greatestFall"] = make([][]float64, len(f.monthly))
	f.fields["greatestFallLen"] = make([][]float64, len(f.monthly))
	for end := range f.monthly {
		f.fields["greatestFall"][end] = make([]float64, len(f.monthly)-end)
		f.fields["greatestFallLen"][end] = make([]float64, len(f.monthly)-end)
		greatestFall := 1.0
		greatestFallLen := 0
		curr := 1.0
		currLen := 0
		for diff := 0; diff < len(f.monthly)-end; diff++ {
			curr *= f.monthly[end+diff]
			currLen++
			if f.monthly[end+diff] < curr {
				curr = f.monthly[end+diff]
				currLen = 1
			}
			if curr < greatestFall {
				greatestFall = curr
				greatestFallLen = currLen
			}
			f.fields["greatestFall"][end][diff] = greatestFall
			f.fields["greatestFallLen"][end][diff] = float64(greatestFallLen)
		}
	}
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
	setRatio(funds)
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
	f := NewFund(fields[0], monthly)
	f.active = fields[4]
	f.min = fields[1]
	return f
}

func (f *Fund) Duration() int {
	return len(f.fields["return"])
}

func (f *Fund) Fields() []string {
	fields := make([]string, len(f.fields))
	i := 0
	for field := range f.fields {
		fields[i] = field
		i++
	}
	return fields
}

// End is inclusive, start is exclusive
func (f *Fund) Weighted(weight map[string]float64, end, start int) float64 {
	total := 0.0
	for field, value := range weight {
		total += f.ratio[field][end][start-1-end] * value
	}
	return total
}

func (f *Fund) Print() string {
	return fmt.Sprintf("%v\t%v\t%v", f.name, f.active, f.min)
}

func setRatio(funds []*Fund) {
	optimum := &Fund{fields: make(map[string][][]float64)}
	duration := maxDuration(funds)
	for _, field := range funds[0].Fields() {
		optimum.fields[field] = make([][]float64, duration)
		for _, f := range funds {
			f.ratio[field] = make([][]float64, duration)
		}
		for end := range optimum.fields[field] {
			optimum.fields[field][end] = make([]float64, duration-end)
			for _, f := range funds {
				f.ratio[field][end] = make([]float64, duration-end)
			}
			for diff := 0; diff < duration-end; diff++ {
				optimum.fields[field][end][diff] = -999999.99
				for _, fund := range funds {
					if end+diff >= fund.Duration() {
						continue
					}
					if fund.fields[field][end][diff] > optimum.fields[field][end][diff] {
						optimum.fields[field][end][diff] = fund.fields[field][end][diff]
					}
				}
				for _, f := range funds {
					if f.Duration() <= end+diff {
						continue
					}
					if optimum.fields[field][end][diff] == 0 {
						f.ratio[field][end][diff] = 1
						continue
					}
					f.ratio[field][end][diff] = f.fields[field][end][diff] / optimum.fields[field][end][diff]
				}
			}
		}
	}
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
