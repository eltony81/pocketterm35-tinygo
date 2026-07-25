package main

import (
	"machine"
	"machine/usb/hid/keyboard"
	"time"
)

// Custom keys for special functions
const (
	FnKey = 0x1000 + iota
	FnMute
	FnVolDown
	FnVolUp
	FnLockScreen
	FnBLControlScreen
	FnBLPWMDown
	FnBLPWMUp
	ShiftGrave
	ShiftBackslash
	ShiftLeftBracket
	ShiftRightBracket
)

// Hardware GPIO pin definitions
var rowPins = []machine.Pin{
	machine.GP0, machine.GP1, machine.GP2, machine.GP3, machine.GP4, machine.GP5, machine.GP6,
}

var colPins = []machine.Pin{
	machine.GP7, machine.GP8, machine.GP9, machine.GP10, machine.GP11, machine.GP12, machine.GP13, machine.GP14, machine.GP15, machine.GP16,
}

// Special status LED and PWM pins
var (
	pinGP22 = machine.GP22 // CapsLock LED
	pinGP19 = machine.GP19 // Audio Mute
	pinGP21 = machine.GP21 // Breathing / Status LED
	pinGP20 = machine.GP20 // LCD Backlight PWM
	pinGP18 = machine.GP18 // Audio Volume PWM
)

// PWM Hardware state
var (
	blPWMChan uint8
	blPWMVal  uint32 = 32767 // Initial LCD backlight duty cycle (~50%)

	adPWMChan uint8
	adPWMVal  uint32 = 32700 // Initial audio duty cycle
)

// Default Key Matrix Map (7 rows x 10 cols)
var keyMap = [7][10]keyboard.Keycode{
	{keyboard.KeyUp, keyboard.KeyLeft, keyboard.KeyDown, keyboard.KeyRight, keyboard.KeyL, keyboard.KeyR, keyboard.KeyR, keyboard.KeyX, keyboard.KeyY, keyboard.KeyB},
	{keyboard.Key1, keyboard.Key2, keyboard.Key3, keyboard.Key4, keyboard.Key5, keyboard.Key6, keyboard.Key7, keyboard.Key8, keyboard.Key9, keyboard.Key0},
	{keyboard.KeyQ, keyboard.KeyW, keyboard.KeyE, keyboard.KeyR, keyboard.KeyT, keyboard.KeyY, keyboard.KeyU, keyboard.KeyI, keyboard.KeyO, keyboard.KeyP},
	{keyboard.KeyA, keyboard.KeyS, keyboard.KeyD, keyboard.KeyF, keyboard.KeyG, keyboard.KeyH, keyboard.KeyJ, keyboard.KeyK, keyboard.KeyL, keyboard.KeyBackspace},
	{keyboard.KeyZ, keyboard.KeyX, keyboard.KeyC, keyboard.KeyV, keyboard.KeyB, keyboard.KeyN, keyboard.KeyM, keyboard.KeySlash, keyboard.KeyEnter, 0},
	{keyboard.KeyTab, keyboard.KeyCapsLock, keyboard.KeyMinus, keyboard.KeyEqual, keyboard.KeySemicolon, keyboard.KeyQuote, keyboard.KeyComma, keyboard.KeyPeriod, keyboard.KeyLeftShift, 0},
	{keyboard.Keycode(FnKey), keyboard.KeyLeftCtrl, keyboard.KeyLeftAlt, keyboard.KeyPrintscreen, keyboard.KeySpace, keyboard.KeyPause, keyboard.KeyRightAlt, keyboard.KeyLeftGUI, keyboard.Keycode(FnKey), 0},
}

