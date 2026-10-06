# Map providers

The viewer supports three providers:

- `demo`: synthetic geometry around the GPX bounds, labelled as demo data.
- `mapbox`: Mapbox Streets v8 vector tiles, using your own access token.
- `none`: imported GPX geometry only.

## Mapbox requests

Tiles come from the [Vector Tiles API](https://docs.mapbox.com/api/maps/vector-tiles/),
using the `mapbox.mapbox-streets-v8` tileset. The renderer reads water, roads and
place labels. Tile zoom is capped at 14; visible coverage is limited to 64 tiles
by reducing tile zoom when necessary.

Requests use HTTPS. Tokens, request URLs and provider response bodies aren't
included in application diagnostics. Redirects are disabled, requests time out
after eight seconds, and tile bodies are limited to 8 MiB.

## Caching and failures

- Each process keeps up to 64 decoded tiles in an in-memory LRU cache.
- Response cache headers determine freshness, capped at 12 hours; `Age` is subtracted.
- `no-store` and `no-cache` responses aren't retained for reuse.
- There are no persistent tile files, offline-region downloads or bundled Mapbox tiles.
- Camera changes are debounced; one tile batch runs at a time with sequential requests.
- Failures pause requests for ten seconds; HTTP 429 pauses for at least 60 seconds
  and honours integer `Retry-After` values.
- A stationary view doesn't periodically refresh. Pan/zoom can trigger another batch.
- Imported geometry remains usable when the provider fails.

## Attribution

[Mapbox's attribution requirements](https://docs.mapbox.com/help/dive-deeper/attribution/)
include a graphical logo and text credits for maps using its data. The current
terminal renderer displays text-only `© Mapbox`, `© OpenStreetMap` and an Improve
this map entry. Press `v` for links; CLI/MCP render results include attribution.

The official graphical logo isn't implemented, and a terminal-specific exception
hasn't been verified. The text credits alone don't fulfil that published requirement.
