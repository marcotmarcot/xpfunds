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
	strats := make([]strategy, 50)
	for i := 0; i < 50; i++ {
		strats[i] = newGeneticStrategy()
	}
	for i := 0; i < 50; i++ {
		strats = top10([]strategy(strats))
		a := strats[0].(geneticStrategy)
		sum := 0.0
		for _, v := range a {
			sum += v
		}
		m := make(map[int]float64)
		for i, v := range a {
			p := 100 * v / sum
			if p > 1 {
				m[i] = p
			}
		}
		fmt.Println(evaluate(strats[0]), m)
		strats = nextGen(strats)
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
	max = make([][]float64, size + 1)
	for start := size; start > 0; start-- {
		max[start] = make([]float64, start)
		for end := start - 1; end >= 0; end-- {
			m := 0.0
			for fi := range fs {
				if table[fi][start][end] > m {
					m = table[fi][start][end]
				}
			}
			max[start][end] = m
		}
	}
}

var (
	table [][][]float64
	max [][]float64
)

type strategy interface {
	choose(start, end int) []*chosen
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

func evaluate(strat strategy) float64 {
	var num float64
	var den int
	for t := size - 1; t > 0; t-- {
		cs := strat.choose(size, t)
		p := profitability(cs, t, 0)
		if p == 0 {
			continue
		}
		num += (annual(max[t][0], t) - annual(p, t)) * float64(t)
		den += t
	}
	return 100.0 * num / float64(den)
}

func top10(strats []strategy) []strategy {
	sls := make([]*stratWithLoss, len(strats))
	for i, s := range strats {
		sls[i] = &stratWithLoss{s, evaluate(s)}
	}
	sort.Sort(stratsWithLoss(sls))
	top := make([]strategy, 20)
	for i := 0; i < 20; i++ {
		top[i] = sls[i].strat
	}
	return top
}

type stratWithLoss struct {
	strat strategy
	loss float64
}

type stratsWithLoss []*stratWithLoss

func (s stratsWithLoss) Len() int {
	return len(s)
}

func (s stratsWithLoss) Less(i, j int) bool {
	return s[i].loss < s[j].loss
}

func (s stratsWithLoss) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type geneticStrategy []float64

func (g geneticStrategy) choose(start, end int) []*chosen {
	cs := make([]*chosen, len(g))
	for i, v := range g {
		cs[i] = &chosen{i, v}
	}
	return cs
}

func newGeneticStrategy() strategy {
	g := make([]float64, len(fs))
	n := rand.Intn(len(fs))
	for i := 0; i < n; i++ {
		g[rand.Intn(len(fs))] = 1.0
	}
	return geneticStrategy(g)
}

func breed(p1, p2 strategy) strategy {
	a := p1.(geneticStrategy)
	b := p2.(geneticStrategy)
	g := make([]float64, len(fs))
	for i := range g {
		g[i] = (a[i] + b[i]) / 2.0
	}
	return geneticStrategy(g)
}

func mutate(p1 strategy) strategy {
	a := p1.(geneticStrategy)
	g := make([]float64, len(fs))
	copy(g, a)
	n := rand.Intn(len(fs) / 4)
	for i := 0; i < n; i++ {
		g[rand.Intn(len(fs))] *= 2
		g[rand.Intn(len(fs))] /= 2
		g[rand.Intn(len(fs))] = 0
		g[rand.Intn(len(fs))] = 1
	}
	return geneticStrategy(g)
}

func nextGen(strats []strategy) []strategy {
	var new []strategy
	for _, a := range strats {
		for _, b := range strats {
			new = append(new, breed(a, b))
		}
	}
	strats = append(strats, new...)
	new = []strategy{}
	for _, s := range strats {
		new = append(new, mutate(s))
	}
	strats = append(strats, new...)
	return strats
}
