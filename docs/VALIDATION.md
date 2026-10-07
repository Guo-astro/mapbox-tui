# Validation

## Automated checks

From the repository root:

```sh
make test
```

This builds `bin/shoe`, runs Go tests and `go vet`, then checks the CLI and local
stdio MCP server. It covers GPX topology and missing measurements, camera/picking,
playback/profile behaviour, tile decoding and cache policy, token-free provider
errors, tool schemas, and CLI/MCP result consistency. No Mapbox account is required.

Optional checks, after building:

```sh
make fixtures
make bench
python3 scripts/interactions.py
```

Interaction checks require tmux and `rsvg-convert` and run in headless Linux
pseudo-terminals. They exercise keyboard/mouse input, dialogs, playback, resizing
and terminal restoration. They don't verify physical mouse forwarding or other
operating systems. Benchmark and capture output is ignored under `reports/`.

With a locally configured token, these checks make billable Mapbox requests:

```sh
python3 scripts/mapbox_check.py
python3 scripts/interactions.py --mapbox
```

## Manual smoke check

1. Open `./bin/shoe open fixtures/example.gpx --demo` and check separate parts and waypoints.
2. Pan, zoom, select points and use `f` to fit the geometry.
3. Play with Space, change speeds, toggle elevation with `e`, and pause/seek.
4. Open help, the command palette and GPX picker; check Esc closes each dialog.
5. Resize the terminal and try ASCII mode with `a`.
6. Quit with `q` or Ctrl-C and check the terminal is restored.
