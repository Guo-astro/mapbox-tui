# Shell on Earth

A personal proof of concept for viewing GPX files in a terminal. Built with Go,
Bubble Tea, Bubbles and Lip Gloss. The executable is `shoe`.

- Pan, zoom and inspect tracks, routes and waypoints.
- Show an elevation profile and replay recorded positions at adjustable speeds.
- Use a synthetic offline basemap or Mapbox Streets vector tiles.
- Get structured results through the CLI or a local MCP server.

The bundled GPX fixtures are synthetic. The viewer currently needs a GPX file;
it doesn't provide directions, traffic data or live GPS navigation.

## Quick start

Requires Go **1.27.1** and Make. Run from the repository root:

```sh
make build
./bin/shoe open fixtures/example.gpx --demo
```

Run `./bin/shoe` without a file to open the GPX picker. Use `q` to quit.
If you manage Go with mise, prefix Make commands with `mise exec go@1.27.1 --`.

Braille rendering works best in a Unicode terminal. Use `--ascii` for a text
fallback. The minimum terminal size is 30 columns by 12 rows.

## Mapbox basemap

Supply your own Mapbox public access token through `MAPBOX_ACCESS_TOKEN`, or copy
the placeholder in `mapbox.example.json` into an ignored `mapbox.local.json`
and replace it locally. The environment variable takes precedence; an explicitly
empty value disables the local token. `.env` files aren't loaded automatically.

```sh
./bin/shoe open fixtures/example.gpx --provider mapbox
```

Without an explicit provider, the CLI/viewer uses Mapbox when a token is available
and the synthetic demo otherwise. `--demo` always forces synthetic geometry;
`--provider none` displays only the imported geometry.

Mapbox requests use your account and may incur charges. Provider errors leave the
GPX inspectable. See [provider notes](docs/PROVIDERS.md) for caching and attribution
requirements, including the unresolved graphical-logo requirement.

Credentials, personal GPX/FIT/TCX files, binaries and generated reports are ignored
by Git. Only the four explicitly allowed synthetic GPX fixtures are tracked.

## Controls

| Action | Keyboard | Mouse |
|---|---|---|
| Open another GPX | Ctrl-O or `o` | |
| Command palette / help | Ctrl-P / `?` | |
| Pan | Arrows or `h j k l` | Left-drag |
| Zoom | `+` / `-` | Wheel, anchored to pointer |
| Fit geometry | `f` / `r` | |
| Select recorded points | `[` / `]` | Click near a track |
| Seek along the track | PgUp/PgDn, Home/End | |
| Inspect selected position | Enter | |
| Play / pause | Space | |
| Change playback speed | `p` picker, `s` / `S` cycle | |
| Center and follow | `c` | |
| Show / hide elevation | `e` | |
| Focus route details | Tab | Wheel over sidebar |
| Rotate / north-up | `,` / `.` / `n` | Right-drag |
| Toggle ASCII / camera animation | `a` / `m` | |
| Provider credits | `v` | |
| Close dialog / quit | Esc / `q` or Ctrl-C | |

`S` and `E` mark part endpoints, `W` marks waypoints, and `@` marks the selected
or interpolated position. Pan and zoom keep playback running; selecting a point
pauses it. Manual pan disables follow mode, and `c` restores it.

Playback speeds are 1x, 5x, 20x, 60x and 120x. Parts with complete, increasing
timestamps use recorded timing. Other parts use a labelled **6 m/s estimate**.
Disconnected parts jump without inventing a connecting path. Interpolated values
are prefixed with `~`; missing elevations and timestamps remain unavailable.

## CLI

The GPX path comes before options:

```sh
./bin/shoe summary fixtures/example.gpx
./bin/shoe inspect fixtures/example.gpx --point-index 3
./bin/shoe inspect fixtures/example.gpx --distance-m 1000
./bin/shoe render fixtures/example.gpx --demo --width 120 --height 40
./bin/shoe render fixtures/example.gpx --demo --json
```

`summary` and `inspect` return JSON. `render` returns map text with attribution;
`render --json` returns plain map text and scene metadata, without ANSI sequences.
Indices are zero-based, distances are metres, and bounds are
`[west, south, east, north]`. Distance inspection returns the nearest recorded
point, not an interpolated measurement.

Camera options: `--center-lon` and `--center-lat` together, `--zoom` (0-20),
`--bearing` (-360 to 360), and `--cell-aspect` (1-3, default 2). Static map sizes
are 10-300 columns by 5-100 rows. Viewer options include `--ascii` and
`--no-animation`.

## MCP

```sh
./bin/shoe mcp
```

The local stdio server exposes four tools:

| Tool | Inputs | Result |
|---|---|---|
| `load_track` | Local `path` | Ephemeral track ID, bounds and summary |
| `summarize_track` | `track_id` | Geometry and measurement summary |
| `inspect_track` | `track_id` and either `point_index` or `distance_m` | Recorded point |
| `render_map` | `track_id` and viewport/provider options | Plain map text and scene metadata |

Use [mcp.example.json](mcp.example.json) with the absolute path to your binary.
Full schemas are in [shared/tools.json](shared/tools.json). Track IDs and tile
caches last only for that server process. MCP runs independently of the viewer.
Its render default is `demo`; an agent must request `provider: "mapbox"` explicitly
to use your account. For a client launched outside the repository, supply the
token through its environment.

## Development and checks

Python 3 is required for the development and validation scripts:

```sh
make dev
make dev ARGS='fixtures/example.gpx --demo --zoom 13'
make test
make fixtures
make bench
```

The development watcher rebuilds and restarts the viewer when Go/build inputs
change. It requires a Unix-like interactive terminal; view state resets on restart.

`make test` builds, runs Go tests and vet, then checks CLI/MCP consistency without
Mapbox credentials. Benchmarks and captures are generated locally under ignored
`reports/`. See [validation notes](docs/VALIDATION.md) for interaction and live
provider checks.

Dependencies are pinned in `go.mod` with checksums in `go.sum`. Setup follows the
dependency-minimisation and deterministic-install principles in this
[supply-chain reference](https://github.com/lirantal/npm-security-best-practices).

## Limitations

- Antimeridian-crossing geometry and bounds wider than 180 degrees are rejected.
- Display uses Mercator projection; distances use spherical Haversine calculations.
- Map labels are ASCII-only. FIT files and GPX activity extensions aren't parsed.
- Basemap detail is capped at tile zoom 14, even when the viewer zooms further.
- Distance includes each imported track and route, excluding gaps between parts.
- Terminal integration checks cover headless Linux/tmux; other environments aren't verified.
