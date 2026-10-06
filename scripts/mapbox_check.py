from pathlib import Path
import json
import os
import time
from validate import MCP, ROOT, BINS


def main():
    folder = ROOT / "reports/generated"
    folder.mkdir(exist_ok=True, parents=True)
    results = []
    for binary in BINS:
        m = MCP(binary, env=os.environ.copy())
        try:
            m.send("initialize", {"protocolVersion": "2025-11-25", "clientInfo": {"name": "mapbox-check", "version": "1"}, "capabilities": {}})
            m.send("notifications/initialized", {}, notify=True)
            track = m.call("load_track", {"path": str(ROOT / "fixtures/example.gpx")})["result"]["structuredContent"]
            args = {"track_id": track["track_id"], "provider": "mapbox", "width": 120, "height": 40, "center_lon": -0.1065, "center_lat": 51.505, "zoom": 13, "point_index": 3}
            runs = []
            for label in ("cold", "warm"):
                start = time.perf_counter()
                result = m.call("render_map", args)["result"]["structuredContent"]
                elapsed = (time.perf_counter() - start) * 1000
                scene = result["scene"]
                runs.append({"phase": label, "elapsed_ms": elapsed, "tile_stats": scene["tile_stats"], "status": scene["status"]})
                (folder / f"{binary.name}-mapbox-{label}.txt").write_text(result["text"] + "\n" + scene["attribution"] + "\n" + "\n".join(scene["attribution_links"]) + "\n")
                assert scene["status"] == "", scene["status"]
                assert scene["tile_stats"]["Requests"] > 0
                assert scene["tile_stats"]["Errors"] == 0
            assert runs[1]["tile_stats"]["Requests"] == runs[0]["tile_stats"]["Requests"], "warm scene refetched tiles"
            assert runs[1]["tile_stats"]["Hits"] > 0
            record = {"binary": binary.name, "viewport": args, "runs": runs}
            results.append(record)
            print(f'{binary.name}: {runs[0]["tile_stats"]["Requests"]} cold requests, {runs[1]["tile_stats"]["Hits"]} warm hits, {runs[0]["elapsed_ms"]:.0f} ms cold / {runs[1]["elapsed_ms"]:.1f} ms warm; no errors')
        finally:
            m.close()
    (ROOT / "reports/mapbox.json").write_text(json.dumps(results, indent=2) + "\n")


if __name__ == "__main__":
    main()
