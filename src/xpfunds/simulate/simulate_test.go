package simulate

import "testing"

func TestBestChoose(t *testing.T) {
	funds = []*fund{
		{[][]float64{{1.02}}},
		{[][]float64{{1.04}}},
		{[][]float64{{1.03}}},
		{[][]float64{{1.01}}}}
	choice := best{}.choose(2, 0, 0, 0)
	if want, got := 2, len(choice); want != got {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if want, got := 1, choice[0]; want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
	if want, got := 2, choice[1]; want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}

func TestBestChooseEqual(t *testing.T) {
	funds = []*fund{
		{[][]float64{{1}}},
		{[][]float64{{1}}}}
	choice := best{}.choose(2, 0, 0, 0)
	if want, got := 2, len(choice); want != got {
		t.Fatalf("want=%v, got=%v", want, got)
	}
	if want, got := 0, choice[0]; want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
	if want, got := 1, choice[1]; want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}

func TestRentability(t *testing.T) {
	r := fund{[][]float64{{1}, {1, 1}}}.rentability(1, 1)
	if r == nil {
		t.Fatalf("r == nil")
	}
	if want, got := 1.0, *r; want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}

func TestLossBest(t *testing.T) {
	funds = []*fund{
		{[][]float64{{1}, {1, 1}}},
		{[][]float64{{1}, {1, 2}}}}
	optimum = &fund{[][]float64{{1}, {1, 2}}}
	l := loss(best{}, 2, 0, 0, 1, 0)
	if l == nil {
		t.Fatalf("l == nil")
	}
	if want, got := 0.75, *l; want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}
