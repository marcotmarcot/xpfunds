package xpfunds

import (
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"xpfunds/binarysearch"
	"xpfunds/check"
	"xpfunds/median"
)

type Fund struct {
	name string

	active string

	min string

	// The monthly return of the fund, starting from the last month.
	monthly []float64

	// The position of the first slice determines the dimension. The position of
	// the second slice indicates an end time of a period and the third position
	// the difference from the start time to the end time of a period. Arbitrary
	// range.
	features [][][]float64

	// Same as fieldValues, but holds the ratio of the value in this fund to the
	// value in the fund with the highest value of this field. Range: 0-1.
	ratio [][][]float64
}

func NewFund(monthly []float64) *Fund {
	f := &Fund{
		monthly: monthly,
	}
	f.setFeatures()
	f.makeRatio()
	return f
}

func (f *Fund) setFeatures() {
	f.setReturn()
	f.setMedian()
	f.setStdDev()
	f.setNegativeMonthRatio()
	f.setGreatestFall()
}

func (f *Fund) setReturn() {
	ret := make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		ret[end] = make([]float64, len(f.monthly)-end)
		ret[end][0] = monthly
		for diff := 1; diff < len(f.monthly)-end; diff++ {
			ret[end][diff] = ret[end][diff-1] * f.monthly[end+diff]
		}
	}
	f.features = append(f.features, ret)
}

func (f *Fund) setMedian() {
	med := make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		med[end] = make([]float64, len(f.monthly)-end)
		med[end][0] = monthly
		returns := []float64{monthly}
		for diff := 1; diff < len(f.monthly)-end; diff++ {
			returns = binarysearch.InsertInSorted(returns, f.monthly[end+diff])
			med[end][diff] = median.MedianFromSorted(returns)
		}
	}
	f.features = append(f.features, med)
}

func (f *Fund) setStdDev() {
	stdDev := make([][]float64, len(f.monthly))
	for end, monthly := range f.monthly {
		stdDev[end] = make([]float64, len(f.monthly)-end)
		stdDev[end][0] = 0
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
			stdDev[end][diff] = math.Sqrt(sumDiffs / count)
		}
	}
	f.features = append(f.features, stdDev)
}

func (f *Fund) setNegativeMonthRatio() {
	nmr := make([][]float64, len(f.monthly))
	for end := range f.monthly {
		nmr[end] = make([]float64, len(f.monthly)-end)
		negative := 0
		nonNegative := 0
		for diff := 0; diff < len(f.monthly)-end; diff++ {
			if f.monthly[end+diff] < 1 {
				negative++
			} else {
				nonNegative++
			}
			nmr[end][diff] = float64(negative) / float64(negative+nonNegative)
		}
	}
	f.features = append(f.features, nmr)
}

func (f *Fund) setGreatestFall() {
	gf := make([][]float64, len(f.monthly))
	gfl := make([][]float64, len(f.monthly))
	for end := range f.monthly {
		gf[end] = make([]float64, len(f.monthly)-end)
		gfl[end] = make([]float64, len(f.monthly)-end)
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
			gf[end][diff] = greatestFall
			gfl[end][diff] = float64(greatestFallLen)
		}
	}
	f.features = append(f.features, gf, gfl)
}

func (f *Fund) makeRatio() {
	f.ratio = make([][][]float64, f.FeatureCount())
	for feature := range f.ratio {
		f.ratio[feature] = make([][]float64, f.Duration())
		for end := range f.ratio[feature] {
			f.ratio[feature][end] = make([]float64, f.Duration()-end)
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
	SetRatio(funds)
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
	f := NewFund(monthly)
	f.name = fields[0]
	f.active = fields[4]
	f.min = fields[1]
	return f
}

func (f *Fund) FeatureCount() int {
	return len(f.features)
}

func (f *Fund) Duration() int {
	return len(f.monthly)
}

// End is inclusive, start is exclusive
func (f *Fund) Weighted(weight []float64, end, start int) float64 {
	total := 0.0
	for i, w := range weight {
		total += f.ratio[i][end][start-1-end] * w
	}
	return total
}

func (f *Fund) Return(end, start int) float64 {
	return f.Weighted([]float64{1}, end, start)
}

func (f *Fund) Print() string {
	return fmt.Sprintf("%v\t%v\t%v", f.name, f.active, f.min)
}

func SetRatio(funds []*Fund) {
	duration := MaxDuration(funds)
	for feature := 0; feature < funds[0].FeatureCount(); feature++ {
		for end := range funds[0].features[feature] {
			for diff := 0; diff < duration-end; diff++ {
				highest := -999999.99
				for _, f := range funds {
					if f.Duration() <= end+diff {
						continue
					}
					if f.features[feature][end][diff] > highest {
						highest = f.features[feature][end][diff]
					}
				}
				for _, f := range funds {
					if f.Duration() <= end+diff {
						continue
					}
					if highest == 0 {
						f.ratio[feature][end][diff] = 1
						continue
					}
					f.ratio[feature][end][diff] = f.features[feature][end][diff] / highest
				}
			}
		}
	}
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
