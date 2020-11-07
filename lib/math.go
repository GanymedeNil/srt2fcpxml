package lib

import "math"

func Round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10

}

type Floater struct {
	Accuracy float64
}

func (f Floater) IsEqual(a, b float64) bool {
	return math.Abs(a-b) < f.Accuracy
}
func (f Floater) Bccomp(a, b float64) int8 {
	if math.Abs(a-b) < f.Accuracy {
		return 0
	}
	if math.Max(a, b) == a {
		return 1
	} else {
		return -1
	}
}
