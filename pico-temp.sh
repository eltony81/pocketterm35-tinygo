#!/bin/sh
# PocketTerm35 Temperature Monitor
# Da copiare in ~/tools/ sul PocketTerm35 (Raspberry Pi)

TOOLS_DIR="$(cd "$(dirname "$0")" && pwd)"

# Temperatura RPi5 — cattura prima in variabile, poi processa
# Il pipe diretto "vcgencmd | cut" causa il blocco se vcgencmd è lento
get_rpi_temp() {
    RESULT="$(timeout 3 vcgencmd measure_temp 2>/dev/null)"
    if [ -n "$RESULT" ]; then
        echo "$RESULT" | cut -d= -f2
    else
        echo "N/A"
    fi
}

# Temperatura Pico RP2040 via seriale
get_pico_temp() {
    PY="${TOOLS_DIR}/get_pico_temp.py"
    if [ -f "$PY" ]; then
        timeout 3 python3 "$PY" 2>/dev/null || echo "N/A"
    else
        echo "N/A"
    fi
}

P_TEMP="$(get_pico_temp)"
R_TEMP="$(get_rpi_temp)"

# 1. GUI Mode
if [ "${1:-}" = "--gui" ]; then
    if command -v zenity >/dev/null 2>&1; then
        zenity --info \
          --title="PocketTerm35 Temperatures" \
          --text="<b>🌡️ PocketTerm35 Temperatures</b>\n\n🟢 <b>Raspberry Pi Pico (RP2040):</b>  ${P_TEMP}\n🔴 <b>Raspberry Pi 5 CPU:</b>  ${R_TEMP}" \
          --no-wrap 2>/dev/null
    fi
    exit 0
fi

# 2. Loop Mode
if [ "${1:-}" = "--loop" ]; then
    printf "\033[2J\033[H"
    echo "=============================================="
    echo "    🌡️ PocketTerm35 Real-Time Temperature"
    echo "=============================================="
    echo " Premi Ctrl+C per uscire."
    echo ""
    while true; do
        P_TEMP="$(get_pico_temp)"
        R_TEMP="$(get_rpi_temp)"
        printf "[ %s ]  🟢 Pico: %s  |  🔴 RPi5: %s\n" \
            "$(date +%H:%M:%S)" "${P_TEMP}" "${R_TEMP}"
        sleep 2
    done
    exit 0
fi

# 3. Output normale
printf "\033[0m\n"
echo "=================================================="
echo "    🌡️  PocketTerm35 Temperature Summary"
echo "=================================================="
echo ""
printf "  🟢 Raspberry Pi Pico (RP2040):  %s\n" "${P_TEMP}"
printf "  🔴 Raspberry Pi 5 CPU:           %s\n" "${R_TEMP}"
echo ""
echo "=================================================="
echo ""