// Fn Key Matrix Map (7 rows x 10 cols)
var fnMap = [7][10]keyboard.Keycode{
	{keyboard.KeyUp, keyboard.KeyLeft, keyboard.KeyDown, keyboard.KeyRight, keyboard.KeyL, keyboard.KeyR, keyboard.KeyR, keyboard.KeyX, keyboard.KeyY, keyboard.KeyB},
	{keyboard.KeyF1, keyboard.KeyF2, keyboard.KeyF3, keyboard.KeyF4, keyboard.KeyF5, keyboard.KeyF6, keyboard.KeyF7, keyboard.KeyF8, keyboard.KeyF9, keyboard.KeyF10},
	{keyboard.Keycode(FnBLControlScreen), keyboard.KeyUp, keyboard.KeyEsc, keyboard.KeyHome, keyboard.KeyPageUp, keyboard.KeyPageDown, keyboard.KeyEnd, keyboard.KeyInsert, keyboard.KeyDelete, keyboard.KeyF11},
	{keyboard.KeyLeft, keyboard.KeyDown, keyboard.KeyRight, keyboard.KeyF12, keyboard.Keycode(FnLockScreen), keyboard.Keycode(ShiftGrave), keyboard.Keycode(ShiftBackslash), keyboard.Keycode(ShiftLeftBracket), keyboard.Keycode(ShiftRightBracket), keyboard.KeyDelete},
	{keyboard.Keycode(FnBLPWMDown), keyboard.Keycode(FnBLPWMUp), keyboard.KeyTilde, keyboard.KeyBackslash, keyboard.KeyLeftBrace, keyboard.KeyRightBrace, keyboard.KeyMenu, keyboard.KeySlash, keyboard.KeyEnter, 0},
	{keyboard.Keycode(FnMute), keyboard.KeyCapsLock, keyboard.Keycode(FnVolDown), keyboard.Keycode(FnVolUp), keyboard.KeySemicolon, keyboard.KeyQuote, keyboard.KeyComma, keyboard.KeyPeriod, keyboard.KeyLeftShift, 0},
	{keyboard.Keycode(FnKey), keyboard.KeyLeftCtrl, keyboard.KeyLeftAlt, keyboard.KeyPrintscreen, keyboard.KeySpace, keyboard.KeyPause, keyboard.KeyRightAlt, keyboard.KeyLeftGUI, keyboard.Keycode(FnKey), 0},
}

// Fixed-size matrix state trackers (zero heap allocation)
var (
	activeKeys [7][10]keyboard.Keycode
	rawScan    [7][10]bool
)

func initPWM() {
	// Backlight PWM (GP20)
	pinGP20.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.PWM2.Configure(machine.PWMConfig{Period: 1000000000 / 5000}) // 5kHz
	blPWMChan, _ = machine.PWM2.Channel(pinGP20)
	machine.PWM2.Set(blPWMChan, blPWMVal)

	// Audio PWM (GP18)
	pinGP18.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.PWM1.Configure(machine.PWMConfig{Period: 1000000000 / 5000}) // 5kHz
	adPWMChan, _ = machine.PWM1.Channel(pinGP18)
	machine.PWM1.Set(adPWMChan, adPWMVal)
}

func blPwmUp() {
	if blPWMVal >= 65535 {
		blPWMVal = 65535
	} else {
		blPWMVal += 6553
		if blPWMVal > 65535 {
			blPWMVal = 65535
		}
	}
	machine.PWM2.Set(blPWMChan, blPWMVal)
}

func blPwmDown() {
	if blPWMVal <= 6553 {
		blPWMVal = 0
	} else {
		blPWMVal -= 6553
	}
	machine.PWM2.Set(blPWMChan, blPWMVal)
}

func adPwmUp() {
	if adPWMVal >= 65535 {
		adPWMVal = 65535
	} else {
		adPWMVal += 6553
		if adPWMVal > 65535 {
			adPWMVal = 65535
		}
	}
	machine.PWM1.Set(adPWMChan, adPWMVal)
}

func adPwmDown() {
	if adPWMVal <= 6553 {
		adPWMVal = 0
	} else {
		adPWMVal -= 6553
	}
	machine.PWM1.Set(adPWMChan, adPWMVal)
}

