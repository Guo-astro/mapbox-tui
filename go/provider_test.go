package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
)

type fakeTransport func(*http.Request) (*http.Response, error)

func (f fakeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSyntheticVectorTileDecodeAndWarmCache(t *testing.T) {
	road := geojson.NewFeature(orb.LineString{{0, 0}, {4096, 4096}})
	road.Properties["class"] = "primary"
	place := geojson.NewFeature(orb.Point{2048, 2048})
	place.Properties["name_en"] = "Synthetic place"
	data, e := mvt.Marshal(mvt.Layers{&mvt.Layer{Name: "road", Extent: 4096, Version: 2, Features: []*geojson.Feature{road}}, &mvt.Layer{Name: "place_label", Extent: 4096, Version: 2, Features: []*geojson.Feature{place}}})
	if e != nil {
		t.Fatal(e)
	}
	fs, e := decodeTile(data, Tile{2, 1, 1})
	if e != nil {
		t.Fatal(e)
	}
	if len(fs) != 2 || fs[0].Layer != 3 || fs[0].Parts[0][1] != (XY{0.5, 0.5}) || fs[1].Label != "Synthetic place" {
		t.Fatal("MVT layer/coordinate/property decoding failed")
	}
	p := newProvider("mapbox")
	p.token = "synthetic-test-token"
	p.client.Transport = fakeTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=60"}}, Body: io.NopCloser(strings.NewReader(string(data)))}, nil
	})
	track, e := loadTrack("../fixtures/example.gpx")
	if e != nil {
		t.Fatal(e)
	}
	c := fit(track, 120, 40, 2)
	_, cold, status := p.scene(context.Background(), track, c)
	if status != "" || cold.Requests == 0 {
		t.Fatal("cold tile request failed")
	}
	_, warm, status := p.scene(context.Background(), track, c)
	if status != "" || warm.Requests != cold.Requests || warm.Hits == 0 {
		t.Fatal("warm scene refetched tiles")
	}
}
func TestRateLimitBackoffDoesNotLeakToken(t *testing.T) {
	p := newProvider("mapbox")
	p.token = "never-log-this-test-token"
	p.client.Transport = fakeTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": []string{"120"}}, Body: io.NopCloser(strings.NewReader("private provider diagnostic"))}, nil
	})
	track, e := loadTrack("../fixtures/example.gpx")
	if e != nil {
		t.Fatal(e)
	}
	c := fit(track, 120, 40, 2)
	_, s, status := p.scene(context.Background(), track, c)
	if s.Requests != 1 || s.Errors != 1 || strings.Contains(status, p.token) {
		t.Fatal("rate limit behavior incorrect")
	}
	_, s, _ = p.scene(context.Background(), track, c)
	if s.Requests != 1 {
		t.Fatal("backoff refetched tiles")
	}
}
