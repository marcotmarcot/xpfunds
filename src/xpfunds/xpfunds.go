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

	// The return in a period, from end (inclusive) to number of months after
	// end (inclusive). That is, to get the period of months 4 months starting
	// at 1 and ending at 4 is in period[1][3].
	Period [][]float64
}

func newFund(line string) *Fund {
	fields := strings.Split(strings.Trim(line, "\n"), "\t")
	if fields == nil {
		return nil
	}

	f := &Fund{}
	f.Name = fields[0]

	// The monthly return of the fund, starting from the last month.
	var monthly []float64
	for i := 5; i < len(fields); i++ {
		v, err := strconv.ParseFloat(strings.Replace(fields[i], ",", ".", 1), 64)
		Check(err)
		monthly = append(monthly, 1.0+v/100.0)
	}
	if len(monthly) == 0 {
		return nil
	}
	f.Period = make([][]float64, len(monthly))
	for end, Monthly := range monthly {
		f.Period[end] = make([]float64, len(monthly)-end)
		f.Period[end][0] = Monthly
		for diff := 1; diff < len(monthly)-end; diff++ {
			f.Period[end][diff] = f.Period[end][diff-1] * monthly[end+diff]
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
