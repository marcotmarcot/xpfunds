package simulate

import (
	"math"
	"testing"
	"xpfunds"
)

func TestPerformance(t *testing.T) {
	tests := []struct {
		name     string
		monthlys [][]float64
		numFunds int
		weight   float64
		want     float64
	}{{
		"clearBest",
		[][]float64{
			{1, 1.1},
			{2, 1.2},
		},
		1,
		1,
		1,
	}, {
		"reverse",
		[][]float64{
			{2, 1.1},
			{1, 1.2},
		},
		1,
		1,
		0.5,
	}, {
		"twoFunds",
		[][]float64{
			{3, 1.1},
			{1, 1.2},
			{2, 1.3},
		},
		2,
		1,
		0.5,
	}, {
		"clearBestNegative",
		[][]float64{
			{1, 1.1},
			{2, 1.2},
		},
		1,
		-1,
		0.5,
	}, {
		"reverseNegative",
		[][]float64{
			{2, 1.1},
			{1, 1.2},
		},
		1,
		-1,
		1,
	}, {
		"twoFundsNegative",
		[][]float64{
			{3, 1.1},
			{1, 1.2},
			{2, 1.3},
		},
		2,
		-1,
		0.6666666666666666,
	}}
	for _, test := range tests {
		var funds []*xpfunds.Fund
		for _, monthly := range test.monthlys {
			funds = append(funds, xpfunds.NewFund(monthly))
		}
		xpfunds.SetRatio(funds)
		if perf, ok := performance(funds, test.numFunds, NewWeighted(0, []float64{test.weight, 1, 1}), 1); !ok {
			t.Errorf("%v: Not ok", test.name)
		} else if got, want := perf, test.want; !eq(got, want) {
			t.Errorf("%v: got: %v, want: %v", test.name, got, want)
		}
	}
}

func TestMedianPerformance(t *testing.T) {
	tests := []struct {
		name     string
		monthlys [][]float64
		numFunds int
		weight   float64
		want     float64
	}{{
		"clearBest",
		[][]float64{
			{1, 1.1},
			{2, 1.2},
		},
		1,
		1,
		1,
	}, {
		"reverse",
		[][]float64{
			{2, 1.1},
			{1, 1.2},
		},
		1,
		1,
		0.5,
	}, {
		"twoFunds",
		[][]float64{
			{3, 1.1},
			{1, 1.2},
			{2, 1.3},
		},
		2,
		1,
		0.5,
	}, {
		"clearBestNegative",
		[][]float64{
			{1, 1.1},
			{2, 1.2},
		},
		1,
		-1,
		0.5,
	}, {
		"reverseNegative",
		[][]float64{
			{2, 1.1},
			{1, 1.2},
		},
		1,
		-1,
		1,
	}, {
		"twoFundsNegative",
		[][]float64{
			{3, 1.1},
			{1, 1.2},
			{2, 1.3},
		},
		2,
		-1,
		0.6666666666666666,
	}, {
		"3clearBest",
		[][]float64{
			{1, 1, 1.1},
			{1.5, 2, 1.2},
		},
		1,
		1,
		1,
	}, {
		"3reverse",
		[][]float64{
			{1.5, 2, 1.1},
			{1, 1, 1.2},
		},
		1,
		1,
		0.6666666666666663,
	}, {
		"3twoFunds",
		[][]float64{
			{2.5, 3, 1.1},
			{1, 1, 1.2},
			{2, 2, 1.3},
		},
		2,
		1,
		0.6166666666666667,
	}, {
		"3clearBestNegative",
		[][]float64{
			{1, 1, 1.1},
			{1.5, 2, 1.2},
		},
		1,
		-1,
		0.5,
	}, {
		"3reverseNegative",
		[][]float64{
			{1.5, 2, 1.1},
			{1, 1, 1.2},
		},
		1,
		-1,
		0.8333333333333333,
	}, {
		"3twoFundsNegative",
		[][]float64{
			{2.5, 3, 1.1},
			{1, 1, 1.2},
			{2, 2, 1.3},
		},
		2,
		-1,
		0.5833333333333333,
	}}
	for _, test := range tests {
		var funds []*xpfunds.Fund
		for _, monthly := range test.monthlys {
			funds = append(funds, xpfunds.NewFund(monthly))
		}
		xpfunds.SetRatio(funds)
		if got, want := MedianPerformance(funds, 3, test.numFunds, NewWeighted(0, []float64{test.weight, 1, 1})), test.want; !eq(got, want) {
			t.Errorf("%v: got: %v, want: %v", test.name, got, want)
		}
	}
}

func eq(a, b float64) bool {
	return math.Abs(a-b) < 0.000001
}
