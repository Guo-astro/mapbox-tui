import subprocess
import sys
import termios
import time

before = termios.tcgetattr(0)
result = subprocess.run(sys.argv[1:])
after = termios.tcgetattr(0)
print(f"TERMINAL_RESTORED={before == after} EXIT={result.returncode}", flush=True)
time.sleep(15)
