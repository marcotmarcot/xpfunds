package weight

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewWeights(fields []string, n int) []map[string]float64 {
	var weights []map[string]float64
	for i := 0; i < n; i++ {
		weight := make(map[string]float64)
		for _, field := range fields {
			weight[field] = rand.Float64()*2 - 1
		}
		weights = append(weights, weight)
	}
}

func Reproduce(a, b map[string]float64, n int) []map[string]float64 {
	weights := []map[string]float64{a, b}
	for i := 0; i < n-2; i++ {
		weight := make(map[string]float64)
		for field, avalue := range a {
			weight[field] = (avalue+b[field])/2 + rand.NormFloat()
			for weight[field] < -1 {
				weight[field] += 2
			}
			for weight[field] > 1 {
				weight[field] -= 2
			}
		}
		weights = append(weights, weight)
	}
}
