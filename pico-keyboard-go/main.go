package main

import (
	"machine"
	"machine/usb/hid/keyboard"
	"time"
)

// Custom keys for special functions matching Python CUSTOM_KEYS
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

// Hardware GPIO pin definitions matching Python row_pins & col_pins
var rowPins = []machine.Pin{
	machine.GP0, machine.GP1, machine.GP2, machine.GP3, machine.GP4, machine.GP5, machine.GP6,
}

var colPins = []machine.Pin{
	machine.GP7, machine.GP8, machine.GP9, machine.GP10, machine.GP11, machine.GP12, machine.GP13, machine.GP14, machine.GP15, machine.GP16,
}

// Special hardware pins matching Python gp22, gp19, gp21, GP20 (BL), GP18 (AD)
var (
	pinGP22 = machine.GP22 // CapsLock LED
	pinGP19 = machine.GP19 // Audio Mute
	pinGP21 = machine.GP21 // Breathing / Status LED
	pinGP20 = machine.GP20 // LCD Backlight PWM
	pinGP18 = machine.GP18 // Audio Volume PWM
)

// PWM Hardware state matching Python initial values (5000 for BL, 32700 for AD)
var (
	blPWMChan uint8
	blPWMVal  uint32 = 5000

	adPWMChan uint8
	adPWMVal  uint32 = 32700
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

func initPWM() {
	// Backlight PWM (GP20) matching Python (frequency 5000Hz)
	pinGP20.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.PWM2.Configure(machine.PWMConfig{Period: 1000000000 / 5000})
	blPWMChan, _ = machine.PWM2.Channel(pinGP20)
	machine.PWM2.Set(blPWMChan, blPWMVal)

	// Audio PWM (GP18) matching Python (frequency 5000Hz)
	pinGP18.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.PWM1.Configure(machine.PWMConfig{Period: 1000000000 / 5000})
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

// Fixed-size keycode set tracker (zero heap allocation)
type KeySet struct {
	keys [16]keyboard.Keycode
	size int
}

func (s *KeySet) Add(k keyboard.Keycode) {
	if k == 0 {
		return
	}
	for i := 0; i < s.size; i++ {
		if s.keys[i] == k {
			return
		}
	}
	if s.size < len(s.keys) {
		s.keys[s.size] = k
		s.size++
	}
}

func (s *KeySet) Contains(k keyboard.Keycode) bool {
	for i := 0; i < s.size; i++ {
		if s.keys[i] == k {
			return true
		}
	}
	return false
}

// Scans keyboard matrix matching Python scan_keyboard logic exactly
func scanKeyboard() KeySet {
	var pressedCoords [16][2]int
	pressedCount := 0
	fnActive := false

	for cIdx, cPin := range colPins {
		cPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		cPin.High()
		time.Sleep(time.Microsecond * 50)

		for rIdx, rPin := range rowPins {
			if rPin.Get() {
				// 5ms debounce verification matching Python time.sleep(0.01)
				time.Sleep(time.Millisecond * 5)
				if rPin.Get() {
					key := keyMap[rIdx][cIdx]
					if key == keyboard.Keycode(FnKey) {
						fnActive = true
					}
					if pressedCount < len(pressedCoords) {
						pressedCoords[pressedCount][0] = rIdx
						pressedCoords[pressedCount][1] = cIdx
						pressedCount++
					}
				}
			}
		}

		cPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	}

	activeMap := &keyMap
	if fnActive {
		activeMap = &fnMap
	}

	var currentSet KeySet
	for i := 0; i < pressedCount; i++ {
		rIdx := pressedCoords[i][0]
		cIdx := pressedCoords[i][1]
		translatedKey := activeMap[rIdx][cIdx]
		currentSet.Add(translatedKey)
	}

	return currentSet
}

func executeSpecialKey(key keyboard.Keycode) {
	kb := keyboard.Port()
	switch key {
	case keyboard.KeyCapsLock:
		pinGP22.Set(!pinGP22.Get())
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
		time.Sleep(time.Millisecond * 10)
		kb.Down(keyboard.KeyTilde)
		time.Sleep(time.Millisecond * 20)
		kb.Up(keyboard.KeyTilde)
		time.Sleep(time.Millisecond * 10)
		kb.Up(keyboard.KeyLeftShift)
	case keyboard.Keycode(ShiftBackslash): // |
		kb.Down(keyboard.KeyLeftShift)
		time.Sleep(time.Millisecond * 10)
		kb.Down(keyboard.KeyBackslash)
		time.Sleep(time.Millisecond * 20)
		kb.Up(keyboard.KeyBackslash)
		time.Sleep(time.Millisecond * 10)
		kb.Up(keyboard.KeyLeftShift)
	case keyboard.Keycode(ShiftLeftBracket): // {
		kb.Down(keyboard.KeyLeftShift)
		time.Sleep(time.Millisecond * 10)
		kb.Down(keyboard.KeyLeftBrace)
		time.Sleep(time.Millisecond * 20)
		kb.Up(keyboard.KeyLeftBrace)
		time.Sleep(time.Millisecond * 10)
		kb.Up(keyboard.KeyLeftShift)
	case keyboard.Keycode(ShiftRightBracket): // }
		kb.Down(keyboard.KeyLeftShift)
		time.Sleep(time.Millisecond * 10)
		kb.Down(keyboard.KeyRightBrace)
		time.Sleep(time.Millisecond * 20)
		kb.Up(keyboard.KeyRightBrace)
		time.Sleep(time.Millisecond * 10)
		kb.Up(keyboard.KeyLeftShift)
	case keyboard.Keycode(FnLockScreen):
		kb.Down(keyboard.KeyLeftGUI)
		time.Sleep(time.Millisecond * 10)
		kb.Down(keyboard.KeyL)
		time.Sleep(time.Millisecond * 20)
		kb.Up(keyboard.KeyL)
		time.Sleep(time.Millisecond * 10)
		kb.Up(keyboard.KeyLeftGUI)
	}
}

func main() {
	// 1. Initialize USB HID keyboard FIRST
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

	// 3. Initialize PWM for LCD Backlight and Audio Volume matching Python
	initPWM()

	time.Sleep(time.Second * 1)

	var previousSet KeySet

	// Main loop matching Python while True exactly
	for {
		currentSet := scanKeyboard()

		// Release keys that were in previousSet but not in currentSet
		for i := 0; i < previousSet.size; i++ {
			key := previousSet.keys[i]
			if !currentSet.Contains(key) {
				if key < 0x1000 {
					kb.Up(key)
				}
			}
		}

		// Press new keys that are in currentSet but were not in previousSet
		for i := 0; i < currentSet.size; i++ {
			key := currentSet.keys[i]
			if !previousSet.Contains(key) {
				if key >= 0x1000 {
					executeSpecialKey(key)
				} else {
					kb.Down(key)
				}
			}
		}

		previousSet = currentSet
		time.Sleep(time.Millisecond * 10)
	}
}
