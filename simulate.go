package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
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
		min, err := strconv.ParseFloat(strings.Replace(strings.Replace(fields[1], ".", "", -1), ",", ".", 1), 64)
		if err != nil {
			log.Fatal(err)
		}
		f := &fund{
			name: fields[0],
			min: min,
		}
		for i := 4; i < len(fields); i++ {
			read, err := strconv.ParseFloat(strings.Replace(fields[i], ",", ".", 1), 64)
			if err != nil {
				log.Fatal(err)
			}
			f.rend = append([]float64{1.0 + read / 100.0}, f.rend...)
		}
		fs = append(fs, f)
		if len(f.rend) > size {
			size = len(f.rend)
		}
	}
	fillTable()
	fillMaxValidTimeTable()
	fmt.Println("go")
	strats := []*namedStrategy{
		{"all", all{}},
		{"min115000", minimumOnPastBest(115000)},
		{"126,104,154", cnst([]int{126, 104, 154})},
	}
	for i := range fs {
		strats = append(strats, &namedStrategy{"cnst" + strconv.Itoa(i), cnst([]int{i})})
	}
	for i := 1; i <= 20; i++ {
		strats = append(strats, &namedStrategy{"top" + strconv.Itoa(i), top(i)})
		strats = append(strats, &namedStrategy{"topMin" + strconv.Itoa(i), topMin(i)})
		strats = append(strats, &namedStrategy{"random" + strconv.Itoa(i), random(i)})
		strats = append(strats, &namedStrategy{"bottom" + strconv.Itoa(i), bottom(i)})
	}
	for _, strat := range strats {
		fmt.Printf("%v\t%v\n", strat.name, evaluate(strat.strat))
	}
}

var (
	fs []*fund
	size int
)

type fund struct {
	name string
	min float64
	rend []float64
}

func fillTable() {
	table = make([][][]float64, len(fs))
	for fi, f := range fs {
		table[fi] = make([][]float64, size + 1)
		for start := size; start > 0; start-- {
			table[fi][start] = make([]float64, start)
			if start > len(f.rend) {
				continue
			}
			table[fi][start][start - 1] = f.rend[start - 1]
			for end := start - 2; end >= 0; end-- {
				table[fi][start][end] = table[fi][start][end + 1] * f.rend[end]
			}
		}
	}
}

var table [][][]float64

type strategy interface {
	choose(start, end int) []*chosen
}

type namedStrategy struct {
	name string
	strat strategy
}

type chosen struct {
	i int
	value float64
}

func profitability(cs []*chosen, start, end int) float64 {
	total := 0.0
	prof := 0.0
	for _, c := range cs {
		prof += c.value * table[c.i][start][end]
		total += c.value
	}
	return prof / total
}

func annual(value float64, months int) float64 {
	return math.Pow(value, 1.0/(float64(months)/12.0))
}

func max(start, end int) float64 {
	value := 0.0
	for fi := range fs {
		if table[fi][start][end] > value {
			value = table[fi][start][end]
		}
	}
	return value
}

func evaluate(strat strategy) float64 {
	var num float64
	var den int
	for t := 1; t < size; t++ {
		cs := strat.choose(size, t)
		p := profitability(cs, t, 0)
		if p == 0 {
			continue
		}
		num += (annual(max(t, 0), t) - annual(p, t)) * float64(t)
		den += t
	}
	return 100.0 * num / float64(den)
}

type minimumOnPastBest float64

func fillMaxValidTimeTable() {
	maxValidTimeTable = make([][][]float64, len(fs))
	for fi := range fs {
		maxValidTimeTable[fi] = make([][]float64, size + 1)
		for start := size; start > 0; start-- {
			maxValidTimeTable[fi][start] = make([]float64, start)
		}
		for end := 0; end < size; end++ {
			for start := end + 1; start <= size; start++ {
				if table[fi][start][end] != 0 {
					maxValidTimeTable[fi][start][end] = annual(table[fi][start][end], start - end)
					continue
				}
				if start > end + 1 {
					maxValidTimeTable[fi][start][end] = maxValidTimeTable[fi][start - 1][end]
				}
			}
		}
	}
}

var maxValidTimeTable [][][]float64

func (m minimumOnPastBest) choose(start, end int) []*chosen {
	is := make([]int, len(fs))
	for i := range is {
		is[i] = i
	}
	fisLessStart = start
	fisLessEnd = end
	sort.Sort(fis(is))
	var cs []*chosen
	money := float64(m)
	for _, i := range is {
		if money < fs[i].min || maxValidTimeTable[i][start][end] == 0 {
			break
		}
		cs = append(cs, &chosen{i, fs[i].min})
		money -= fs[i].min
	}
	return cs
}

type fis []int

func (f fis) Len() int {
	return len(f)
}

func (f fis) Less(i, j int) bool {
	return maxValidTimeTable[f[i]][fisLessStart][fisLessEnd] > maxValidTimeTable[f[j]][fisLessStart][fisLessEnd]
}


func (f fis) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

var (
	fisLessStart int
	fisLessEnd int
)

type cnst []int

func (c cnst) choose(start, end int) []*chosen {
	cs := make([]*chosen, len(c))
	for i, fi := range c {
		cs[i] = &chosen{int(fi), 1}
	}
	return cs
}

type top int

func (t top) choose(start, end int) []*chosen {
	is := make([]int, len(fs))
	for i := range is {
		is[i] = i
	}
	fisLessStart = start
	fisLessEnd = end
	sort.Sort(fis(is))
	var cs []*chosen
	for i := 0; i < len(fs) && len(cs) < int(t); i++ {
		if maxValidTimeTable[is[i]][start][end] > 0 {
			cs = append(cs, &chosen{is[i], 1})
		}
	}
	return cs
}

type topMin int

func (t topMin) choose(start, end int) []*chosen {
	is := make([]int, len(fs))
	for i := range is {
		is[i] = i
	}
	fisLessStart = start
	fisLessEnd = end
	sort.Sort(fis(is))
	var cs []*chosen
	for i := 0; i < int(t); i++ {
		cs = append(cs, &chosen{is[i], fs[is[i]].min})
	}
	return cs
}

type all struct{}

func (a all) choose(start, end int) []*chosen {
	var cs []*chosen
	for i := 0; i < len(fs); i++ {
		if maxValidTimeTable[i][start][end] > 0 {
			cs = append(cs, &chosen{i, 1})
		}
	}
	return cs
}

type random int

func (r random) choose(start, end int) []*chosen {
	var is []int
	for i := 0; i < len(fs); i++ {
		if maxValidTimeTable[i][start][end] > 0 {
			is = append(is, i)
		}
	}
	cs := make([]*chosen, int(r))
	for i := 0; i < int(r); i++ {
		cs[i] = &chosen{is[rand.Intn(len(is))], 1}
	}
	return cs
}

type bottom int

func (b bottom) choose(start, end int) []*chosen {
	is := make([]int, len(fs))
	for i := range is {
		is[i] = i
	}
	fisLessStart = start
	fisLessEnd = end
	sort.Sort(fis(is))
	var cs []*chosen
	for i := len(fs) - 1; i >= 0 && len(cs) < int(b); i-- {
		if maxValidTimeTable[is[i]][start][end] > 0 {
			cs = append(cs, &chosen{is[i], 1})
		}
	}
	return cs
}
