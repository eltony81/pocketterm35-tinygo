# PocketTerm35 Firmware & Recovery Guide

## 1. TinyGo Custom Firmware Overview

The custom firmware in `pico-keyboard-go/main.go` is a high-performance HID Keyboard and hardware controller compiled with **TinyGo (v0.35.0)** for the Raspberry Pi Pico (RP2040).

### Key Features & Optimizations
- **Single-Pass Zero-Allocation Matrix Scanner**: Scans the 7x10 key matrix directly into fixed stack memory (`[7][10]bool`), eliminating dynamic heap allocations (`make(map...)`) in the 500Hz loop. Reduces latency to <1ms and prevents TinyGo Garbage Collector pauses.
- **Hardware PWM LCD Backlight & Audio Volume Control**:
  - `GP20`: LCD backlight PWM (5kHz). Controllable via `Fn + Z` / `Fn + X` or serial `BRIGHTNESS:UP/DOWN`.
  - `GP18`: Audio volume PWM (5kHz). Controllable via `Fn + Minus` / `Fn + Equals` or serial `VOL:UP/DOWN`.
- **Shift Key Macro Shortcuts**:
  - `ShiftGrave` (`~`): `Fn + H`
  - `ShiftBackslash` (`|`): `Fn + J`
  - `ShiftLeftBracket` (`{`): `Fn + K`
  - `ShiftRightBracket` (`}`): `Fn + L`
- **Breathing LED Animation**: Non-blocking `breathingController` goroutine on CapsLock LED (`GP22`), triggered by `FnBLControlScreen` (`Fn + Q`) and automatically cancelled on keypress.
- **Serial USB CDC Interface & ADC Temperature Reading**:
  - Non-blocking `serialListener` goroutine on USB CDC.
  - Reads Pico RP2040 internal temperature sensor via ADC (`machine.ReadTemperature()`).
  - Serial Commands: `TEMP`, `STATUS`, `BRIGHTNESS:UP`, `BRIGHTNESS:DOWN`, `VOL:UP`, `VOL:DOWN`, `HELP`.

---

## 2. Flashing the TinyGo Firmware

To compile and flash the TinyGo firmware onto the Raspberry Pi Pico:

```bash
# On PocketTerm35 (Raspberry Pi):
./flash_firmware.sh
```

### Manual Flash Procedure via picotool:
```bash
# 1. Reset Pico into BOOTLOADER mode (if running CircuitPython):
python3 -c "
import serial, time
try:
    ser = serial.Serial('/dev/ttyACM0', 115200, timeout=1)
    ser.write(b'\x03\x03\r\n')
    time.sleep(0.2)
    ser.write(b'import microcontroller\r\n')
    time.sleep(0.1)
    ser.write(b'microcontroller.on_next_reset(microcontroller.RunMode.BOOTLOADER)\r\n')
    time.sleep(0.1)
    ser.write(b'microcontroller.reset()\r\n')
except Exception:
    pass
"

# 2. Flash UF2 binary:
echo uofofet | sudo -S picotool load ~/Projects/mysystem/pico-keyboard-go/pico_keyboard_firmware.uf2 -x
```

---

## 3. Rollback & Recovery to Original Waveshare CircuitPython Code

If you ever need to restore the factory Waveshare CircuitPython firmware and original `code.py`:

### Recovery Assets Saved on PocketTerm35:
- **CircuitPython UF2 Firmware**: `~/Projects/mysystem/recovery/adafruit-circuitpython-pico-9.2.4.uf2`
- **Original Waveshare Python Code**: `waveshare_original_code.py` in this repository.

### Step-by-Step Rollback Procedure:

1. **Reset Pico into BOOTLOADER mode**:
   - Hold the `BOOTSEL` button on the Pico while plugging in USB, OR reboot via software if accessible.

2. **Flash CircuitPython UF2**:
   ```bash
   echo uofofet | sudo -S picotool load ~/Projects/mysystem/recovery/adafruit-circuitpython-pico-9.2.4.uf2 -x
   ```

3. **Restore `code.py`**:
   Once CircuitPython reboots, the Pico exposes a USB drive named `CIRCUITPY` (mounted at `/media/tony/CIRCUITPY` or `/mnt/CIRCUITPY`):
   ```bash
   cp /home/tony/Projects/pocketterm35-tinygo/waveshare_original_code.py /media/tony/CIRCUITPY/code.py
   ```
