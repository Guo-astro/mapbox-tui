from pathlib import Path
from datetime import datetime, timedelta, timezone
import math

ROOT = Path(__file__).resolve().parents[1]


def generate(count):
    path = ROOT / "fixtures" / f"dense-{count}.gpx"
    with path.open("w") as f:
        f.write('<gpx version="1.1"><trk><name>Dense riverside loop</name><trkseg>\n')
        for i in range(count):
            if i == count // 2:
                f.write('</trkseg><trkseg>\n')
            u = i / (count - 1) * math.tau
            lat = 51.505 + math.sin(u) * 0.006
            lon = -0.105 + math.cos(u) * 0.018
            ele = 15 + math.sin(u * 3) * 8
            stamp = (datetime(2026, 1, 1, tzinfo=timezone.utc) + timedelta(seconds=i)).isoformat().replace('+00:00', 'Z')
            f.write(f'<trkpt lat="{lat:.9f}" lon="{lon:.9f}"><ele>{ele:.3f}</ele><time>{stamp}</time></trkpt>\n')
        f.write('</trkseg></trk></gpx>\n')
    print(f"Generated {path.relative_to(ROOT)} ({count:,} points)")


if __name__ == "__main__":
    for n in (10_000, 100_000):
        generate(n)
