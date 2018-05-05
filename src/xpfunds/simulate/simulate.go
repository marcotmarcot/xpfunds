package simulate

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

func Main() {
	rand.Seed(42)
	r := bufio.NewReader(os.Stdin)
	for true {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		fields := strings.Split(strings.Trim(line, "\n"), "\t")
		if fields == nil {
			break
		}
		funds = append(funds, newFund(fields))
	}
	newOptimum()
	strategies := []strategy{best{}, worst{}, random{}}
	// for i := range funds {
	// 	strategies = append(strategies, single{i})
	// }
	for num_funds := 8; num_funds <= 8; num_funds++ {
		for min_time := 2; min_time <= 2; min_time++ {
			for num_months := -1; num_months <= 20; num_months += 1 {
				strategies = append(strategies, best{num_funds, min_time, num_months}, worst{num_funds, min_time, num_months}, random{num_funds, min_time, num_months})
			}
		}
	}
	for _, s := range strategies {
		l, d, f, mf := meanPerformance(s, len(optimum.rentabilities), 0)
		fmt.Println(s.name(), l, d, f, mf)
	}
}

var optimum *fund
var funds []*fund

func newFund(fields []string) *fund {
	f := &fund{}
	var raw []float64
	for i := 4; i < len(fields); i++ {
		v, err := strconv.ParseFloat(strings.Replace(fields[i], ",", ".", 1), 64)
		if err != nil {
			log.Fatal(err)
		}
		raw = append(raw, v/100+1)
	}
	f.rentabilities = make([][]float64, len(raw))
	for start := range raw {
		f.rentabilities[start] = make([]float64, start+1)
		f.rentabilities[start][start] = raw[start]
		for end := start - 1; end >= 0; end-- {
			f.rentabilities[start][end] = f.rentabilities[start][end+1] * raw[end]
		}
	}
	return f
}

type fund struct {
	rentabilities [][]float64
}

func (f fund) rentability(start, end int) *float64 {
	if start >= len(f.rentabilities) {
		return nil
	}
	return &f.rentabilities[start][end]
}

func (f fund) average(start, end int) *float64 {
	if end >= len(f.rentabilities) {
		return nil
	}
	if start >= len(f.rentabilities) {
		start = len(f.rentabilities) - 1
	}
	r := annual(f.rentabilities[start][end], start, end)
	return &r
}

func annual(r float64, start, end int) float64 {
	return math.Pow(r, 12.0 / float64(start - end + 1))
}

func newOptimum() {
	optimum = &fund{}
	duration := 0
	for _, f := range funds {
		if len(f.rentabilities) > duration {
			duration = len(f.rentabilities)
		}
	}
	optimum.rentabilities = make([][]float64, duration)
	for start := 0; start < duration; start++ {
		optimum.rentabilities[start] = make([]float64, start+1)
		for end := 0; end <= start; end++ {
			for _, f := range funds {
				r := f.rentability(start, end)
				if r != nil && *r > optimum.rentabilities[start][end] {
					optimum.rentabilities[start][end] = *r
				}
			}
		}
	}
}

func meanPerformance(s strategy, start, end int) (loss float64, diff float64, meanFuture float64, minFuture float64) {
	var losses []float64
	var diffs []float64
	var futures []float64
	for time := 0; time < len(optimum.rentabilities) - 1; time++ {
		l, d, f := performance(s, start, end, time)
		if l != nil {
			losses = append(losses, *l)
		}
		if d != nil {
			diffs = append(diffs, *d)
		}
		if f != nil {
			futures = append(futures, *f)
		}
	}
	mf := 0.0
	if len(futures) > 0 {
		mf = futures[0]
	}
	return mean(losses), mean(diffs), mean(futures), mf
}

func mean(s []float64) float64 {
	if len(s) == 0 {
		return -1
	}
	sort.Float64s(s)
	m := len(s) / 2
	if len(s)%2 == 0 {
		return s[m]
	}
	return (s[m] + s[m]) / 2
}

