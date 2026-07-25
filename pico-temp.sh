#!/usr/bin/env bash
# Real-time Pico & RPi 5 Temperature Monitor

get_pico_temp() {
    python3 -c "
import serial, time
try:
    ser = serial.Serial(/dev/ttyACM0, 115200, timeout=0.4)
    ser.write(bTEMPrn)
    time.sleep(0.15)
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
        P_TEMP=$(get_pico_temp)
        R_TEMP=$(get_rpi_temp)
        zenity --info \
          --title="PocketTerm35 Temperatures" \
          --text="<b>🌡️ PocketTerm35 Temperatures</b>\n\n🟢 <b>Raspberry Pi Pico (RP2040):</b>  ${P_TEMP}\n🔴 <b>Raspberry Pi 5 CPU:</b>             ${R_TEMP}" \
          --no-wrap 2>/dev/null
        exit 0
    fi
fi

# Terminal Mode
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
