package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
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
	cs := minimumOnPastBest(100000).choose(size, size/2)
	for _, c := range cs {
		fmt.Println(fs[c.i].name)
	}
	fmt.Println(math.Pow(profitability(cs, size/2, 0), 1.0/float64(size/2)/12.0))
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

func max(start, end int) (value float64, index int) {
	for fi := range fs {
		if table[fi][start][end] < value {
			continue
		}
		value = table[fi][start][end]
		index = fi
	}
	return value, index
}

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
					maxValidTimeTable[fi][start][end] = math.Pow(table[fi][start][end], 1.0/(float64(len(fs[fi].rend))/12.0))
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
		if money > fs[i].min {
			cs = append(cs, &chosen{i, fs[i].min})
		}
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
