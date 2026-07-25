#!/usr/bin/env python3
import serial
import time

try:
    ser = serial.Serial("/dev/ttyACM0", 115200, timeout=0.3)
    ser.write(b"TEMP\r\n")
    time.sleep(0.15)
    resp = ser.read_all().decode("utf-8", errors="ignore")
    for line in resp.splitlines():
        if "PICO_TEMP:" in line:
            print(line.split("PICO_TEMP:")[1].strip())
            exit(0)
    print("N/A")
except Exception:
    print("N/A")
