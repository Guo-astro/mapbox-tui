package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

func run() error {
	if len(os.Args) < 2 || (len(os.Args) == 2 && os.Args[1] == "open") {
		provider := "demo"
		if accessToken() != "" {
			provider = "mapbox"
		}
		return openViewer("", Args{Provider: provider})
	}
	mode := os.Args[1]
	if mode == "mcp" {
		return serveMCP()
	}
	if len(os.Args) < 3 {
		return fmt.Errorf("provide a GPX path")
	}
	path := os.Args[2]
	fs := flag.NewFlagSet(mode, flag.ContinueOnError)
	w := fs.Int("width", 120, "map width in cells")
	h := fs.Int("height", 40, "map height in cells")
	aspect := fs.Float64("cell-aspect", 2, "cell height/width")
	ascii := fs.Bool("ascii", false, "ASCII rendering")
	demoMode := fs.Bool("demo", false, "force synthetic basemap")
	noAnimation := fs.Bool("no-animation", false, "disable camera transitions")
	kind := fs.String("provider", "", "demo, mapbox, none")
	jsonMode := fs.Bool("json", false, "JSON render result")
	index := fs.Int("point-index", 0, "original point index")
	distance := fs.Float64("distance-m", 0, "nearest point at distance")
	lon := fs.Float64("center-lon", 0, "longitude")
	lat := fs.Float64("center-lat", 0, "latitude")
	zoom := fs.Float64("zoom", 0, "zoom")
	bearing := fs.Float64("bearing", 0, "bearing degrees")
	frames := fs.Int("frames", 100, "benchmark frames")
	if e := fs.Parse(os.Args[3:]); e != nil {
		return e
	}
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	if *kind == "" {
		if accessToken() != "" {
			*kind = "mapbox"
		} else {
			*kind = "demo"
		}
	}
	if *demoMode {
		*kind = "demo"
	}
	if *kind != "demo" && *kind != "mapbox" && *kind != "none" {
		return fmt.Errorf("unknown provider")
	}
	a := Args{Width: w, Height: h, Aspect: aspect, ASCII: *ascii, Provider: *kind, NoAnimation: *noAnimation}
	if set["point-index"] {
		a.PointIndex = index
	}
	if set["distance-m"] {
		a.Distance = distance
	}
	if set["center-lon"] {
		a.Lon = lon
	}
	if set["center-lat"] {
		a.Lat = lat
	}
	if set["zoom"] {
		a.Zoom = zoom
	}
	if set["bearing"] {
		a.Bearing = bearing
	}
	if mode == "open" {
		_, e := viewport(&Track{Bounds: [4]float64{0, 0, 0, 0}}, a)
		if e != nil {
			return e
		}
		return openViewer(path, a)
	}
	start := time.Now()
	t, e := loadTrack(path)
	if e != nil {
		return e
	}
	loadMs := float64(time.Since(start).Microseconds()) / 1000
	var result any
	switch mode {
	case "summary":
		result = t.summary()
	case "inspect":
		if a.PointIndex == nil && a.Distance == nil {
			a.PointIndex = index
		}
		result, e = t.inspect(a.PointIndex, a.Distance)
	case "render":
		result, e = renderResult(context.Background(), t, a, newProvider(*kind))
		if e == nil && !*jsonMode {
			r := result.(map[string]any)
			fmt.Println(r["text"])
			fmt.Println(attribution(*kind))
			return nil
		}
	case "bench":
		if *frames < 1 {
			return fmt.Errorf("frames must be positive")
		}
		c, err := viewport(t, a)
		if err != nil {
			return err
		}
		features := demo(t)
		samples := []float64{}
		start = time.Now()
		for i := 0; i < *frames; i++ {
			now := time.Now()
			c.pan(0.2, 0)
			g := render(t, c, features, -1, *ascii)
			_ = g.plain(*ascii)
			samples = append(samples, float64(time.Since(now).Nanoseconds())/1e6)
		}
		result = map[string]any{"load_ms": loadMs, "frames": *frames, "total_ms": float64(time.Since(start).Nanoseconds()) / 1e6, "frame_ms": samples, "point_count": len(t.Points)}
	default:
		return fmt.Errorf("unknown mode %q", mode)
	}
	if e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
