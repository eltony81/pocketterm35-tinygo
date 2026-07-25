# PocketTerm35 - High-Performance TinyGo Keyboard Firmware & System Setup

Custom TinyGo (Go) firmware, system configurations, and hardware notes for the **Waveshare PocketTerm35** (Raspberry Pi 5 + 3.5" DPI LCD + RP2040 Pico Keyboard).

---

## 🚀 Features & Improvements

* **TinyGo (Go) Microcontroller Firmware**:
  * Ultra-low latency key matrix scanning using Go **goroutines** and **channels**.
  * **34 KB native ARM Cortex-M0+ binary** (replaces 1.2 MB CircuitPython interpreter).
  * Instant typing response & zero keypress delay.
  * Full matrix mapping (7 rows x 10 cols) including `Fn` keys, CapsLock LED (`GP22`), and status LED (`GP21`).
  * Automatic software-triggered bootloader flashing via `picotool`.

* **Display & Wayland Optimization (RPi 5 / Labwc)**:
  * DRM device binding forced to `/dev/dri/card1` (`WLR_DRM_DEVICES=/dev/dri/card1:/dev/dri/card0`) to prevent Labwc `NOOP-1` dummy screen bugs.
  * 5-minute DPMS screen standby (`swayidle`) targeting `HDMI-A-1` display output.

* **USB Power Management**:
  * USB autosuspend disabled via udev (`99-usb-no-autosuspend.rules`) for Pico HID controller (`1209:0001`).

---

## 📂 Repository Structure

```text
.
├── POCKETTERM35_HARDWARE_NOTES.md   # Complete hardware, DRM, and Wayland reference guide
├── waveshare_original_code.py       # Backup of factory Waveshare CircuitPython code
└── pico-keyboard-go/
    ├── main.go                      # TinyGo firmware source code for RP2040 Pico
    ├── go.mod                       # Go module configuration
    └── pico_keyboard_firmware.uf2  # Compiled native binary (34 KB)
```

---

## 🛠️ Building & Flashing Firmware

### Prerequisites
* Go 1.23+
* TinyGo 0.35+
* `picotool`

### Build Binary
```bash
cd pico-keyboard-go
tinygo build -target=pico -o pico_keyboard_firmware.uf2 main.go
```

### Flash to Pico (Software Reboot)
```bash
python3 -c "import serial, time; ser = serial.Serial('/dev/ttyACM0', 115200); ser.write(b'\x03\x03\r\nimport microcontroller\r\nmicrocontroller.on_next_reset(microcontroller.RunMode.BOOTLOADER)\r\nmicrocontroller.reset()\r\n')"
sudo picotool load pico-keyboard-go/pico_keyboard_firmware.uf2 -x
```

---

## 📄 License
MIT License