func dispatchKeyEvent(kb *keyboard.Port, key keyboard.Keycode, pressed bool) {
	if pressed {
		switch key {
		case keyboard.KeyCapsLock:
			pinGP22.Set(!pinGP22.Get())
			kb.Down(keyboard.KeyCapsLock)
			time.Sleep(time.Millisecond * 10)
			kb.Up(keyboard.KeyCapsLock)
		case keyboard.Keycode(FnMute):
			pinGP19.Set(!pinGP19.Get())
		case keyboard.Keycode(FnBLControlScreen):
			pinGP21.Set(!pinGP21.Get())
		case keyboard.Keycode(FnBLPWMUp):
			blPwmUp()
		case keyboard.Keycode(FnBLPWMDown):
			blPwmDown()
		case keyboard.Keycode(FnVolUp):
			adPwmUp()
		case keyboard.Keycode(FnVolDown):
			adPwmDown()
		case keyboard.Keycode(ShiftGrave): // ~
			kb.Down(keyboard.KeyLeftShift)
			time.Sleep(time.Millisecond * 5)
			kb.Down(keyboard.KeyTilde)
			time.Sleep(time.Millisecond * 10)
			kb.Up(keyboard.KeyTilde)
			time.Sleep(time.Millisecond * 5)
			kb.Up(keyboard.KeyLeftShift)
		case keyboard.Keycode(ShiftBackslash): // |
			kb.Down(keyboard.KeyLeftShift)
			time.Sleep(time.Millisecond * 5)
			kb.Down(keyboard.KeyBackslash)
			time.Sleep(time.Millisecond * 10)
			kb.Up(keyboard.KeyBackslash)
			time.Sleep(time.Millisecond * 5)
			kb.Up(keyboard.KeyLeftShift)
		case keyboard.Keycode(ShiftLeftBracket): // {
			kb.Down(keyboard.KeyLeftShift)
			time.Sleep(time.Millisecond * 5)
			kb.Down(keyboard.KeyLeftBrace)
			time.Sleep(time.Millisecond * 10)
			kb.Up(keyboard.KeyLeftBrace)
			time.Sleep(time.Millisecond * 5)
			kb.Up(keyboard.KeyLeftShift)
		case keyboard.Keycode(ShiftRightBracket): // }
			kb.Down(keyboard.KeyLeftShift)
			time.Sleep(time.Millisecond * 5)
			kb.Down(keyboard.KeyRightBrace)
			time.Sleep(time.Millisecond * 10)
			kb.Up(keyboard.KeyRightBrace)
			time.Sleep(time.Millisecond * 5)
			kb.Up(keyboard.KeyLeftShift)
		case keyboard.Keycode(FnLockScreen):
			kb.Down(keyboard.KeyLeftGUI)
			time.Sleep(time.Millisecond * 5)
			kb.Down(keyboard.KeyL)
			time.Sleep(time.Millisecond * 10)
			kb.Up(keyboard.KeyL)
			time.Sleep(time.Millisecond * 5)
			kb.Up(keyboard.KeyLeftGUI)
		default:
			if key < 0x1000 {
				kb.Down(key)
			}
		}
	} else {
		if key < 0x1000 {
			kb.Up(key)
		}
	}
}

func main() {
	// 1. Initialize USB HID keyboard instance FIRST
	kb := keyboard.Port()

	// 2. Initialize GPIO pins
	for _, pin := range rowPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	}
	for _, pin := range colPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInput})
	}

	pinGP22.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinGP19.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinGP21.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinGP22.Low()
	pinGP19.Low()
	pinGP21.Low()

	// 3. Initialize PWM for LCD Backlight and Audio Volume
	initPWM()

	time.Sleep(time.Second * 1)

	// 4. Main Synchronous Loop
	for {
		// Clear raw scan matrix
		for rIdx := 0; rIdx < 7; rIdx++ {
			for cIdx := 0; cIdx < 10; cIdx++ {
				rawScan[rIdx][cIdx] = false
			}
		}

		// Tri-state hardware matrix scan
		for cIdx, cPin := range colPins {
			cPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
			cPin.High()
			time.Sleep(time.Microsecond * 50)

			for rIdx, rPin := range rowPins {
				if rPin.Get() {
					rawScan[rIdx][cIdx] = true
				}
			}

			cPin.Configure(machine.PinConfig{Mode: machine.PinInput})
		}

		// Check Fn Key state (Row 6, Col 0 or Row 6, Col 8)
		fnActive := rawScan[6][0] || rawScan[6][8]

		activeMap := &keyMap
		if fnActive {
			activeMap = &fnMap
		}

		// Update key states & dispatch USB HID events
		for rIdx := 0; rIdx < 7; rIdx++ {
			for cIdx := 0; cIdx < 10; cIdx++ {
				pressed := rawScan[rIdx][cIdx]
				currentKey := activeKeys[rIdx][cIdx]

				if pressed && currentKey == 0 {
					key := activeMap[rIdx][cIdx]
					if key != 0 {
						activeKeys[rIdx][cIdx] = key
						if key != keyboard.Keycode(FnKey) {
							dispatchKeyEvent(kb, key, true)
						}
					}
				} else if !pressed && currentKey != 0 {
					if currentKey != keyboard.Keycode(FnKey) {
						dispatchKeyEvent(kb, currentKey, false)
					}
					activeKeys[rIdx][cIdx] = 0
				}
			}
		}

		time.Sleep(time.Millisecond * 10)
	}
}
