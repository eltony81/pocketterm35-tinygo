# PocketTerm35 Hardware & System Diagnostics Guide

## 1. Hardware Architecture Summary
* **Device**: Waveshare PocketTerm35 case + Raspberry Pi 5 (8GB RAM).
* **Display**: 3.5" DPI/HDMI LCD (640x480 @ 75Hz).
* **Touchscreen**: Goodix Capacitive TouchScreen (`1-005d Capacitive TouchScreen`, `/dev/input/event7`).
* **Keyboard & Controls**: RP2040 Raspberry Pi Pico running Adafruit CircuitPython 10.0.0 (`usb:1209:0001`).
* **Network / SSH**: IP `192.168.1.28`, Hostname `pocketterm35`, User `tony`, Password `uofofet`.

---

## 2. Display System (Labwc Wayland Compositor)
### Primary DRM Card Configuration
The Raspberry Pi 5 has two DRM devices:
* `/dev/dri/card0`: V3D 3D engine (no physical display outputs).
* `/dev/dri/card1`: HDMI / vc4 display controller (connected to `HDMI-A-1` 3.5" LCD).

**CRITICAL REQUIREMENT**:
To prevent Labwc from falling back to a virtual dummy screen (`NOOP-1`), the environment MUST explicitly force `/dev/dri/card1`:
```bash
WLR_DRM_DEVICES=/dev/dri/card1:/dev/dri/card0
```
This is configured in:
* `/etc/environment`
* `~/.config/labwc/environment`
* `/etc/xdg/labwc/environment`

### Screen Standby (Swayidle)
Standby (5 minutes / 300 seconds) is managed by `swayidle` targeting `HDMI-A-1`:
```bash
swayidle -w timeout 300 '[ "$(cat /sys/devices/virtual/tty/tty0/active)" = "tty8" ] && wlr-randr --output HDMI-A-1 --off' resume '[ "$(cat /sys/devices/virtual/tty/tty0/active)" = "tty8" ] && wlr-randr --output HDMI-A-1 --on' &
```
Configured in:
* `~/.config/labwc/autostart`
* `/etc/xdg/labwc-greeter/autostart`

---

## 3. Pico Keyboard Controller (CircuitPython `code.py`)
The physical keyboard and backlight/audio PWM controls are managed by an onboard RP2040 Pico running CircuitPython.

### Key Features & Physical Shortcuts in `/code.py`:
* **Matrix Scanning**: 7 rows x 10 columns matrix.
* **USB HID Devices**: Emulates standard Keyboard (`usb_hid.devices`) & ConsumerControl keys.
* **Fn Key Mappings**:
  * `Fn + -` / `Fn + +` / `Fn + Z` / `Fn + X`: Controls Screen Backlight Brightness (PWM on `GP20`).
  * `Fn + Vol Up` / `Fn + Vol Down` / `Fn + Mute`: Controls Audio PWM on `GP18`.
  * `Fn + Breathing Light`: Controls `GP21` LED effect.
  * `CapsLock`: Toggles `GP22` LED and emits `Keycode.CAPS_LOCK`.

### Serial Command Protocol (`/dev/ttyACM0`)
The updated `/code.py` listens for non-blocking text commands over serial without interrupting HID key scanning:
* `BRIGHTNESS:UP` or `BL_UP`: Increases brightness (same as `Fn + +`)
* `BRIGHTNESS:DOWN` or `BL_DOWN`: Decreases brightness (same as `Fn + -`)
* `POWERSAVE:ON` or `POWERSAVE`: Sets screen brightness to dim powersave mode (15%)
* `POWERSAVE:OFF` or `FULL`: Restores screen brightness to 100%
* `BRIGHTNESS:<0-100>`: Sets brightness directly to a percentage (e.g. `BRIGHTNESS:30`)

### Command-Line Utility (`pocketterm-brightness`)
A helper CLI command is installed at `/usr/local/bin/pocketterm-brightness`:
```bash
pocketterm-brightness powersave    # Dims screen to 15% (Powersave mode)
pocketterm-brightness full         # Restores screen to 100% brightness
pocketterm-brightness up           # Increases brightness
pocketterm-brightness down         # Decreases brightness
pocketterm-brightness 40           # Sets brightness to 40%
```

### USB Autosuspend Prevention
To prevent Linux from putting the Pico USB HID controller to sleep after 2 seconds, USB autosuspend is disabled via udev (`/etc/udev/rules.d/99-usb-no-autosuspend.rules`):
```udev
ACTION=="add", SUBSYSTEM=="usb", ATTR{idVendor}=="1209", ATTR{idProduct}=="0001", ATTR{power/control}="on", ATTR{power/autosuspend_delay_ms}="-1"
```

---

## 4. LightDM Autologin & User Session
LightDM is configured for automatic login of user `tony` into `rpd-labwc` via `/etc/lightdm/lightdm.conf.d/00-autologin.conf`:
```ini
[Seat:*]
autologin-user=tony
autologin-user-timeout=0
autologin-session=rpd-labwc
```

To toggle CapsLock state via software if capital letters appear:
```bash
WAYLAND_DISPLAY=wayland-0 XDG_RUNTIME_DIR=/run/user/1001 DISPLAY=:0 xdotool key Caps_Lock
```
