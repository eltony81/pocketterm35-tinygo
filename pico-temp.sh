#!/usr/bin/env bash
# Real-time Pico & RPi 5 Temperature Monitor

get_pico_temp() {
    python3 -c "
import serial, time
try:
    ser = serial.Serial(/dev/ttyACM0, 115200, timeout=0.5)
    ser.write(bTEMPrn)
    time.sleep(0.2)
    resp = ser.read_all().decode(utf-8, errors=ignore)
    for line in resp.splitlines():
        if PICO_TEMP: in line:
            print(line.split(PICO_TEMP:)[1].strip())
            exit(0)
    print(N/A)
except Exception:
    print(N/A)
"
}

get_rpi_temp() {
    vcgencmd measure_temp 2>/dev/null | cut -d= -f2 || echo "N/A"
}

if [ "$1" = "--gui" ] || [ -n "$DISPLAY" ] || [ -n "$WAYLAND_DISPLAY" ]; then
    if which zenity >/dev/null 2>&1; then
        (
            while true; do
                P_TEMP=$(get_pico_temp)
                R_TEMP=$(get_rpi_temp)
                echo "100"
                echo "# 🌡️ Real-Time Temperature Monitor\n\n📌 Raspberry Pi Pico (RP2040):  $P_TEMP\n📌 Raspberry Pi 5 CPU:              $R_TEMP"
                sleep 2
            done
        ) | zenity --progress --title="PocketTerm35 Temperatures" --text="Reading temperatures..." --percentage=100 --no-cancel --width=380 --height=160 2>/dev/null
        exit 0
    fi
fi

# Terminal Mode (Live updates every 2 seconds)
clear
echo "=============================================="
echo "    🌡️ PocketTerm35 Real-Time Temperature"
echo "=============================================="
echo " Press Ctrl+C to exit."
echo ""

while true; do
    P_TEMP=$(get_pico_temp)
    R_TEMP=$(get_rpi_temp)
    printf "\r[ %s ] 🟢 RP2040 Pico Temp: %-10s | 🔴 RPi 5 CPU Temp: %-10s" "$(date +%H:%M:%S)" "$P_TEMP" "$R_TEMP"
    sleep 2
done
