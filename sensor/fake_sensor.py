import sys
import time

time.sleep(1)
while True:
    ts = int(time.time())
    sys.stdout.write(f"{ts},25.2296875,43.159619678029735,1009.1293692371094,70.40053920398444,0")
    sys.stdout.flush()
    time.sleep(1)
