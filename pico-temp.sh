#!/usr/bin/env bash
# Real-time Pico & RPi 5 Temperature Monitor

get_pico_temp() {
    python3 /home/tony/tools/get_pico_temp.py
}

get_rpi_temp() {
    vcgencmd measure_temp 2>/dev/null | cut -d= -f2 || echo "N/A"
}

# GUI Mode ONLY if explicitly requested via --gui
if [ "$1" = "--gui" ]; then
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

# Terminal Mode (Shell Console / TUI)
clear 2>/dev/null || true
echo "=============================================="
echo "    🌡️ PocketTerm35 Real-Time Temperature"
echo "=============================================="
echo " Press Ctrl+C to exit."
echo ""

while true; do
    P_TEMP=$(get_pico_temp)
    R_TEMP=$(get_rpi_temp)
    printf "\r[ %s ]  🟢 RP2040 Pico: %-10s | 🔴 RPi 5 CPU: %-10s" "$(date +%H:%M:%S)" "$P_TEMP" "$R_TEMP"
    sleep 2
done
