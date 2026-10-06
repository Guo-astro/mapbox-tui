package main

import (
	"math"
	"time"
)

const zoomStep = 0.1

type wheelZoom struct {
	last      time.Time
	direction int
	budget    float64
}

func (w *wheelZoom) step(direction int, now time.Time) float64 {
	if w.last.IsZero() || direction != w.direction {
		w.budget = 2 * zoomStep
	} else {
		w.budget = math.Min(2*zoomStep, w.budget+now.Sub(w.last).Seconds()*2.5)
	}
	w.last, w.direction = now, direction
	if w.budget+1e-12 < zoomStep {
		return 0
	}
	w.budget = math.Max(0, w.budget-zoomStep)
	return float64(direction) * zoomStep
}
