package binarysearch

func InsertInSorted(s []float64, e float64) []float64 {
	i := UpperBound(s, e)
	return append(s[:i], append([]float64{e}, s[i:]...)...)
}

func UpperBound(s []float64, e float64) int {
	return upperBound(s, 0, len(s), e)
}

func upperBound(s []float64, begin, end int, e float64) int {
	if end-begin == 0 {
		return begin
	}
	pivot := (begin + end) / 2
	if s[pivot] == e {
		return pivot
	} else if s[pivot] < e {
		return upperBound(s, pivot+1, end, e)
	}
	return upperBound(s, begin, pivot, e)
}

func MedianFromSorted(s []float64) float64 {
	if len(s)%2 == 1 {
		return s[len(s)/2]
	}
	return (s[len(s)/2] + s[len(s)/2-1]) / 2
}
