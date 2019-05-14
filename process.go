package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

var numMonths = 2

func main() {
	ipca := newIndex("ipca")
	get, err := ioutil.ReadFile("get.tsv")
	if err != nil {
		log.Fatal(err)
	}
	var funds []*fund
	for _, line := range strings.Split(string(get), "\n") {
		f := newFund(line, ipca)
		if f == nil {
			break
		}
		funds = append(funds, f)
	}
	for _, fi := range fields {
		fmt.Printf("%s\t", fi.name)
	}
	fmt.Printf("\n")
	for _, f := range funds {
		for _, fi := range fields {
			fmt.Printf("%s\t", fi.value(f))
		}
		fmt.Printf("\n")
	}
}

type fund struct {
	fundName string

	// The minimum value for investment. As we don't use it in
	// processing so we leave it as string in the same way it came
	// from get.go.
	min string

	// The number of days we need to wait to get the money in an
	// withdraw.
	days int

	// Whether this fund is active or not. Not used in processing
	// so we keep the string from get.go.
	fundActive string

	// The monthly gain for this fund, from the more recent to the
	// less recent.
	raw []float64

	// The greatest fall this fund had.
	greatFall float64

	// The number of months the greatest fall took.
	greatFallLen int
}

// line is in the format produced by get.go.
func newFund(line string, ix *index) *fund {
	f := &fund{}
	fields := strings.Split(strings.Trim(line, "\n"), "\t")
	if len(fields) < 2 {
		return nil
	}
	f.fundName = fields[0]
	f.min = fields[1]
	f.setDays(fields)
	f.fundActive = fields[4]
	f.setRaw(fields, ix)
	f.setGreatFall()
	return f
}

func (f *fund) setDays(fields []string) {
	cot, err := strconv.Atoi(fields[2])
	if err != nil {
		log.Fatal(err)
	}
	liq, err := strconv.Atoi(fields[3])
	if err != nil {
		log.Fatal(err)
	}
	f.days = cot + liq
}

func (f *fund) setRaw(fields []string, ix *index) {
	for i := 5; i < len(fields); i++ {
		v, err := strconv.ParseFloat(strings.Replace(fields[i], ",", ".", 1), 64)
		if err != nil {
			log.Fatal(err)
		}
		f.raw = append(f.raw, v-ix.raw[i])
	}
}

func (f *fund) setGreatFall() {
	f.greatFall = 1.0
	curr := 1.0
	currLen := 0
	for _, v := range f.raw {
		a := absolute(v)
		curr *= a
		currLen++
		if a < curr {
			curr = a
			currLen = 1
		}
		if curr < f.greatFall {
			f.greatFall = curr
			f.greatFallLen = currLen
		}
	}
}

func (f *fund) name() string {
	return f.fundName
}

func (f *fund) minimum() string {
	return f.min
}

func (f *fund) daysForWithdraw() string {
	return strconv.Itoa(f.days)
}

func (f *fund) age() string {
	return strconv.Itoa(len(f.raw))
}

func (f *fund) stddev() string {
	sum := 0.0
	for _, v := range f.raw {
		sum += v
	}
	avg := sum / float64(len(f.raw))
	sumDiffs := 0.0
	for _, v := range f.raw {
		diff := v - avg
		sumDiffs += diff * diff
	}
	return formatFloat(math.Sqrt(sumDiffs / float64(len(f.raw))))
}

func (f *fund) negativeMonths() string {
	n := 0
	for _, v := range f.raw {
		if v < 0 {
			n++
		}
	}
	return strconv.Itoa(n)
}

func (f *fund) greatestFall() string {
	return relative(f.greatFall)
}

func (f *fund) greatestFallLen() string {
	return strconv.Itoa(f.greatFallLen)
}

func (f *fund) yearly() string {
	total := 1.0
	for _, v := range f.raw {
		total *= absolute(v)
	}
	return relative(math.Pow(total, 12.0/float64(len(f.raw))))
}

func (f *fund) lastMonths() string {
	total := 1.0
	n := 0
	for i, v := range f.raw {
		if i >= numMonths {
			break
		}
		total *= absolute(v)
		n++
	}
	return relative(math.Pow(total, 12.0/float64(n)))
}

func (f *fund) active() string {
	return f.fundActive
}

func (f *fund) median() string {
	if len(f.raw) == 0 {
		return ""
	}
	sorted := make([]float64, len(f.raw))
	copy(sorted, f.raw)
	sort.Float64s(sorted)
	return relative(math.Pow(absolute(sorted[len(sorted)/2]), 12.0))
}

type field struct {
	name  string
	value func(f *fund) string
}

var fields = []field{
	{"Nome", (*fund).name},
	{"Mínimo", (*fund).minimum},
	{"Dias para resgate", (*fund).daysForWithdraw},
	{"Idade em meses", (*fund).age},
	{"Desvio padrão", (*fund).stddev},
	{"Meses negativos", (*fund).negativeMonths},
	{"Maior queda", (*fund).greatestFall},
	{"Número de meses da maior queda", (*fund).greatestFallLen},
	{"Rentabilidade anualizada", (*fund).yearly},
	{"Rentabilidade nos últimos meses", (*fund).lastMonths},
	{"Mediana da rentabilidade", (*fund).median},
	{"Investível", (*fund).active},
}

type index struct {
	raw []float64
}

func newIndex(name string) *index {
	file, err := ioutil.ReadFile(name + ".tsv")
	if err != nil {
		log.Fatal(err)
	}
	ix := &index{}
	for _, f := range strings.Split(string(file), "\n") {
		v, err := strconv.ParseFloat(strings.Replace(f, ",", ".", 1), 64)
		if err != nil {
			break
		}
		ix.raw = append(ix.raw, v)
	}
	return ix
}

func formatFloat(f float64) string {
	return strings.Replace(strconv.FormatFloat(f, 'f', 2, 64), ".", ",", 1)
}

// Converts a gain in relative percentage to absolute value. For
// instance, absolute(1) == 1.01.
func absolute(f float64) float64 {
	return 1.0 + f/100.0
}

// The reverse function of absolute, that also transforms it to a
// string.
func relative(f float64) string {
	return formatFloat((f - 1.0) * 100.0) + "%"
}