func performance(s strategy, start, end, time int) (loss *float64, diff *float64, future *float64) {
	fis := s.choose(start, time)
	nfuture := 0.0
	dfuture := 0.0
	var pasts []float64
	for _, fi := range fis {
		future := funds[fi].rentability(time + 1, end)
		if future != nil {
			nfuture += *future
			dfuture++
		}
		past := funds[fi].average(start, time)
		if past != nil {
			pasts = append(pasts, *past)
		}
	}
	var l *float64
	var f *float64
	if dfuture > 0 {
		vf := annual(nfuture / dfuture, time + 1, end)
		f = &vf
		vl := annual(nfuture / dfuture, time + 1, end) / annual(optimum.rentabilities[time + 1][end], time + 1, end)
		l = &vl
	}
	var d *float64
	if len(pasts) != 0 && dfuture > 0 {
		vd := annual(nfuture / dfuture, time + 1, end) / mean(pasts)
		d = &vd
	}
	return l, d, f
}

type strategy interface {
	name() string
	choose(start, end int) []int
}

type best struct {
	num_funds int
	min_time int
	num_months int
}

func (b best) name() string {
	return "Best" + strconv.Itoa(b.num_funds) + "," + strconv.Itoa(b.min_time) + "," + strconv.Itoa(b.num_months)
}

func (b best) choose(start, end int) []int {
	return sortAndPick(b.num_funds, b.min_time, b.num_months, start, end, false, false)
}

type worst struct {
	num_funds int
	min_time int
	num_months int
}

func (w worst) name() string {
	return "Worst" + strconv.Itoa(w.num_funds) + "," + strconv.Itoa(w.min_time) + "," + strconv.Itoa(w.num_months)
}

func (w worst) choose(start, end int) []int {
	return sortAndPick(w.num_funds, w.min_time, w.num_months, start, end, true, false)
}

type random struct {
	num_funds int
	min_time int
	num_months int
}

func (r random) name() string {
	return "Random" + strconv.Itoa(r.num_funds) + "," + strconv.Itoa(r.min_time) + "," + strconv.Itoa(r.num_months)
}

func (r random) choose(start, end int) []int {
	return sortAndPick(r.num_funds, r.min_time, r.num_months, start, end, false, true)
}

type single struct {
	fund int
}

func (s single) name() string {
	return strconv.Itoa(s.fund)
}

func (s single) choose(start, end int) []int {
	return []int{s.fund}
}


func sortAndPick(num_funds, min_time, num_months, start, end int, worst, random bool) []int {
	var fis []int
	for fi := 0; fi < len(funds); fi++ {
		fis = append(fis, fi)
	}
	sort.Sort(byAverage{num_months, fis, start, end, worst, random})
	var choice []int
	for _, fi := range fis {
		if funds[fi].average(start, end) == nil || len(funds[fi].rentabilities) - end < min_time {
			continue
		}
		choice = append(choice, fi)
		if len(choice) >= num_funds {
			break
		}
	}
	return choice
}

type byAverage struct {
	num_months int
	fis []int
	start, end int
	worst bool
	random bool
}

func (b byAverage) Len() int {
	return len(b.fis)
}

func (b byAverage) Swap(i, j int) {
	b.fis[i], b.fis[j] = b.fis[j], b.fis[i]
}

func (b byAverage) Less(i, j int) bool {
	if b.num_months != -1 {
		b.start = b.end + b.num_months
	}
	ri := funds[b.fis[i]].average(b.start, b.end)
	rj := funds[b.fis[j]].average(b.start, b.end)
	if b.random {
		vri := rand.Float64()
		ri = &vri
		vrj := rand.Float64()
		rj = &vrj
	}
	if ri == nil {
		return false
	}
	if rj == nil {
		return true
	}
	if b.worst {
		return *rj > *ri
	}
	return *ri > *rj
}
