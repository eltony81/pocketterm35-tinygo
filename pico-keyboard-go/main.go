package main

import (
	"fmt"
	"machine"
	"machine/usb/hid/keyboard"
	"strings"
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

type KeyEvent struct {
	Key     keyboard.Keycode
	Pressed bool
}

// Global control channels
var (
	keyChan       = make(chan KeyEvent, 64)
	breathingChan = make(chan bool, 1)
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

func getPicoTemperature() float32 {
	milliC := machine.ReadTemperature()
	temp := float32(milliC) / 1000.0
	// RP2040 internal sensor VREF offset calibration
	if temp < 0 {
		temp += 35.0
	}
	return temp
}

func main() {
	// Initialize ADC
	machine.InitADC()

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

	// Initialize PWM for LCD Backlight and Audio Volume
	initPWM()

	// Goroutine 1: High speed zero-allocation Matrix Scanner
	go matrixScanner(keyChan)

	// Goroutine 2: Non-blocking Serial Listener for Commands / Temperature
	go serialListener()

	// Goroutine 3: Breathing Light Handler
	go breathingController(breathingChan)

	// Main Loop: USB HID Event Dispatcher
	handleEvents(keyChan)
}

// High-speed, zero-heap-allocation matrix scanner
func matrixScanner(ch chan<- KeyEvent) {
	var rawScan [7][10]bool
	var stateMatrix [7][10]bool

	for {
		// 1. Single-pass hardware matrix scan
		for cIdx, cPin := range colPins {
			cPin.High()
			time.Sleep(time.Microsecond * 10) // Brief pin settling time
			for rIdx, rPin := range rowPins {
				rawScan[rIdx][cIdx] = rPin.Get()
			}
			cPin.Low()
		}

		// 2. Check Fn Key state (Row 6, Col 0 or Row 6, Col 8)
		fnActive := rawScan[6][0] || rawScan[6][8]

		// 3. Select active key map
		activeMap := &keyMap
		if fnActive {
			activeMap = &fnMap
		}

		// 4. Compare current hardware scan against previous state matrix
		for rIdx := 0; rIdx < 7; rIdx++ {
			for cIdx := 0; cIdx < 10; cIdx++ {
				pressed := rawScan[rIdx][cIdx]
				if pressed != stateMatrix[rIdx][cIdx] {
					stateMatrix[rIdx][cIdx] = pressed
					key := activeMap[rIdx][cIdx]
					if key != 0 {
						ch <- KeyEvent{Key: key, Pressed: pressed}
						// Stop breathing animation on any keypress
						if pressed {
							select {
							case breathingChan <- false:
							default:
							}
						}
					}
				}
			}
		}

		time.Sleep(time.Millisecond * 2)
	}
}

// Non-blocking breathing light controller
func breathingController(controlChan <-chan bool) {
	active := false
	for {
		select {
		case state := <-controlChan:
			active = state
			if !active {
				pinGP22.Low()
			}
		default:
		}

		if active {
			// Blinking animation on CapsLock LED (GP22)
			pinGP22.High()
			time.Sleep(time.Millisecond * 250)
			pinGP22.Low()
			time.Sleep(time.Millisecond * 250)
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func serialListener() {
	var buf [64]byte
	bufIdx := 0

	for {
		if machine.Serial.Buffered() > 0 {
			b, err := machine.Serial.ReadByte()
			if err == nil {
				if b == 0x0A || b == 0x0D {
					if bufIdx > 0 {
						cmd := strings.TrimSpace(string(buf[:bufIdx]))
						bufIdx = 0
						cmdUpper := strings.ToUpper(cmd)
						switch cmdUpper {
						case "TEMP", "PICO_TEMP", "STATUS":
							temp := getPicoTemperature()
							fmt.Printf("PICO_TEMP: %.2f C\r\n", temp)
						case "BRIGHTNESS:UP":
							blPwmUp()
							fmt.Printf("BRIGHTNESS: %d\r\n", blPWMVal)
						case "BRIGHTNESS:DOWN":
							blPwmDown()
							fmt.Printf("BRIGHTNESS: %d\r\n", blPWMVal)
						case "VOL:UP":
							adPwmUp()
							fmt.Printf("VOL: %d\r\n", adPWMVal)
						case "VOL:DOWN":
							adPwmDown()
							fmt.Printf("VOL: %d\r\n", adPWMVal)
						case "HELP":
							fmt.Printf("COMMANDS: TEMP, BRIGHTNESS:UP, BRIGHTNESS:DOWN, VOL:UP, VOL:DOWN, STATUS\r\n")
						}
					}
				} else if bufIdx < len(buf)-1 {
					buf[bufIdx] = b
					bufIdx++
				}
			}
		}
		time.Sleep(time.Millisecond * 10)
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
				select {
				case breathingChan <- true:
				default:
				}
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
				kb.Down(keyboard.KeyTilde)
				kb.Up(keyboard.KeyTilde)
				kb.Up(keyboard.KeyLeftShift)
			case keyboard.Keycode(ShiftBackslash): // |
				kb.Down(keyboard.KeyLeftShift)
				kb.Down(keyboard.KeyBackslash)
				kb.Up(keyboard.KeyBackslash)
				kb.Up(keyboard.KeyLeftShift)
			case keyboard.Keycode(ShiftLeftBracket): // {
				kb.Down(keyboard.KeyLeftShift)
				kb.Down(keyboard.KeyLeftBrace)
				kb.Up(keyboard.KeyLeftBrace)
				kb.Up(keyboard.KeyLeftShift)
			case keyboard.Keycode(ShiftRightBracket): // }
				kb.Down(keyboard.KeyLeftShift)
				kb.Down(keyboard.KeyRightBrace)
				kb.Up(keyboard.KeyRightBrace)
				kb.Up(keyboard.KeyLeftShift)
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
