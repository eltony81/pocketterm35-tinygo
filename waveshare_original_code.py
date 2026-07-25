import usb_hid
from hid.keyboard import Keyboard
from hid.keycode import Keycode
from hid.consumer_control import ConsumerControl
from hid.consumer_control_code import ConsumerControlCode
from hid.keyboard_layout_us import KeyboardLayoutUS
import board
import microcontroller
import digitalio
import time
import pwmio

bl_pwm_value = 32767
ad_pwm_value = 32767

try:
    kbd = Keyboard(usb_hid.devices)
    layout = KeyboardLayoutUS(kbd)
    consumer_control = ConsumerControl(usb_hid.devices)
except Exception as e:
    print(f"Initialization failed: {e}")
    microcontroller.reset()

gp22 = digitalio.DigitalInOut(board.GP22)
gp22.direction = digitalio.Direction.OUTPUT
gp22.value = False

gp19 = digitalio.DigitalInOut(board.GP19)
gp19.direction = digitalio.Direction.OUTPUT
gp19.value = False

gp21 = digitalio.DigitalInOut(board.GP21)
gp21.direction = digitalio.Direction.OUTPUT
gp21.value = False

BL_PWM_RP = pwmio.PWMOut(board.GP20, frequency=5000, duty_cycle=5000)
AD_PWM_RP = pwmio.PWMOut(board.GP18, frequency=5000, duty_cycle=32700)

def toggle_gp22():
    gp22.value = not gp22.value
    print(f"GP22 state toggled to: {gp22.value}")
    kbd.press(Keycode.CAPS_LOCK)
    kbd.release(Keycode.CAPS_LOCK)

def start_breathing_light():
    tmp_time = 0
    stop_loop = True
    kbd = None
    gp22_temp = gp22.value

    for i in range(1, 11):
        gp22.value = not gp22.value
        time.sleep(0.3)

    try:
        while stop_loop:
            tmp_time += 1
            if tmp_time >= 255:
                gp22.value = not gp22.value
                tmp_time = 0
            current_keys = set(scan_keyboard())
            for key in current_keys:
                if key is not None:
                    time.sleep(0.2)
                    if key is not None:
                        if kbd:
                            kbd.release_all()
                        gp22.value = gp22_temp
                        gp21.value = not gp21.value
                        stop_loop = False
                        break
    finally:
        if kbd:
            kbd.release_all()
            gp22.value = gp22_temp

def toggle_gp19():
    gp19.value = not gp19.value

def toggle_gp21():
    gp21.value = not gp21.value
    start_breathing_light()

def ad_pwm_down():
    global ad_pwm_value
    if ad_pwm_value <= 0:
        ad_pwm_value = 0
    else:
        ad_pwm_value -= 6553
    ad_pwm_value = min(max(ad_pwm_value, 0), 65535)
    AD_PWM_RP.duty_cycle = ad_pwm_value
    print("pwm_value == ", ad_pwm_value)

def ad_pwm_up():
    global ad_pwm_value
    if ad_pwm_value >= 65535:
        ad_pwm_value = 65535
    else:
        ad_pwm_value += 6553
    ad_pwm_value = min(max(ad_pwm_value, 0), 65535)
    AD_PWM_RP.duty_cycle = ad_pwm_value
    print("pwm_value == ", ad_pwm_value)

def bl_pwm_up():
    global bl_pwm_value
    if bl_pwm_value <= 0:
        bl_pwm_value = 0
    else:
        bl_pwm_value -= 6553
    bl_pwm_value = min(max(bl_pwm_value, 0), 65535)
    BL_PWM_RP.duty_cycle = bl_pwm_value
    print("pwm_value == ", bl_pwm_value)

def bl_pwm_down():
    global bl_pwm_value
    if bl_pwm_value >= 65535:
        bl_pwm_value = 65535
    else:
        bl_pwm_value += 6553
    bl_pwm_value = min(max(bl_pwm_value, 0), 65535)
    BL_PWM_RP.duty_cycle = bl_pwm_value
    print("pwm_value == ", bl_pwm_value)

def shift_left_bracket():
    kbd.press(Keycode.SHIFT, Keycode.LEFT_BRACKET)
    kbd.release_all()

def shift_right_bracket():
    kbd.press(Keycode.SHIFT, Keycode.RIGHT_BRACKET)
    kbd.release_all()

def shift_backslash():
    kbd.press(Keycode.SHIFT, Keycode.BACKSLASH)
    kbd.release_all()

def shift_grave_accent():
    kbd.press(Keycode.SHIFT, Keycode.GRAVE_ACCENT)
    kbd.release_all()

def lock_screen():
    kbd.press(Keycode.WINDOWS, Keycode.L)
    kbd.release_all()

