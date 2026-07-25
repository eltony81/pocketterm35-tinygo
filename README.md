# PocketTerm35 - High-Performance TinyGo Keyboard Firmware & System Setup

Custom TinyGo (Go) keyboard firmware, real-time temperature monitoring, system configurations, and hardware notes for the **Waveshare PocketTerm35** (Raspberry Pi 5 + 3.5" DPI LCD + RP2040 Pico Keyboard).

---

## 🚀 Key Features

* **TinyGo (Go) Keyboard Firmware**:
  * Ultra-low latency key matrix scanning using Go **goroutines** and **channels**.
  * **34 KB native ARM Cortex-M0+ binary** (replaces 1.2 MB CircuitPython interpreter).
  * Instant typing response & zero keypress delay.
  * Full matrix mapping (7 rows x 10 cols) including `Fn` key combinations, CapsLock LED (`GP22`), and status LED (`GP21`).
  * **Internal RP2040 Temperature Sensor Reader**: Reads internal die temperature via ADC and reports real-time data over serial (`TEMP` / `PICO_TEMP`).

* **Real-Time Temperature Monitoring**:
  * Desktop GUI shortcut (`Pico Temperature.desktop`) displaying live RP2040 Pico & RPi 5 CPU temperatures side-by-side.
  * Interactive Tools Menu entry (`~/tools/pico-temp.sh` & `~/tools/tools_menu.sh`).

* **Display & Wayland Optimization (RPi 5 / Labwc)**:
  * DRM device binding forced to `/dev/dri/card1` (`WLR_DRM_DEVICES=/dev/dri/card1:/dev/dri/card0`) in `/usr/bin/labwc-pi` to prevent Labwc `NOOP-1` dummy screen bugs.
  * 5-minute DPMS screen standby (`swayidle`) targeting `HDMI-A-1` display output with instant touch/key wake-up.

* **USB Power Management**:
  * USB autosuspend disabled via udev (`99-usb-no-autosuspend.rules`) for Pico HID controller (`1209:0001`).

---

## 📂 Repository Structure

```text
.
├── README.md                      # Complete project documentation
├── POCKETTERM35_HARDWARE_NOTES.md # Detailed hardware notes, DRM setup & Wayland guide
├── pico-temp.sh                  # Real-time temperature monitor script (GUI & CLI)
├── flash_firmware.sh             # Automated script to compile & flash TinyGo firmware
├── waveshare_original_code.py    # Backup of original factory CircuitPython code
└── pico-keyboard-go/
    ├── main.go                   # TinyGo keyboard firmware source code with temp reader
    ├── go.mod                    # Go module configuration
    └── pico_keyboard_firmware.uf2 # Pre-compiled native firmware binary (34 KB)
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
./flash_firmware.sh
```

---

## 📄 License
MIT License
