#!/usr/bin/env python3
import serial
import time

try:
    ser = serial.Serial()
    ser.port = "/dev/ttyACM0"
    ser.baudrate = 115200
    ser.timeout = 0.5
    ser.dtr = True
    ser.rts = True
    ser.open()

    ser.reset_input_buffer()
    ser.reset_output_buffer()
    ser.write(b"TEMP\r\n")
    ser.flush()

    start = time.time()
    temp_found = False
    while time.time() - start < 1.0:
        if ser.in_waiting > 0:
            line = ser.readline().decode("utf-8", errors="ignore").strip()
            if "PICO_TEMP:" in line:
                print(line.split("PICO_TEMP:")[1].strip())
                temp_found = True
                break
        time.sleep(0.05)

    if not temp_found:
        print("N/A")

    ser.close()
except Exception as e:
    print("N/A")
