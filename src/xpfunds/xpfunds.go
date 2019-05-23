package xpfunds

import (
	"io/ioutil"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func ReadFunds(index string) []*Fund {
	var ix *Fund
	if index != "" {
		ix = FundFromFile(index)
	}
	text, err := ioutil.ReadFile("get.tsv")
	Check(err)
	var funds []*Fund
	for _, line := range strings.Split(string(text), "\n") {
		f := fundFromLine(line, ix)
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

	// The annualized return in a period, from end (inclusive) to number of
	// months after end (inclusive). That is, to get the period of months 4
	// months starting at 1 and ending at 4 is in period[1][3].
	Period [][]float64

	// The mean annualized return in all subperiods inside this period. Same
	// rule as Period.
	MeanSubPeriods [][]float64

	// The annualized median return in a period.
	Median [][]float64
}

func fundFromLine(line string, ix *Fund) *Fund {
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
		if ix != nil {
			v -= (ix.Monthly[i - 5]-1.0)*100.0
		}
		f.Monthly = append(f.Monthly, 1.0+v/100.0)
	}
	f.setPeriod()
	return f
}

func FundFromFile(file string) *Fund {
	text, err := ioutil.ReadFile(file)
	Check(err)
	f := &Fund{}
	for _, line := range strings.Split(string(text), "\n") {
		value, err := strconv.ParseFloat(strings.Replace(strings.Trim(line, "\n"), ",", ".", 1), 64)
		if err != nil {
			break
		}
		f.Monthly = append(f.Monthly, 1.0+value/100.0)
	}
	f.setPeriod()
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

func (f *Fund) MeanSubPeriodsReturn(end, start int) float64 {
	return f.MeanSubPeriods[end][start-1-end]
}

func (f *Fund) MedianReturn(end, start int) float64 {
	return f.Median[end][start-1-end]
}

func (f *Fund) Duration() int {
	return len(f.Period)
}

func (f *Fund) setPeriod() {
	f.Period = make([][]float64, len(f.Monthly))
	for end, monthly := range f.Monthly {
		f.Period[end] = make([]float64, len(f.Monthly)-end)
		f.Period[end][0] = monthly
		for diff := 1; diff < len(f.Monthly)-end; diff++ {
			f.Period[end][diff] = f.Period[end][diff-1] * f.Monthly[end+diff]
		}
	}
	file := path.Join("subperiods", f.Name+".tsv")
	text, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		return
	}
	Check(err)
	lines := strings.Split(string(text), "\n")
	f.MeanSubPeriods = make([][]float64, len(f.Monthly))
	for end := range f.MeanSubPeriods {
		f.MeanSubPeriods[end] = make([]float64, len(f.Monthly)-end)
		fields := strings.Split(lines[end], "\t")
		for diff := range f.MeanSubPeriods[end] {
			f.MeanSubPeriods[end][diff], err = strconv.ParseFloat(strings.Replace(fields[diff], ",", ".", 1), 64)
			Check(err)
		}
	}
	f.Median = make([][]float64, len(f.Monthly))
	for end := range f.Monthly {
		f.Median[end] = make([]float64, len(f.Monthly)-end)
		var returns []float64
		for diff := 0; diff < len(f.Monthly)-end; diff++ {
			returns = Insert(returns, f.Monthly[end+diff])
			f.Median[end][diff] = returns[len(returns)/2]
		}
	}
}

func Annual(value float64, end, start int) float64 {
	// We subtract 1 because start is exclusive.
	return math.Pow(value, 12.0/float64(start-end))
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

func Mean(s []float64) float64 {
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

func Insert(list []float64, elem float64) []float64 {
	var i int
	for i = 0; i < len(list); i++ {
		if list[i] >= elem {
			break
		}
	}
	return append(list[:i], append([]float64{elem}, list[i:]...)...)
}
