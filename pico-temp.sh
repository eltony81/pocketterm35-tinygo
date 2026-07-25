#!/usr/bin/env bash
# Real-time Pico & RPi 5 Temperature Monitor

# Reset terminal attributes and colors so text is never black-on-black
stty sane 2>/dev/null || true
printf "\033[0m"

get_pico_temp() {
    python3 /home/tony/tools/get_pico_temp.py
}

get_rpi_temp() {
    vcgencmd measure_temp 2>/dev/null | cut -d= -f2 || echo "N/A"
}

P_TEMP=$(get_pico_temp)
R_TEMP=$(get_rpi_temp)

# 1. GUI Mode (Zenity Desktop Popup when double-clicking Desktop Icon)
if [ "${1:-}" = "--gui" ]; then
    if which zenity >/dev/null 2>&1; then
        zenity --info \
          --title="PocketTerm35 Temperatures" \
          --text="<b>🌡️ PocketTerm35 Temperatures</b>\n\n🟢 <b>Raspberry Pi Pico (RP2040):</b>  ${P_TEMP}\n🔴 <b>Raspberry Pi 5 CPU:</b>             ${R_TEMP}" \
          --no-wrap 2>/dev/null
        exit 0
    fi
fi

# 2. Continuous Loop Mode (--loop)
if [ "${1:-}" = "--loop" ]; then
    printf "\033[0m\033[2J\033[H"
    echo "=============================================="
    echo "    🌡️ PocketTerm35 Real-Time Temperature"
    echo "=============================================="
    echo " Press Ctrl+C to exit."
    echo ""

    while true; do
        P_TEMP=$(get_pico_temp)
        R_TEMP=$(get_rpi_temp)
        echo -e "[ $(date +%H:%M:%S) ] \033[1;32m🟢 RP2040 Pico: ${P_TEMP}\033[0m  |  \033[1;31m🔴 RPi 5 CPU: ${R_TEMP}\033[0m"
        sleep 2
    done
    exit 0
fi

# 3. Default Shell / menu.sh Mode (High-contrast bright ANSI text)
printf "\033[0m"
echo ""
echo -e "\033[1;36m==================================================\033[0m"
echo -e "\033[1;37m    🌡️ PocketTerm35 Temperature Summary\033[0m"
echo -e "\033[1;36m==================================================\033[0m"
echo ""
echo -e "  \033[1;32m🟢 Raspberry Pi Pico (RP2040):\033[0m  \033[1;37m${P_TEMP}\033[0m"
echo -e "  \033[1;31m🔴 Raspberry Pi 5 CPU:\033[0m             \033[1;37m${R_TEMP}\033[0m"
echo ""
echo -e "\033[1;36m==================================================\033[0m"
echo ""
