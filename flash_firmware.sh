#!/bin/bash
set -e
echo "=== PocketTerm35 TinyGo Firmware Flasher ==="
PICO_IP="192.168.1.28"
SSH_USER="tony"
PASS="uofofet"

echo "[1/3] Compiling firmware in TinyGo..."
cd "$(dirname "$0")/pico-keyboard-go"
/usr/local/tinygo/bin/tinygo build -target=pico -o pico_keyboard_firmware.uf2 main.go
echo "Firmware built: $(ls -lh pico_keyboard_firmware.uf2)"

echo "[2/3] Uploading binary to PocketTerm35 ($PICO_IP)..."
sshpass -p "$PASS" scp -o StrictHostKeyChecking=no pico_keyboard_firmware.uf2 $SSH_USER@$PICO_IP:/home/tony/Projects/mysystem/pico-keyboard-go/

echo "[3/3] Rebooting Pico to BOOTLOADER & Flashing via picotool..."
sshpass -p "$PASS" ssh -o StrictHostKeyChecking=no $SSH_USER@$PICO_IP '
python3 -c "
import serial, time
try:
    ser = serial.Serial(\"/dev/ttyACM0\", 115200, timeout=1)
    ser.write(b\"\x03\x03\r\n\")
    time.sleep(0.2)
    ser.write(b\"import microcontroller\r\n\")
    time.sleep(0.1)
    ser.write(b\"microcontroller.on_next_reset(microcontroller.RunMode.BOOTLOADER)\r\n\")
    time.sleep(0.1)
    ser.write(b\"microcontroller.reset()\r\n\")
    time.sleep(0.2)
except Exception as e:
    pass
"
sleep 0.8
echo uofofet | sudo -S picotool load /home/tony/Projects/mysystem/pico-keyboard-go/pico_keyboard_firmware.uf2 -x
'
echo "=== Flash Completed Successfully! ==="
