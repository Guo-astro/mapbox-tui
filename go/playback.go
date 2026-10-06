package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

var playbackSpeeds = [...]float64{1, 5, 20, 60, 120}

type ridePosition struct {
	Index     int
	Blend     float64
	World     XY
	Distance  float64
	Elevation *float64
	Time      *string
}

type playback struct {
	times   []float64
	stamps  []*time.Time
	Elapsed float64
	Index   int
	Blend   float64
	Playing bool
	speed   int
	Mode    string
}

func newPlayback(t *Track) *playback {
	p := &playback{times: make([]float64, len(t.Points)), stamps: make([]*time.Time, len(t.Points))}
	for i, point := range t.Points {
		if point.Time != nil {
			stamp, e := time.Parse(time.RFC3339Nano, *point.Time)
			if e == nil {
				p.stamps[i] = &stamp
			}
		}
	}
	elapsed, timedParts, estimatedParts := 0.0, 0, 0
	for _, s := range t.Segments {
		timed := s.End-s.Start >= 2
		for i := s.Start; i < s.End; i++ {
			if p.stamps[i] == nil || (i > s.Start && (p.stamps[i-1] == nil || !p.stamps[i].After(*p.stamps[i-1]))) {
				timed = false
			}
		}
		if s.End-s.Start >= 2 {
			if timed {
				timedParts++
			} else {
				estimatedParts++
			}
		}
		for i := s.Start; i < s.End; i++ {
			if i > s.Start {
				if timed {
					elapsed += secondsBetween(*p.stamps[i-1], *p.stamps[i])
				} else {
					elapsed += (t.Points[i].Distance - t.Points[i-1].Distance) / 6
				}
			}
			p.times[i] = elapsed
		}
	}
	p.Mode = "recorded time"
	if timedParts == 0 {
		p.Mode = "estimated 6 m/s"
	} else if estimatedParts > 0 {
		p.Mode = "mixed time/estimate"
	}
	if p.duration() == 0 {
		p.Mode = "no playable geometry"
	}
	return p
}
func secondsBetween(a, b time.Time) float64 {
	return float64(b.Unix()-a.Unix()) + float64(b.Nanosecond()-a.Nanosecond())/1e9
}
func (p *playback) duration() float64 {
	if len(p.times) == 0 {
		return 0
	}
	return p.times[len(p.times)-1]
}
func (p *playback) multiplier() float64 { return playbackSpeeds[p.speed] }
func (p *playback) ratio() float64 {
	if p.duration() == 0 {
		return 0
	}
	return p.Elapsed / p.duration()
}
func (p *playback) cycle(direction int) {
	p.speed = (p.speed + direction + len(playbackSpeeds)) % len(playbackSpeeds)
}
func (p *playback) seek(i int) {
	if i < 0 || i >= len(p.times) {
		return
	}
	p.Index = i
	p.Blend = 0
	p.Elapsed = p.times[i]
	p.Playing = false
}
func (p *playback) toggle() {
	if p.Playing {
		p.Playing = false
		return
	}
	if p.duration() == 0 {
		return
	}
	if p.Elapsed >= p.duration() {
		p.seek(0)
	}
	p.Playing = true
}
func (p *playback) advance(t *Track, seconds float64) bool {
	if !p.Playing || seconds <= 0 {
		return false
	}
	p.Elapsed = math.Min(p.duration(), p.Elapsed+seconds*p.multiplier())
	if p.Elapsed >= p.duration() {
		p.Index = len(p.times) - 1
		p.Blend = 0
		p.Playing = false
		return true
	}
	i := sort.Search(len(p.times), func(i int) bool { return p.times[i] > p.Elapsed }) - 1
	p.Index = max(0, i)
	p.Blend = 0
	if p.Index+1 < len(t.Points) && t.Points[p.Index].Segment == t.Points[p.Index+1].Segment && p.times[p.Index+1] > p.times[p.Index] {
		p.Blend = (p.Elapsed - p.times[p.Index]) / (p.times[p.Index+1] - p.times[p.Index])
	}
	return true
}
func (p *playback) nearest() int {
	if p.Blend > 0.5 {
		return p.Index + 1
	}
	return p.Index
}
func (p *playback) position(t *Track) *ridePosition {
	if p.Index < 0 || p.Index >= len(t.Points) {
		return nil
	}
	a := t.Points[p.Index]
	pos := &ridePosition{Index: p.Index, World: a.World, Distance: a.Distance, Elevation: a.Elevation, Time: a.Time}
	if p.Blend <= 0 || p.Index+1 >= len(t.Points) {
		return pos
	}
	b := t.Points[p.Index+1]
	if a.Segment != b.Segment {
		return pos
	}
	u := p.Blend
	pos.Blend = u
	pos.World = XY{a.World.X + (b.World.X-a.World.X)*u, a.World.Y + (b.World.Y-a.World.Y)*u}
	pos.Distance = a.Distance + (b.Distance-a.Distance)*u
	pos.Elevation = nil
	if a.Elevation != nil && b.Elevation != nil {
		e := *a.Elevation + (*b.Elevation-*a.Elevation)*u
		pos.Elevation = &e
	}
	pos.Time = nil
	if p.stamps[p.Index] != nil && p.stamps[p.Index+1] != nil && p.stamps[p.Index+1].After(*p.stamps[p.Index]) {
		stamp := p.stamps[p.Index]
		seconds := float64(stamp.Unix()) + float64(stamp.Nanosecond())/1e9 + secondsBetween(*stamp, *p.stamps[p.Index+1])*u
		whole := math.Floor(seconds)
		s := time.Unix(int64(whole), int64((seconds-whole)*1e9)).UTC().Format(time.RFC3339)
		pos.Time = &s
	}
	return pos
}
func clockLabel(seconds float64) string {
	n := int(seconds)
	if n >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", n/3600, n/60%60, n%60)
	}
	return fmt.Sprintf("%02d:%02d", n/60, n%60)
}
func (p *playback) status() string {
	state := "Paused"
	if p.Playing {
		state = "Playing"
	} else if p.duration() > 0 && p.Elapsed >= p.duration() {
		state = "Finished"
	}
	return fmt.Sprintf("%s %.0fx  %s / %s  %s", state, p.multiplier(), clockLabel(p.Elapsed), clockLabel(p.duration()), p.Mode)
}
func followPosition(c *Camera, world XY) bool {
	s := c.screen(world)
	w, h := float64(c.Width*2), float64(c.Height*4)
	if s.X < w*0.2 || s.X > w*0.8 || s.Y < h*0.2 || s.Y > h*0.8 {
		c.Center = world
		return true
	}
	return false
}
