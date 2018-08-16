package xpfunds

import (
	"io/ioutil"
	"math"
	"strconv"
	"strings"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func ReadFunds() []*Fund {
	text, err := ioutil.ReadFile("get.tsv")
	Check(err)
	var funds []*Fund
	for _, line := range strings.Split(string(text), "\n") {
		f := newFund(line)
		if f == nil {
			continue
		}
		funds = append(funds, f)
	}
	return funds
}

type Fund struct {
	Name string

	// The minimum value that can be invested in this fund.
	Min int

	// The number of days that we need to wait to invest on this fund.
	Days int

	// The monthly return of the fund, starting from the last month.
	Monthly []float64

	// The return in a period, from end (inclusive) to number of months after
	// end (inclusive). That is, to get the period of months 4 months starting
	// at 1 and ending at 4 is in period[1][3].
	Period [][]float64
}

func newFund(line string) *Fund {
	fields := strings.Split(strings.Trim(line, "\n"), "\t")
	if len(fields) < 6 {
		return nil
	}

	f := &Fund{}
	f.Name = fields[0]

	var err error
	f.Min, err = strconv.Atoi(strings.Replace(strings.Replace(fields[1], ",00", "", 1), ".", "", -1))
	Check(err)

	liq, err := strconv.Atoi(fields[2])
	Check(err)
	cot, err := strconv.Atoi(fields[3])
	Check(err)
	f.Days = liq + cot

	for i := 5; i < len(fields); i++ {
		v, err := strconv.ParseFloat(strings.Replace(fields[i], ",", ".", 1), 64)
		Check(err)
		f.Monthly = append(f.Monthly, 1.0+v/100.0)
	}
	f.Period = make([][]float64, len(f.Monthly))
	for end, monthly := range f.Monthly {
		f.Period[end] = make([]float64, len(f.Monthly)-end)
		f.Period[end][0] = monthly
		for diff := 1; diff < len(f.Monthly)-end; diff++ {
			f.Period[end][diff] = f.Period[end][diff-1] * f.Monthly[end+diff]
		}
	}
	return f
}

// Return in the Period. end in the inclusive, start is exclusive.
func (f *Fund) Return(end, start int) float64 {
	return f.Period[end][start-1-end]
}

// Annualized return in the Period. end in the inclusive, start is exclusive.
func (f *Fund) Annual(end, start int) float64 {
	return Annual(f.Return(end, start), end, start)
}

func Annual(value float64, end, start int) float64 {
	// We subtract 1 because start is exclusive.
	return math.Pow(value, 12.0/float64(start-end))
}

func (f *Fund) Duration() int {
	return len(f.Period)
}

func MaxDuration(funds []*Fund) int {
	duration := 0
	for _, f := range funds {
		if f.Duration() > duration {
			duration = f.Duration()
		}
	}
	return duration
}

func ReadLines(file string) []float64 {
	text, err := ioutil.ReadFile(file)
	Check(err)
	var values []float64
	for _, line := range strings.Split(string(text), "\n") {
		value, err := strconv.ParseFloat(strings.Replace(strings.Trim(line, "\n"), ",", ".", 1), 64)
		if err != nil {
			break
		}
		values = append(values, 1.0+value/100.0)
	}
	return values
}

func NewOptimum(funds []*Fund) *Fund {
	optimum := &Fund{}
	duration := MaxDuration(funds)
	optimum.Period = make([][]float64, duration)
	for end := range optimum.Period {
		optimum.Period[end] = make([]float64, duration-end)
		for diff := 0; diff < duration-end; diff++ {
			for _, fund := range funds {
				if end+diff >= fund.Duration() {
					continue
				}
				if fund.Period[end][diff] > optimum.Period[end][diff] {
					optimum.Period[end][diff] = fund.Period[end][diff]
				}
			}
		}
	}
	return optimum
}
