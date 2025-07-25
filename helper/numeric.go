package helper

import (
	"math"

	"golang.org/x/exp/constraints"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Round(val float64, n int) float64 {
	base := math.Pow(10, float64(n))
	return math.Round(base*val) / base
}
