package main

import (
	"time"
	"math"
    glfw "github.com/go-gl/glfw3"
)

type sincos struct {
	currPhase []float64
	currSpeed float64 // rad/s
	lastUpdated time.Time
} 


func (s *sincos)didScroll(w *glfw.Window, xoff float64, yoff float64) {
	s.currSpeed *= math.Pow(1.01,yoff)
	if s.currSpeed < 0.01 {
		s.currSpeed = 0.01
	}
}

func (s *sincos)Begin(w *glfw.Window, num int) error {
	w.SetScrollCallback(s.didScroll)
	s.currPhase = make([]float64, num)
	s.currPhase[num - 1] = math.Pi;
	s.currSpeed = 1.0;
	return nil
}

func (s *sincos)Update(intensity []float32) error {
	now := time.Now()
	timeDelta := now.Sub(s.lastUpdated)
	for i := range s.currPhase {
		s.currPhase[i] += timeDelta.Seconds() * s.currSpeed
		s.currPhase[i] = math.Mod(s.currPhase[i], math.Pi*2)
		intensity[i] = float32((math.Sin(s.currPhase[i]) + 1)/2)
	}
	s.lastUpdated = now
	return nil
}


