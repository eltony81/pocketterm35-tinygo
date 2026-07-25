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

// Special status LED pins
var (
	pinGP22 = machine.GP22 // CapsLock LED
	pinGP19 = machine.GP19 // Audio Mute
	pinGP21 = machine.GP21 // Breathing / Status LED
	pinGP20 = machine.GP20 // LED PWM
	pinGP18 = machine.GP18 // Audio PWM
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
	{keyboard.KeyTab, keyboard.KeyCapsLock, keyboard.Keycode(FnVolDown), keyboard.Keycode(FnVolUp), keyboard.KeySemicolon, keyboard.KeyQuote, keyboard.KeyComma, keyboard.KeyPeriod, keyboard.KeyLeftShift, 0},
	{keyboard.Keycode(FnKey), keyboard.KeyLeftCtrl, keyboard.KeyLeftAlt, keyboard.KeyPrintscreen, keyboard.KeySpace, keyboard.KeyPause, keyboard.KeyRightAlt, keyboard.KeyLeftGUI, keyboard.Keycode(FnKey), 0},
}

type KeyEvent struct {
	Key     keyboard.Keycode
	Pressed bool
}

func main() {
	// Initialize GPIO pins
	for _, pin := range rowPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	}
	for _, pin := range colPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin.Low()
	}

	pinGP22.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinGP19.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinGP21.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinGP22.Low()
	pinGP19.Low()
	pinGP21.Low()

	keyChan := make(chan KeyEvent, 32)

	// Goroutine 1: High speed Matrix Scanner
	go matrixScanner(keyChan)

	// Main Loop: USB HID Event Dispatcher
	handleEvents(keyChan)
}

func matrixScanner(ch chan<- KeyEvent) {
	activeKeys := make(map[keyboard.Keycode]bool)

	for {
		fnActive := false
		pressedThisScan := make(map[keyboard.Keycode]bool)

		// 1. Scan for Fn Key first
		for cIdx, cPin := range colPins {
			cPin.High()
			for rIdx, rPin := range rowPins {
				if rPin.Get() {
					time.Sleep(time.Microsecond * 50)
					if rPin.Get() {
						key := keyMap[rIdx][cIdx]
						if key == keyboard.Keycode(FnKey) {
							fnActive = true
						}
					}
				}
			}
			cPin.Low()
		}

		// 2. Select active map based on Fn state
		activeMap := keyMap
		if fnActive {
			activeMap = fnMap
		}

		// 3. Scan matrix for all pressed keys
		for cIdx, cPin := range colPins {
			cPin.High()
			for rIdx, rPin := range rowPins {
				if rPin.Get() {
					time.Sleep(time.Microsecond * 50)
					if rPin.Get() {
						key := activeMap[rIdx][cIdx]
						if key != 0 {
							pressedThisScan[key] = true
						}
					}
				}
			}
			cPin.Low()
		}

		// 4. Send Key Press Events
		for key := range pressedThisScan {
			if !activeKeys[key] {
				activeKeys[key] = true
				ch <- KeyEvent{Key: key, Pressed: true}
			}
		}

		// 5. Send Key Release Events
		for key := range activeKeys {
			if !pressedThisScan[key] {
				delete(activeKeys, key)
				ch <- KeyEvent{Key: key, Pressed: false}
			}
		}

		time.Sleep(time.Millisecond * 2)
	}
}

func handleEvents(ch <-chan KeyEvent) {
	kb := keyboard.Port()
	for evt := range ch {
		key := evt.Key
		if evt.Pressed {
			switch key {
			case keyboard.KeyCapsLock:
				pinGP22.Set(!pinGP22.Get())
				kb.Down(keyboard.KeyCapsLock)
			case keyboard.Keycode(FnMute):
				pinGP19.Set(!pinGP19.Get())
			case keyboard.Keycode(FnBLControlScreen):
				pinGP21.Set(!pinGP21.Get())
			case keyboard.Keycode(FnLockScreen):
				kb.Down(keyboard.KeyLeftGUI)
				kb.Down(keyboard.KeyL)
				kb.Up(keyboard.KeyL)
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
}
