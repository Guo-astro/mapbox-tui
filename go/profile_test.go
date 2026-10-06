package main

import "testing"

func TestElevationProfileKeepsMissingMeasurementsAndPartGaps(t *testing.T) {
	e0, e10, e20 := 0.0, 10.0, 20.0
	missing := &Track{Distance: 40, Segments: []Segment{{Start: 0, End: 5}}, Points: []Point{{Distance: 0, Elevation: &e0}, {Distance: 10, Elevation: &e10}, {Distance: 20}, {Distance: 30, Elevation: &e20}, {Distance: 40, Elevation: &e0}}}
	p := newProfile(missing, 41, 5)
	for y := 0; y < 5; y++ {
		if p.Grid.Cells[y*41+20].Bits != 0 {
			t.Fatal("profile filled missing elevation")
		}
	}
	parts := &Track{Distance: 20, Segments: []Segment{{Start: 0, End: 2}, {Start: 2, End: 4}}, Points: []Point{{Distance: 0, Elevation: &e0}, {Distance: 10, Elevation: &e20}, {Distance: 10, Elevation: &e0}, {Distance: 20, Elevation: &e20}}}
	p = newProfile(parts, 41, 5)
	if p.Grid.Cells[2*41+20].Bits != 0 {
		t.Fatal("profile connected part boundary vertically")
	}
	flat := &Track{Distance: 10, Segments: []Segment{{Start: 0, End: 2}}, Points: []Point{{Elevation: &e0}, {Distance: 10, Elevation: &e0}}}
	p = newProfile(flat, 41, 5)
	if !p.Available || p.Grid.Cells[2*41+20].Bits == 0 {
		t.Fatal("flat profile failed")
	}
}