def scan_previous_track():
    consumer_control.press(ConsumerControlCode.SCAN_PREVIOUS_TRACK)
    consumer_control.release()

def play_pause():
    consumer_control.press(ConsumerControlCode.PLAY_PAUSE)
    consumer_control.release()

def scan_next_track():
    consumer_control.press(ConsumerControlCode.SCAN_NEXT_TRACK)
    consumer_control.release()

CUSTOM_KEYS = {
    "FN_KEY"                : -100,
    "FN_MUTE"               : -101,
    "FN_VOLUME_DOWN"        : -102,
    "FN_VOLUME_UP"          : -103,
    "FN_LOCK_SCREEN"        : -104,
    "FN_BL_CONTROL_SCREEN"  : -105,
    "FN_BL_PWM_DOWN"        : -106,
    "FN_BL_PWM_UP"          : -107,
    "SHIFT_GRAVE_ACCENT"    : -108,
    "SHIFT_BACKSLASH"       : -109,
    "SHIFT_LEFT_BRACKET"    : -110,
    "SHIFT_RIGHT_BRACKET"   : -111
}

SPECIAL_KEY_FUNCTIONS = {
    CUSTOM_KEYS["FN_MUTE"]                  : toggle_gp19,
    CUSTOM_KEYS["FN_VOLUME_DOWN"]           : ad_pwm_down,
    CUSTOM_KEYS["FN_VOLUME_UP"]             : ad_pwm_up,
    CUSTOM_KEYS["FN_LOCK_SCREEN"]           : lock_screen,
    CUSTOM_KEYS["FN_BL_CONTROL_SCREEN"]     : toggle_gp21,
    CUSTOM_KEYS["FN_BL_PWM_DOWN"]           : bl_pwm_down,
    CUSTOM_KEYS["FN_BL_PWM_UP"]             : bl_pwm_up,
    CUSTOM_KEYS["SHIFT_GRAVE_ACCENT"]       : shift_grave_accent,
    CUSTOM_KEYS["SHIFT_BACKSLASH"]          : shift_backslash,
    CUSTOM_KEYS["SHIFT_LEFT_BRACKET"]       : shift_left_bracket,
    CUSTOM_KEYS["SHIFT_RIGHT_BRACKET"]      : shift_right_bracket,
    Keycode.CAPS_LOCK                       : toggle_gp22,
    ConsumerControlCode.SCAN_PREVIOUS_TRACK : scan_previous_track,
    ConsumerControlCode.PLAY_PAUSE          : play_pause,
    ConsumerControlCode.SCAN_NEXT_TRACK     : scan_next_track
}

NUM_ROWS = 7
NUM_COLS = 10

KEY_MAP = [
    [Keycode.UP_ARROW,          Keycode.LEFT_ARROW,                 Keycode.DOWN_ARROW,                 Keycode.RIGHT_ARROW,                Keycode.L,                          Keycode.R,                              Keycode.R,                              Keycode.X,                              Keycode.Y,                          Keycode.B,              Keycode.A],
    [Keycode.ONE,               Keycode.TWO,                        Keycode.THREE,                      Keycode.FOUR,                       Keycode.FIVE,                       Keycode.SIX,                            Keycode.SEVEN,                          Keycode.EIGHT,                          Keycode.NINE,                       Keycode.ZERO],
    [Keycode.Q,                 Keycode.W,                          Keycode.E,                          Keycode.R,                          Keycode.T,                          Keycode.Y,                              Keycode.U,                              Keycode.I,                              Keycode.O,                          Keycode.P],
    [Keycode.A,                 Keycode.S,                          Keycode.D,                          Keycode.F,                          Keycode.G,                          Keycode.H,                              Keycode.J,                              Keycode.K,                              Keycode.L,                          Keycode.BACKSPACE],
    [Keycode.Z,                 Keycode.X,                          Keycode.C,                          Keycode.V,                          Keycode.B,                          Keycode.N,                              Keycode.M,                              Keycode.FORWARD_SLASH,                  Keycode.ENTER,                      None],
    [Keycode.TAB,               Keycode.CAPS_LOCK,                  Keycode.MINUS,                      Keycode.EQUALS,                     Keycode.SEMICOLON,                  Keycode.QUOTE,                          Keycode.COMMA,                          Keycode.PERIOD,                         Keycode.SHIFT,                      None],
    [CUSTOM_KEYS["FN_KEY"],     Keycode.CONTROL,                    Keycode.LEFT_ALT,                   Keycode.PRINT_SCREEN,               Keycode.SPACE,                      Keycode.PAUSE,                          Keycode.RIGHT_ALT,                      Keycode.WINDOWS,                        CUSTOM_KEYS["FN_KEY"],              None]
]

