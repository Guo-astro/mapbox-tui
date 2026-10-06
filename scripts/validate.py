from pathlib import Path
import json
import math
import os
import selectors
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
ENV = {**os.environ, "MAPBOX_ACCESS_TOKEN": ""}
BINS = [ROOT / "bin/wtsh"]


def cli(binary, mode, path, *options, success=True):
    p = subprocess.run([str(binary), mode, str(path), *options], cwd=ROOT, env=ENV, capture_output=True, text=True, timeout=20)
    assert (p.returncode == 0) == success, (binary.name, p.stderr)
    return json.loads(p.stdout) if success else None


def equivalent(a, b):
    if isinstance(a, float) or isinstance(b, float):
        assert math.isclose(a, b, abs_tol=1e-6, rel_tol=1e-10), (a, b)
    elif isinstance(a, dict):
        assert a.keys() == b.keys(), (a.keys(), b.keys())
        for key in a:
            equivalent(a[key], b[key])
    elif isinstance(a, list):
        assert len(a) == len(b)
        for x, y in zip(a, b):
            equivalent(x, y)
    else:
        assert a == b, (a, b)


class MCP:
    def __init__(self, binary, env=None):
        self.p = subprocess.Popen([str(binary), "mcp"], cwd=ROOT, env=ENV if env is None else env, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, bufsize=1)
        self.id = 0
        self.selector = selectors.DefaultSelector()
        self.selector.register(self.p.stdout, selectors.EVENT_READ)

    def send(self, method, params, notify=False):
        message = {"jsonrpc": "2.0", "method": method, "params": params}
        if not notify:
            self.id += 1
            message["id"] = self.id
        self.p.stdin.write(json.dumps(message) + "\n")
        self.p.stdin.flush()
        if notify:
            return
        assert self.selector.select(15), f"MCP timeout for {method}"
        line = self.p.stdout.readline()
        assert line, f"MCP exited during {method}: {self.p.poll()}"
        result = json.loads(line)
        assert result["id"] == self.id, result
        return result

    def call(self, name, args):
        return self.send("tools/call", {"name": name, "arguments": args})

    def close(self):
        self.p.stdin.close()
        self.p.wait(timeout=10)
        diagnostics = self.p.stderr.read()
        assert self.p.returncode == 0, diagnostics
        self.selector.close()


def validate_mcp(binary):
    m = MCP(binary)
    try:
        init = m.send("initialize", {"protocolVersion": "2025-11-25", "clientInfo": {"name": "prototype-validation", "version": "1"}, "capabilities": {}})
        assert "tools" in init["result"]["capabilities"]
        m.send("notifications/initialized", {}, notify=True)
        tools = m.send("tools/list", {})["result"]["tools"]
        expected = json.loads((ROOT / "shared/tools.json").read_text())
        assert {t["name"]: t["inputSchema"] for t in tools} == {t["name"]: t["inputSchema"] for t in expected}
        assert not m.send("ping", {}).get("error")
        loaded = m.call("load_track", {"path": str(ROOT / "fixtures/example.gpx")})["result"]["structuredContent"]
        track_id = loaded["track_id"]
        summary = m.call("summarize_track", {"track_id": track_id})["result"]["structuredContent"]
        equivalent(summary, loaded["summary"])
        equivalent(summary, cli(binary, "summary", ROOT / "fixtures/example.gpx"))
        for args in ({"point_index": 6}, {"distance_m": summary["distance_m"] / 2}):
            point = m.call("inspect_track", {"track_id": track_id, **args})["result"]["structuredContent"]
            assert point["point"]["segment_index"] >= 0
            options = [item for key, value in args.items() for item in ("--" + key.replace("_", "-"), str(value))]
            equivalent(point, cli(binary, "inspect", ROOT / "fixtures/example.gpx", *options))
        for args in ({"width": 120, "height": 40}, {"width": 160, "height": 50, "bearing": 40, "ascii": True}, {"provider": "mapbox"}):
            result = m.call("render_map", {"track_id": track_id, **args})["result"]["structuredContent"]
            assert "\x1b" not in result["text"]
            assert all(ord(c) >= 32 or c == "\n" for c in result["text"])
            assert len(result["text"].splitlines()) == result["scene"]["height"]
            options = []
            for key, value in args.items():
                flag = "--" + key.replace("_", "-")
                options.extend([flag] if value is True else [flag, str(value)])
            equivalent(result, cli(binary, "render", ROOT / "fixtures/example.gpx", "--json", *options))
            if args.get("provider") == "mapbox":
                assert result["scene"]["tile_stats"]["Requests"] == 0
                assert "missing" in result["scene"]["status"]
        for name, args in [
            ("inspect_track", {"track_id": track_id, "point_index": 999}),
            ("inspect_track", {"track_id": track_id, "point_index": 0, "distance_m": 0}),
            ("summarize_track", {"track_id": "unknown"}),
            ("render_map", {"track_id": track_id, "width": 999999}),
            ("render_map", {"track_id": track_id, "center_lon": 0}),
            ("load_track", {"path": str(ROOT / "fixtures/invalid.gpx")}),
        ]:
            result = m.call(name, args)
            assert result.get("error") or result.get("result", {}).get("isError"), (name, result)
        return summary
    finally:
        m.close()


def main():
    for binary in BINS:
        validate_mcp(binary)
        result=cli(binary,"render",ROOT / "fixtures/example.gpx","--demo","--provider","mapbox","--json")
        assert result["scene"]["provider"]=="demo"
        for flag in ("--zoom","--bearing","--cell-aspect"):
            cli(binary,"render",ROOT / "fixtures/example.gpx",flag,"NaN",success=False)
    for name in ("example.gpx", "waypoints.gpx", "dense-10000.gpx", "dense-100000.gpx"):
        path = ROOT / "fixtures" / name
        if not path.exists():
            continue
        for binary in BINS:
            summary = cli(binary, "summary", path)
            assert summary["bounds"] is not None
            for width, height in ((120, 40), (160, 50)):
                result = cli(binary, "render", path, "--demo", "--json", "--width", str(width), "--height", str(height))
                assert result["scene"]["width"] == width
                assert result["scene"]["height"] == height
                lines = result["text"].splitlines()
                assert len(lines) == height
                assert all(len(line) == width for line in lines)
                assert "\x1b" not in result["text"]
    invalid = ["<gpx><trk>", '<gpx><wpt lat="NaN" lon="0"/></gpx>', '<gpx><wpt lat="51oops" lon="0"/></gpx>', '<gpx><rte><rtept lat="0" lon="179"/><rtept lat="0" lon="-179"/></rte></gpx>']
    with tempfile.TemporaryDirectory() as folder:
        path = Path(folder) / "bad.gpx"
        for body in invalid:
            path.write_text(body)
            for binary in BINS:
                cli(binary, "summary", path, success=False)
    print("Validated Go CLI and MCP consistency, tool schemas and all four tools, GPX errors, summaries and map scenes.")


if __name__ == "__main__":
    main()
