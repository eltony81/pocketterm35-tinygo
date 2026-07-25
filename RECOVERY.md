# PocketTerm35 RP2040 Firmware Architecture & Recovery Manual

This repository contains the Go source code and build system for replacing the default CircuitPython RP2040 keyboard controller on the Waveshare PocketTerm35 with an optimized **TinyGo** firmware.

---

## 1. System Overview

The **Waveshare PocketTerm35** uses a dual-chip architecture:
1. **Raspberry Pi 5 / 4B**: Main SBC running Linux OS (Raspberry Pi OS, TTY shell `sh`/`bash`, X11/Wayland).
2. **Raspberry Pi Pico (RP2040)**: Microcontroller managing the 70-key physical matrix, LCD backlight PWM, audio volume PWM, CapsLock LED, Mute output, and Breathing LED.

Communication between the RP2040 Pico and the Raspberry Pi 5 takes place over an **internal USB HID Keyboard bus**.

---

## 2. Reverse Engineering of Waveshare Python Firmware (`waveshare_original_code.py`)

### 2.1 Hardware Connections & Peripheral Map

| Function | RP2040 Pin | Hardware Mode | Default / Rest Value |
| :--- | :--- | :--- | :--- |
| **Matrix Rows (7)** | `GP0` - `GP6` | Digital Input (`Pull.DOWN`) | Active HIGH |
| **Matrix Columns (10)** | `GP7` - `GP16` | Digital Output | `False` (0V) at rest, `True` (3.3V) on scan |
| **CapsLock LED** | `GP22` | Digital Output | `False` (Off) |
| **Audio Mute** | `GP19` | Digital Output | `False` (Unmuted) |
| **Status / Screen LED** | `GP21` | Digital Output | `False` (Off) |
| **LCD Backlight PWM** | `GP20` | Hardware PWM (5000Hz) | `duty_cycle = 5000` (~7.6%) |
| **Audio Volume PWM** | `GP18` | Hardware PWM (5000Hz) | `duty_cycle = 32700` (~50%) |

---

### 2.2 Key Dispatching Architecture (Set Difference)

CircuitPython evaluates active keypresses using a **Set-Difference Dispatcher**:
- `current_keys = set(scan_keyboard())`
- `released_keys = previous_keys - current_keys` $\rightarrow$ calls `kbd.release(key)`
- `pressed_keys = current_keys - previous_keys` $\rightarrow$ calls `kbd.press(key)` or executes special functions
- Loop frequency: 100Hz (`time.sleep(0.01)`)

---

## 3. TinyGo Conversion Specifications

To achieve 100% feature parity and native Linux TTY `sh` shell compatibility:

1. **Standard 8-Byte USB HID Boot Keyboard Protocol**:
   - `bInterfaceClass`: `0x03` (HID)
   - `bInterfaceSubClass`: `0x01` (Boot Interface Subclass)
   - `bInterfaceProtocol`: `0x01` (Keyboard)
   - Packet format: `[Modifiers, Reserved(0x00), Key1, Key2, Key3, Key4, Key5, Key6]` without composite Report IDs.

2. **Zero-Allocation Set-Difference Scanner**:
   - Replicates CircuitPython's set-difference algorithm in Go using stack-allocated structs `KeySet` (`[16]keyboard.Keycode`).

3. **PWM Frequency & Duty Cycle Matching**:
   - GP20 (Backlight): 5kHz PWM, 5000 initial duty cycle.
   - GP18 (Audio Volume): 5kHz PWM, 32700 initial duty cycle.

---

## 4. Recovery Procedure (Rollback to Original Factory Firmware)

If you ever need to restore the original Waveshare CircuitPython firmware:

### Step 1: Boot into RP2040 Bootloader
Reset the USB bus or trigger 1200 baud open on `/dev/ttyACM0`:
```bash
python3 -c "import serial; s=serial.Serial('/dev/ttyACM0', 1200); s.close()"
```

### Step 2: Flash CircuitPython UF2
```bash
sudo picotool load adafruit-circuitpython-pico-9.2.4.uf2 -x
```

### Step 3: Copy `code.py` to `CIRCUITPY` Drive
```bash
sudo mkdir -p /mnt/circuitpy
sudo mount /dev/sda1 /mnt/circuitpy
sudo cp waveshare_original_code.py /mnt/circuitpy/code.py
sync
```
