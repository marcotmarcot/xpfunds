package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	var size int
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
	m, i := max(fs, size, 0)
	fmt.Println(m, i)
}

var fs []*fund

type fund struct {
	name string
	min float64
	rend []float64
}

type pCacheI struct {
	fi, start, end int
}

func max(fs []*fund, start, end int) (value float64, index int) {
	for fi := range fs {
		p := profitability(fi, start, end)
		if p < value {
			continue
		}
		value = p
		index = fi
	}
	return value, index
}

func profitability(fi int, start, end int) float64 {
	pci := pCacheI{fi, start, end}
	v, ok := pCache[pci]
	if ok {
		return v
	}
	p := 1.0
	if start > len(fs[fi].rend) {
		pCache[pci] = 0.0
		return 0.0
	}
	for i := start - 1; i >= end; i-- {
		p *= fs[fi].rend[i]
	}
	pCache[pci] = math.Pow(p, 1.0/(float64(start - end)/12.0))
	return pCache[pci]
}

var pCache = make(map[pCacheI]float64)

type strategy interface {
	choose(fs []*fund, start, end int) []int
}

type current struct {}

func (c *current) choose(fs []*fund, start, end int) []int {
	return nil
}

type fis []int

func (f fis) Len() int {
	return len(f)
}

func (f fis) Less(i, j int) bool {
	return profitability(f[i], fisLessStart, fisLessEnd) > profitability(f[j], fisLessStart, fisLessEnd)
}

func (f fis) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

var (
	fisLessStart int
	fisLessEnd int
)