FN_MAP = [
    [Keycode.UP_ARROW,          Keycode.LEFT_ARROW,                 Keycode.DOWN_ARROW,                 Keycode.RIGHT_ARROW,                Keycode.L,                          Keycode.R,                              Keycode.R,                              Keycode.X,                              Keycode.Y,                          Keycode.B,              Keycode.A],
    [Keycode.F1,                Keycode.F2,                         Keycode.F3,                         Keycode.F4,                         Keycode.F5,                         Keycode.F6,                             Keycode.F7,                             Keycode.F8,                             Keycode.F9,                         Keycode.F10],
    [CUSTOM_KEYS["FN_BL_CONTROL_SCREEN"],Keycode.UP_ARROW,          Keycode.ESCAPE,                     Keycode.HOME,                       Keycode.PAGE_UP,                    Keycode.PAGE_DOWN,                      Keycode.END,                            Keycode.INSERT,                         Keycode.DELETE,                     Keycode.F11],
    [Keycode.LEFT_ARROW,        Keycode.DOWN_ARROW,                 Keycode.RIGHT_ARROW,                Keycode.F12,                        CUSTOM_KEYS["FN_LOCK_SCREEN"],      CUSTOM_KEYS["SHIFT_GRAVE_ACCENT"],      CUSTOM_KEYS["SHIFT_BACKSLASH"],         CUSTOM_KEYS["SHIFT_LEFT_BRACKET"],      CUSTOM_KEYS["SHIFT_RIGHT_BRACKET"], Keycode.DELETE],
    [CUSTOM_KEYS["FN_BL_PWM_DOWN"],CUSTOM_KEYS["FN_BL_PWM_UP"],    Keycode.GRAVE_ACCENT,               Keycode.BACKSLASH,                  Keycode.LEFT_BRACKET,               Keycode.RIGHT_BRACKET,                  Keycode.APPLICATION,                    Keycode.FORWARD_SLASH,                  Keycode.ENTER,                      None],
    [Keycode.TAB,               Keycode.CAPS_LOCK,                  CUSTOM_KEYS["FN_VOLUME_DOWN"],      CUSTOM_KEYS["FN_VOLUME_UP"],        Keycode.SEMICOLON,                  Keycode.QUOTE,                          Keycode.COMMA,                          Keycode.PERIOD,                         Keycode.SHIFT,                      None],
    [CUSTOM_KEYS["FN_KEY"],     Keycode.CONTROL,                    Keycode.LEFT_ALT,                   Keycode.PRINT_SCREEN,               Keycode.SPACE,                      Keycode.PAUSE,                          Keycode.RIGHT_ALT,                      Keycode.WINDOWS,                        CUSTOM_KEYS["FN_KEY"],              None]
]

row_pins = [board.GP0, board.GP1, board.GP2, board.GP3, board.GP4, board.GP5, board.GP6]
col_pins = [board.GP7, board.GP8, board.GP9, board.GP10, board.GP11, board.GP12, board.GP13, board.GP14, board.GP15, board.GP16]

row_gpio = [digitalio.DigitalInOut(pin) for pin in row_pins]
col_gpio = [digitalio.DigitalInOut(pin) for pin in col_pins]

def scan_keyboard():
    keys_pressed = []
    fn_active = False

    for col_index, col_pin in enumerate(col_gpio):
        col_pin.direction = digitalio.Direction.OUTPUT
        col_pin.value = True

        for row_index, row_pin in enumerate(row_gpio):
            row_pin.direction = digitalio.Direction.INPUT
            row_pin.pull = digitalio.Pull.DOWN

            if row_pin.value:
                time.sleep(0.01)
                if row_pin.value:
                    key = KEY_MAP[row_index][col_index]
                    if key == CUSTOM_KEYS["FN_KEY"]:
                        fn_active = True
                    keys_pressed.append((row_index, col_index))

        col_pin.value = False

    active_map = FN_MAP if fn_active else KEY_MAP

    translated_keys = []
    for row_index, col_index in keys_pressed:
        translated_key = active_map[row_index][col_index]
        translated_keys.append(translated_key)
        print("ROW", row_index, "COL", col_index)

    return translated_keys

previous_keys = set()

try:
    while True:
        current_keys = set(scan_keyboard())

        for key in previous_keys - current_keys:
            if key is not None and key >= 0:
                kbd.release(key)

        for key in current_keys - previous_keys:
            if key is not None:
                if key in SPECIAL_KEY_FUNCTIONS:
                    SPECIAL_KEY_FUNCTIONS[key]()
                elif key >= 0:
                    kbd.press(key)

        previous_keys = current_keys

        time.sleep(0.01)

except Exception as e:
    print(f"An error occurred: {e}")
    try:
        kbd.release_all()
    except Exception as cleanup_error:
        print(f"Error during cleanup: {cleanup_error}")
    microcontroller.reset()
