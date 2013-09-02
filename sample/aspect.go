package main

import (
	"github.com/jackvalmadre/golden"
	"math"
)

func OptimalAspect(aspects []float64) float64 {
	f := func(a float64) (y float64) {
		for _, xi := range aspects {
			y += 1 - 1/math.Max(xi/a, a/xi)
		}
		return
	}
	amin := aspects[argmin(aspects)]
	amax := aspects[argmax(aspects)]
	return golden.Search(f, amin, amax, 1e-6)
}

// Assumes that len(x) >= 1.
func argmin(x []float64) int {
	var arg int
	for i, xi := range x {
		if xi < x[arg] {
			arg = i
		}
	}
	return arg
}

// Assumes that len(x) >= 1.
func argmax(x []float64) int {
	var arg int
	for i, xi := range x {
		if xi > x[arg] {
			arg = i
		}
	}
	return arg
}
