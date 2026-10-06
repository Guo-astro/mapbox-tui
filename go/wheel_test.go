package main

import (
	"math"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestWheelZoomUsesFineStepsAndLimitsBursts(t *testing.T) {
	now := time.Unix(100, 0)
	var wheel wheelZoom
	if delta := wheel.step(-1, now); math.Abs(delta+0.1) > 1e-12 {
		t.Fatalf("single step: %v", delta)
	}
	rest := 0.0
	for range 7 {
		rest += wheel.step(-1, now)
	}
	if math.Abs(rest+0.1) > 1e-12 {
		t.Fatalf("burst remainder: %v", rest)
	}
	if wheel.step(-1, now) != 0 {
		t.Fatal("exhausted burst still zoomed")
	}
	if wheel.step(-1, now.Add(20*time.Millisecond)) != 0 {
		t.Fatal("emitted a partial zoom step")
	}
	if delta := wheel.step(-1, now.Add(40*time.Millisecond)); math.Abs(delta+0.1) > 1e-12 {
		t.Fatalf("sustained scroll: %v", delta)
	}
	if delta := wheel.step(-1, now.Add(200*time.Millisecond)); math.Abs(delta+0.1) > 1e-12 {
		t.Fatalf("fresh scroll: %v", delta)
	}
}

func TestWheelBurstAndImmediateReversePreservePointerAnchor(t *testing.T) {
	now := time.Unix(100, 0)
	m := &viewer{camera: Camera{project(-0.1, 51.5), 14, 45, 2.4, 120, 40}}
	pointer := XY{33, 117}
	before := m.camera.world(pointer)
	for range 8 {
		m.zoomWheel(pointer, -1, now)
	}
	if math.Abs(m.camera.Zoom-13.8) > 1e-12 {
		t.Fatalf("burst jumped to %v", m.camera.Zoom)
	}
	m.zoomWheel(pointer, 1, now)
	if math.Abs(m.camera.Zoom-13.9) > 1e-12 {
		t.Fatal("direction reversal delayed")
	}
	after := m.camera.world(pointer)
	if math.Hypot(before.X-after.X, before.Y-after.Y) > 1e-12 {
		t.Fatal("wheel burst moved anchor")
	}
}

func TestHorizontalWheelDoesNotZoomGoViewer(t *testing.T) {
	track, e := loadTrack("../fixtures/example.gpx")
	if e != nil {
		t.Fatal(e)
	}
	c := fit(track, 88, 34, 2)
	m := &viewer{track: track, camera: c}
	for _, button := range []tea.MouseButton{tea.MouseWheelLeft, tea.MouseWheelRight} {
		m.Update(tea.MouseWheelMsg{X: 20, Y: 10, Button: button})
	}
	if m.camera != c || m.dirty {
		t.Fatal("horizontal scrolling zoomed map")
	}
}
