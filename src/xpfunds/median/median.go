package median

import "sort"

func Median(s []float64) float64 {
	if len(s) == 0 {
		return -1
	}
	if len(s) == 1 {
		return s[0]
	}
	sort.Float64s(s)
	return MedianFromSorted(s)
}

func MedianFromSorted(s []float64) float64 {
	if len(s)%2 == 1 {
		return s[len(s)/2]
	}
	return (s[len(s)/2] + s[len(s)/2-1]) / 2
}
