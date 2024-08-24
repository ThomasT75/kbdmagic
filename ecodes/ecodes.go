package ecodes

import (
	"kbdmagic/internal/log"

	"github.com/grafov/evdev"
)

// ecodes types

// Normalize the input from the underlying system
// it should be given by a function in ecodes.FromEvdev
type NormalizedIndex int

// A normalized index who has a state bound to it 
//
// ex input with NormalizedIndex of 1 and value (aka state) of 0 
// would give a different StatedIndex than if the value was 1 with 
// the same NormalizedIndex as before
// 
// state can be any number but only positive or negative is taken into account
// 0 is counted as negative
type StatedIndex int

const MAX_STATED_INDEX = ECODES_MAX * 2

func (si StatedIndex) String() string {
  var text string = "RELEASED: "
  if IsStateIndexOn(si) {
    text = "PRESSED: "
  }

  e, t := toEvdev(toNormalizedIndex(si))
  switch t {
  case evdev.EV_KEY:
    if e == 0 {
      text = "NoKey"
      goto PRINT
    }
    s, ok := evdev.KEY[int(e)]
    if ok {
      text += s
      goto PRINT
    }
    s, ok = evdev.BTN[int(e)]
    if ok {
      text += s
      goto PRINT
    }
  case evdev.EV_REL:
    s, ok := evdev.REL[int(e)]
    if ok {
      text += s
      goto PRINT
    }
  }
  text += "NoKey"
  PRINT:
  return text
}

//evdev doesn't export ecodes map so we do this for ok check
var REL_MAP = map[string]int{}
var KEY_MAP = map[string]int{}
var ABS_MAP = map[string]int{}
var BTN_MAP = map[string]int{}

var GP_BTN_MAP = map[string]int{
  //DEFAULT BINDS
  "GP_BTN_NORTH": evdev.BTN_Y, 
  "GP_BTN_WEST": evdev.BTN_X,
  "GP_BTN_SOUTH": evdev.BTN_A,
  "GP_BTN_EAST": evdev.BTN_B,
  "GP_BTN_START": evdev.BTN_START,
  "GP_BTN_SELECT": evdev.BTN_SELECT,
  "GP_BTN_MODE": evdev.BTN_MODE,
  "GP_BTN_TL": evdev.BTN_TL,
  "GP_BTN_TL2": evdev.BTN_TL2,
  "GP_BTN_TR": evdev.BTN_TR,
  "GP_BTN_TR2": evdev.BTN_TR2,
  "GP_BTN_DPAD_UP": evdev.BTN_DPAD_UP,
  "GP_BTN_DPAD_DOWN": evdev.BTN_DPAD_DOWN,
  "GP_BTN_DPAD_LEFT": evdev.BTN_DPAD_LEFT,
  "GP_BTN_DPAD_RIGHT": evdev.BTN_DPAD_RIGHT,
  "GP_BTN_THUMBL": evdev.BTN_THUMBL,
  "GP_BTN_THUMBR": evdev.BTN_THUMBR,
  // DPAD
  "GP_DPAD_UP": evdev.BTN_DPAD_UP,
  "GP_DPAD_DOWN": evdev.BTN_DPAD_DOWN,
  "GP_DPAD_LEFT": evdev.BTN_DPAD_LEFT,
  "GP_DPAD_RIGHT": evdev.BTN_DPAD_RIGHT,
  //XBOX ONE/360
  "GP_BTN_X": evdev.BTN_X,
  "GP_BTN_Y": evdev.BTN_Y,
  "GP_BTN_B": evdev.BTN_B,
  "GP_BTN_A": evdev.BTN_A,
  "GP_BTN_XBOX_START": evdev.BTN_START,
  "GP_BTN_XBOX_MENU": evdev.BTN_START,
  "GP_BTN_MENU": evdev.BTN_START,
  "GP_BTN_XBOX_VIEW": evdev.BTN_SELECT,
  "GP_BTN_VIEW": evdev.BTN_SELECT,
  "GP_BTN_XBOX_BACK": evdev.BTN_SELECT,
  "GP_BTN_BACK": evdev.BTN_SELECT,
  "GP_BTN_XBOX_BUTTON": evdev.BTN_MODE,
  "GP_BTN_XBOX_GUIDE": evdev.BTN_MODE,
  "GP_BTN_LB": evdev.BTN_TL,
  "GP_BTN_LT": evdev.BTN_TL2,
  "GP_BTN_RB": evdev.BTN_TR,
  "GP_BTN_RT": evdev.BTN_TR2,
  "GP_BTN_LSB": evdev.BTN_THUMBL,
  "GP_BTN_RSB": evdev.BTN_THUMBR,
  "GP_BTN_LS": evdev.BTN_THUMBL,
  "GP_BTN_RS": evdev.BTN_THUMBR,
  //DS4 we don't support everything a ds4 does btw ex: share button/touch pos/motion sensor etc
  "GP_BTN_SQUARE": evdev.BTN_X,
  "GP_BTN_TRIANGLE": evdev.BTN_Y,
  "GP_BTN_CIRCLE": evdev.BTN_B,
  "GP_BTN_CROSS": evdev.BTN_A,
  "GP_BTN_DS4_OPTIONS": evdev.BTN_START,
  "GP_BTN_DS4_TOUCH_PAD": evdev.BTN_SELECT,
  "GP_BTN_OPTIONS": evdev.BTN_START,
  "GP_BTN_PS_OPTIONS": evdev.BTN_START,
  "GP_BTN_PS_TOUCH_PAD": evdev.BTN_SELECT,
  "GP_BTN_PS_BUTTON": evdev.BTN_MODE,
  "GP_BTN_L1": evdev.BTN_TL,
  "GP_BTN_L2": evdev.BTN_TL2,
  "GP_BTN_R1": evdev.BTN_TR,
  "GP_BTN_R2": evdev.BTN_TR2,
  "GP_BTN_L3": evdev.BTN_THUMBL,
  "GP_BTN_R3": evdev.BTN_THUMBR,
  // nintendo switch = ns
  "GP_BTN_NS_X": evdev.BTN_Y,
  "GP_BTN_NS_Y": evdev.BTN_X,
  "GP_BTN_NS_B": evdev.BTN_A,
  "GP_BTN_NS_A": evdev.BTN_B,
  "GP_BTN_HOME": evdev.BTN_MODE,
  "GP_BTN_NS_BUTTON": evdev.BTN_MODE,
  "GP_BTN_NS_HOME": evdev.BTN_MODE,
  "GP_BTN_L": evdev.BTN_TL,
  "GP_BTN_ZL": evdev.BTN_TL2,
  "GP_BTN_R": evdev.BTN_TR,
  "GP_BTN_ZR": evdev.BTN_TR2,
  "GP_BTN_PLUS": evdev.BTN_START,
  "GP_BTN_MINUS": evdev.BTN_SELECT,
}

var _GP_BTN_MAP_STRING = map[int]string {
  evdev.BTN_Y: "GP_BTN_NORTH",
  evdev.BTN_X: "GP_BTN_WEST",
  evdev.BTN_A: "GP_BTN_SOUTH",
  evdev.BTN_B: "GP_BTN_EAST",
  evdev.BTN_START: "GP_BTN_START",
  evdev.BTN_SELECT: "GP_BTN_SELECT",
  evdev.BTN_MODE: "GP_BTN_MODE",
  evdev.BTN_TL: "GP_BTN_TL",
  evdev.BTN_TL2: "GP_BTN_TL2",
  evdev.BTN_TR: "GP_BTN_TR",
  evdev.BTN_TR2: "GP_BTN_TR2",
  evdev.BTN_DPAD_UP: "GP_BTN_DPAD_UP",
  evdev.BTN_DPAD_DOWN: "GP_BTN_DPAD_DOWN",
  evdev.BTN_DPAD_LEFT: "GP_BTN_DPAD_LEFT",
  evdev.BTN_DPAD_RIGHT: "GP_BTN_DPAD_RIGHT",
  evdev.BTN_THUMBL: "GP_BTN_THUMBL",
  evdev.BTN_THUMBR: "GP_BTN_THUMBR",
}

func GP_STRING(c int) (string, bool) {
  s, ok := _GP_BTN_MAP_STRING[c]
  if ok {
    return s, true
  }
  return "", false
}

const (
  ABS_X_NEGATIVE = -evdev.ABS_X*2 - 2
  ABS_Y_NEGATIVE = -evdev.ABS_Y*2 - 2
  ABS_RX_NEGATIVE = -evdev.ABS_RX*2 - 2
  ABS_RY_NEGATIVE = -evdev.ABS_RY*2 - 2

  ABS_X_POSITIVE = -evdev.ABS_X*2 - 1
  ABS_Y_POSITIVE = -evdev.ABS_Y*2 - 1
  ABS_RX_POSITIVE = -evdev.ABS_RX*2 - 1
  ABS_RY_POSITIVE = -evdev.ABS_RY*2 - 1

  ABS_X = evdev.ABS_X // 0
  ABS_Y = evdev.ABS_Y // 1
  ABS_RX = evdev.ABS_RX // 3
  ABS_RY = evdev.ABS_RY // 4
)


var GP_AXIS_MAP = map[string]int{
  "GP_AXIS_X_NEGATIVE": ABS_X_NEGATIVE,
  "GP_AXIS_Y_NEGATIVE": ABS_Y_NEGATIVE,
  "GP_AXIS_RX_NEGATIVE": ABS_RX_NEGATIVE,
  "GP_AXIS_RY_NEGATIVE": ABS_RY_NEGATIVE,
  "GP_AXIS_X_POSITIVE": ABS_X_POSITIVE,
  "GP_AXIS_Y_POSITIVE": ABS_Y_POSITIVE,
  "GP_AXIS_RX_POSITIVE": ABS_RX_POSITIVE,
  "GP_AXIS_RY_POSITIVE": ABS_RY_POSITIVE,
  "GP_AXIS_X": ABS_X,
  "GP_AXIS_Y": ABS_Y,
  "GP_AXIS_RX": ABS_RX,
  "GP_AXIS_RY": ABS_RY,
}

func init() {
  for k, v := range evdev.REL {
    REL_MAP[v] = k
  }
  for k, v := range evdev.KEY {
    KEY_MAP[v] = k
  }
  for k, v := range evdev.ABS {
    ABS_MAP[v] = k
  }
  for k, v := range evdev.BTN {
    BTN_MAP[v] = k
  }
  //some keys have a race codition in evdev mapping
  //basically they try to map the they name into the same int
  //and which one gets there last depends
  BTN_MAP["BTN_LEFT"] = BTN_LEFT
  BTN_MAP["BTN_MOUSE"] = BTN_MOUSE
  BTN_MAP["BTN_A"] = BTN_A
  BTN_MAP["BTN_GAMEPAD"] = BTN_GAMEPAD
  BTN_MAP["BTN_SOUTH"] = BTN_SOUTH
  BTN_MAP["BTN_EAST"] = BTN_EAST
  BTN_MAP["BTN_B"] = BTN_B
  BTN_MAP["BTN_NORTH"] = BTN_NORTH
  BTN_MAP["BTN_X"] = BTN_X
  BTN_MAP["BTN_WEST"] = BTN_WEST
  BTN_MAP["BTN_Y"] = BTN_Y
  BTN_MAP["BTN_MISC"] = BTN_MISC
  BTN_MAP["BTN_0"] = BTN_0

  KEY_MAP["KEY_SCREENLOCK"] = KEY_SCREENLOCK
  KEY_MAP["KEY_COFFEE"] = KEY_COFFEE
}

//for speed i don't check if etype is of type EV_REL or EV_KEY else it will break
func NormalizeFromInputSys(ecode uint16, etype uint16) NormalizedIndex {
  var i int = int(ecode)
  if etype == evdev.EV_REL {
    i += KEY_MAX + 1
  } 
  return NormalizedIndex(i)
}

//don't use it outside of String and formating reasons
func toEvdev(ni NormalizedIndex) (ecode uint16, etype uint16) {
  ecode = uint16(ni)
  etype = evdev.EV_KEY
  if ni - (KEY_MAX + 1) >= 0 {
    ecode = uint16(ni - (KEY_MAX + 1))
    etype = evdev.EV_REL
  }

  return ecode, etype
}

// basically convert 1 number and 2 states (on or off) into 1 unique number
func ToStateIndex(index NormalizedIndex, evalue int32) StatedIndex {
  si := index * 2
  if evalue <= 0 {
    si = si - 1
  }
  return StatedIndex(si)
}

//don't use it outside of String and formating reasons
func toNormalizedIndex(si StatedIndex) NormalizedIndex {
  return NormalizedIndex((si / 2) + (si % 2))
}

func IsStateIndexOn(stateIndex StatedIndex) bool {
  return stateIndex % 2 == 0 
}

// for use with MapGP*ToIndex() functions
// there are 15 buttons: Triggers + C and Z buttons
// there are 4 directions for a dpad
// there are 12 axis directions: X, Y, X-, X+, Y-, Y+ for each stick

const GP_INDEX_BASE int = 0
const GP_BTN_INDEX_BASE int = GP_INDEX_BASE // as in first valid index
const GP_BTN_INDEX_MAX int = GP_BTN_INDEX_BASE + 15 // as in len() 
const GP_DPAD_INDEX_BASE int = GP_BTN_INDEX_MAX 
const GP_DPAD_INDEX_MAX int = GP_DPAD_INDEX_BASE + 4
const GP_AXIS_INDEX_BASE int = GP_DPAD_INDEX_MAX
const GP_AXIS_INDEX_MAX int = GP_AXIS_INDEX_BASE + 12
const GP_INDEX_MAX int = GP_AXIS_INDEX_MAX

// don't rely on this function for permanent indexing
// use GP_INDEX_MAX to know the max index this function produces
func MapGPToIndex(index int) int {
  // the following is valid input
  // btn range
  // 0x130 to 0x13e 
  // dpad range
  // 0x220 to 0x223 
  // axis ranges 
  // 0x00, 0x01 
  // 0x03, 0x04
  // -4 to -1
  // -10 to -7

  switch {
  case index >= 0x130 && index <= 0x13e:
    // index + to alignment + base
    return index - 0x130 + GP_BTN_INDEX_BASE
  case index >= 0x220 && index <= 0x223:
    // index + to alignment + base
    return index - 0x220 + GP_DPAD_INDEX_BASE
  case index == 0x00, index == 0x01:
    // index + base // joys of being already aligned
    return index + GP_AXIS_INDEX_BASE
  case index == 0x03, index == 0x04:
    // index + to alignment + base
    return index - 1 + GP_AXIS_INDEX_BASE
  case index >= -4 && index <= -1:
    // index + to positive + to alignment + base
    return index + 4 + 4 + GP_AXIS_INDEX_BASE
  case index >= -10 && index <= -7:
    // index + to positive + to alignment + base
    return index + 10 + 8 + GP_AXIS_INDEX_BASE
  default:
    log.Fatal("invalid input index can't convert to mapped index")
  }
  return -1
}

//copied from evdev package and removed everything that is not a KEY/BTN/REL 
//and shifted everything to be in 1 single range so no overlap
const (
	KEY_RESERVED                 = 0
	KEY_ESC                      = 1
	KEY_1                        = 2
	KEY_2                        = 3
	KEY_3                        = 4
	KEY_4                        = 5
	KEY_5                        = 6
	KEY_6                        = 7
	KEY_7                        = 8
	KEY_8                        = 9
	KEY_9                        = 10
	KEY_0                        = 11
	KEY_MINUS                    = 12
	KEY_EQUAL                    = 13
	KEY_BACKSPACE                = 14
	KEY_TAB                      = 15
	KEY_Q                        = 16
	KEY_W                        = 17
	KEY_E                        = 18
	KEY_R                        = 19
	KEY_T                        = 20
	KEY_Y                        = 21
	KEY_U                        = 22
	KEY_I                        = 23
	KEY_O                        = 24
	KEY_P                        = 25
	KEY_LEFTBRACE                = 26
	KEY_RIGHTBRACE               = 27
	KEY_ENTER                    = 28
	KEY_LEFTCTRL                 = 29
	KEY_A                        = 30
	KEY_S                        = 31
	KEY_D                        = 32
	KEY_F                        = 33
	KEY_G                        = 34
	KEY_H                        = 35
	KEY_J                        = 36
	KEY_K                        = 37
	KEY_L                        = 38
	KEY_SEMICOLON                = 39
	KEY_APOSTROPHE               = 40
	KEY_GRAVE                    = 41
	KEY_LEFTSHIFT                = 42
	KEY_BACKSLASH                = 43
	KEY_Z                        = 44
	KEY_X                        = 45
	KEY_C                        = 46
	KEY_V                        = 47
	KEY_B                        = 48
	KEY_N                        = 49
	KEY_M                        = 50
	KEY_COMMA                    = 51
	KEY_DOT                      = 52
	KEY_SLASH                    = 53
	KEY_RIGHTSHIFT               = 54
	KEY_KPASTERISK               = 55
	KEY_LEFTALT                  = 56
	KEY_SPACE                    = 57
	KEY_CAPSLOCK                 = 58
	KEY_F1                       = 59
	KEY_F2                       = 60
	KEY_F3                       = 61
	KEY_F4                       = 62
	KEY_F5                       = 63
	KEY_F6                       = 64
	KEY_F7                       = 65
	KEY_F8                       = 66
	KEY_F9                       = 67
	KEY_F10                      = 68
	KEY_NUMLOCK                  = 69
	KEY_SCROLLLOCK               = 70
	KEY_KP7                      = 71
	KEY_KP8                      = 72
	KEY_KP9                      = 73
	KEY_KPMINUS                  = 74
	KEY_KP4                      = 75
	KEY_KP5                      = 76
	KEY_KP6                      = 77
	KEY_KPPLUS                   = 78
	KEY_KP1                      = 79
	KEY_KP2                      = 80
	KEY_KP3                      = 81
	KEY_KP0                      = 82
	KEY_KPDOT                    = 83
	KEY_ZENKAKUHANKAKU           = 85
	KEY_102ND                    = 86
	KEY_F11                      = 87
	KEY_F12                      = 88
	KEY_RO                       = 89
	KEY_KATAKANA                 = 90
	KEY_HIRAGANA                 = 91
	KEY_HENKAN                   = 92
	KEY_KATAKANAHIRAGANA         = 93
	KEY_MUHENKAN                 = 94
	KEY_KPJPCOMMA                = 95
	KEY_KPENTER                  = 96
	KEY_RIGHTCTRL                = 97
	KEY_KPSLASH                  = 98
	KEY_SYSRQ                    = 99
	KEY_RIGHTALT                 = 100
	KEY_LINEFEED                 = 101
	KEY_HOME                     = 102
	KEY_UP                       = 103
	KEY_PAGEUP                   = 104
	KEY_LEFT                     = 105
	KEY_RIGHT                    = 106
	KEY_END                      = 107
	KEY_DOWN                     = 108
	KEY_PAGEDOWN                 = 109
	KEY_INSERT                   = 110
	KEY_DELETE                   = 111
	KEY_MACRO                    = 112
	KEY_MUTE                     = 113
	KEY_VOLUMEDOWN               = 114
	KEY_VOLUMEUP                 = 115
	KEY_POWER                    = 116
	KEY_KPEQUAL                  = 117
	KEY_KPPLUSMINUS              = 118
	KEY_PAUSE                    = 119
	KEY_SCALE                    = 120
	KEY_KPCOMMA                  = 121
	KEY_HANGEUL                  = 122
	KEY_HANGUEL                  = KEY_HANGEUL
	KEY_HANJA                    = 123
	KEY_YEN                      = 124
	KEY_LEFTMETA                 = 125
	KEY_RIGHTMETA                = 126
	KEY_COMPOSE                  = 127
	KEY_STOP                     = 128
	KEY_AGAIN                    = 129
	KEY_PROPS                    = 130
	KEY_UNDO                     = 131
	KEY_FRONT                    = 132
	KEY_COPY                     = 133
	KEY_OPEN                     = 134
	KEY_PASTE                    = 135
	KEY_FIND                     = 136
	KEY_CUT                      = 137
	KEY_HELP                     = 138
	KEY_MENU                     = 139
	KEY_CALC                     = 140
	KEY_SETUP                    = 141
	KEY_SLEEP                    = 142
	KEY_WAKEUP                   = 143
	KEY_FILE                     = 144
	KEY_SENDFILE                 = 145
	KEY_DELETEFILE               = 146
	KEY_XFER                     = 147
	KEY_PROG1                    = 148
	KEY_PROG2                    = 149
	KEY_WWW                      = 150
	KEY_MSDOS                    = 151
	KEY_COFFEE                   = 152
	KEY_SCREENLOCK               = KEY_COFFEE
	KEY_ROTATE_DISPLAY           = 153
	KEY_DIRECTION                = KEY_ROTATE_DISPLAY
	KEY_CYCLEWINDOWS             = 154
	KEY_MAIL                     = 155
	KEY_BOOKMARKS                = 156
	KEY_COMPUTER                 = 157
	KEY_BACK                     = 158
	KEY_FORWARD                  = 159
	KEY_CLOSECD                  = 160
	KEY_EJECTCD                  = 161
	KEY_EJECTCLOSECD             = 162
	KEY_NEXTSONG                 = 163
	KEY_PLAYPAUSE                = 164
	KEY_PREVIOUSSONG             = 165
	KEY_STOPCD                   = 166
	KEY_RECORD                   = 167
	KEY_REWIND                   = 168
	KEY_PHONE                    = 169
	KEY_ISO                      = 170
	KEY_CONFIG                   = 171
	KEY_HOMEPAGE                 = 172
	KEY_REFRESH                  = 173
	KEY_EXIT                     = 174
	KEY_MOVE                     = 175
	KEY_EDIT                     = 176
	KEY_SCROLLUP                 = 177
	KEY_SCROLLDOWN               = 178
	KEY_KPLEFTPAREN              = 179
	KEY_KPRIGHTPAREN             = 180
	KEY_NEW                      = 181
	KEY_REDO                     = 182
	KEY_F13                      = 183
	KEY_F14                      = 184
	KEY_F15                      = 185
	KEY_F16                      = 186
	KEY_F17                      = 187
	KEY_F18                      = 188
	KEY_F19                      = 189
	KEY_F20                      = 190
	KEY_F21                      = 191
	KEY_F22                      = 192
	KEY_F23                      = 193
	KEY_F24                      = 194
	KEY_PLAYCD                   = 200
	KEY_PAUSECD                  = 201
	KEY_PROG3                    = 202
	KEY_PROG4                    = 203
	KEY_DASHBOARD                = 204
	KEY_SUSPEND                  = 205
	KEY_CLOSE                    = 206
	KEY_PLAY                     = 207
	KEY_FASTFORWARD              = 208
	KEY_BASSBOOST                = 209
	KEY_PRINT                    = 210
	KEY_HP                       = 211
	KEY_CAMERA                   = 212
	KEY_SOUND                    = 213
	KEY_QUESTION                 = 214
	KEY_EMAIL                    = 215
	KEY_CHAT                     = 216
	KEY_SEARCH                   = 217
	KEY_CONNECT                  = 218
	KEY_FINANCE                  = 219
	KEY_SPORT                    = 220
	KEY_SHOP                     = 221
	KEY_ALTERASE                 = 222
	KEY_CANCEL                   = 223
	KEY_BRIGHTNESSDOWN           = 224
	KEY_BRIGHTNESSUP             = 225
	KEY_MEDIA                    = 226
	KEY_SWITCHVIDEOMODE          = 227
	KEY_KBDILLUMTOGGLE           = 228
	KEY_KBDILLUMDOWN             = 229
	KEY_KBDILLUMUP               = 230
	KEY_SEND                     = 231
	KEY_REPLY                    = 232
	KEY_FORWARDMAIL              = 233
	KEY_SAVE                     = 234
	KEY_DOCUMENTS                = 235
	KEY_BATTERY                  = 236
	KEY_BLUETOOTH                = 237
	KEY_WLAN                     = 238
	KEY_UWB                      = 239
	KEY_UNKNOWN                  = 240
	KEY_VIDEO_NEXT               = 241
	KEY_VIDEO_PREV               = 242
	KEY_BRIGHTNESS_CYCLE         = 243
	KEY_BRIGHTNESS_AUTO          = 244
	KEY_BRIGHTNESS_ZERO          = KEY_BRIGHTNESS_AUTO
	KEY_DISPLAY_OFF              = 245
	KEY_WWAN                     = 246
	KEY_WIMAX                    = KEY_WWAN
	KEY_RFKILL                   = 247
	KEY_MICMUTE                  = 248
	BTN_MISC                     = 0x100
	BTN_0                        = 0x100
	BTN_1                        = 0x101
	BTN_2                        = 0x102
	BTN_3                        = 0x103
	BTN_4                        = 0x104
	BTN_5                        = 0x105
	BTN_6                        = 0x106
	BTN_7                        = 0x107
	BTN_8                        = 0x108
	BTN_9                        = 0x109
	BTN_MOUSE                    = 0x110
	BTN_LEFT                     = 0x110
	BTN_RIGHT                    = 0x111
	BTN_MIDDLE                   = 0x112
	BTN_SIDE                     = 0x113
	BTN_EXTRA                    = 0x114
	BTN_FORWARD                  = 0x115
	BTN_BACK                     = 0x116
	BTN_TASK                     = 0x117
	BTN_JOYSTICK                 = 0x120
	BTN_TRIGGER                  = 0x120
	BTN_THUMB                    = 0x121
	BTN_THUMB2                   = 0x122
	BTN_TOP                      = 0x123
	BTN_TOP2                     = 0x124
	BTN_PINKIE                   = 0x125
	BTN_BASE                     = 0x126
	BTN_BASE2                    = 0x127
	BTN_BASE3                    = 0x128
	BTN_BASE4                    = 0x129
	BTN_BASE5                    = 0x12a
	BTN_BASE6                    = 0x12b
	BTN_DEAD                     = 0x12f
	BTN_GAMEPAD                  = 0x130
	BTN_SOUTH                    = 0x130
	BTN_A                        = BTN_SOUTH
	BTN_EAST                     = 0x131
	BTN_B                        = BTN_EAST
	BTN_C                        = 0x132
	BTN_NORTH                    = 0x133
	BTN_X                        = BTN_NORTH
	BTN_WEST                     = 0x134
	BTN_Y                        = BTN_WEST
	BTN_Z                        = 0x135
	BTN_TL                       = 0x136
	BTN_TR                       = 0x137
	BTN_TL2                      = 0x138
	BTN_TR2                      = 0x139
	BTN_SELECT                   = 0x13a
	BTN_START                    = 0x13b
	BTN_MODE                     = 0x13c
	BTN_THUMBL                   = 0x13d
	BTN_THUMBR                   = 0x13e
	BTN_DIGI                     = 0x140
	BTN_TOOL_PEN                 = 0x140
	BTN_TOOL_RUBBER              = 0x141
	BTN_TOOL_BRUSH               = 0x142
	BTN_TOOL_PENCIL              = 0x143
	BTN_TOOL_AIRBRUSH            = 0x144
	BTN_TOOL_FINGER              = 0x145
	BTN_TOOL_MOUSE               = 0x146
	BTN_TOOL_LENS                = 0x147
	BTN_TOOL_QUINTTAP            = 0x148
	BTN_TOUCH                    = 0x14a
	BTN_STYLUS                   = 0x14b
	BTN_STYLUS2                  = 0x14c
	BTN_TOOL_DOUBLETAP           = 0x14d
	BTN_TOOL_TRIPLETAP           = 0x14e
	BTN_TOOL_QUADTAP             = 0x14f
	BTN_WHEEL                    = 0x150
	BTN_GEAR_DOWN                = 0x150
	BTN_GEAR_UP                  = 0x151
	KEY_OK                       = 0x160
	KEY_SELECT                   = 0x161
	KEY_GOTO                     = 0x162
	KEY_CLEAR                    = 0x163
	KEY_POWER2                   = 0x164
	KEY_OPTION                   = 0x165
	KEY_INFO                     = 0x166
	KEY_TIME                     = 0x167
	KEY_VENDOR                   = 0x168
	KEY_ARCHIVE                  = 0x169
	KEY_PROGRAM                  = 0x16a
	KEY_CHANNEL                  = 0x16b
	KEY_FAVORITES                = 0x16c
	KEY_EPG                      = 0x16d
	KEY_PVR                      = 0x16e
	KEY_MHP                      = 0x16f
	KEY_LANGUAGE                 = 0x170
	KEY_TITLE                    = 0x171
	KEY_SUBTITLE                 = 0x172
	KEY_ANGLE                    = 0x173
	KEY_ZOOM                     = 0x174
	KEY_MODE                     = 0x175
	KEY_KEYBOARD                 = 0x176
	KEY_SCREEN                   = 0x177
	KEY_PC                       = 0x178
	KEY_TV                       = 0x179
	KEY_TV2                      = 0x17a
	KEY_VCR                      = 0x17b
	KEY_VCR2                     = 0x17c
	KEY_SAT                      = 0x17d
	KEY_SAT2                     = 0x17e
	KEY_CD                       = 0x17f
	KEY_TAPE                     = 0x180
	KEY_RADIO                    = 0x181
	KEY_TUNER                    = 0x182
	KEY_PLAYER                   = 0x183
	KEY_TEXT                     = 0x184
	KEY_DVD                      = 0x185
	KEY_AUX                      = 0x186
	KEY_MP3                      = 0x187
	KEY_AUDIO                    = 0x188
	KEY_VIDEO                    = 0x189
	KEY_DIRECTORY                = 0x18a
	KEY_LIST                     = 0x18b
	KEY_MEMO                     = 0x18c
	KEY_CALENDAR                 = 0x18d
	KEY_RED                      = 0x18e
	KEY_GREEN                    = 0x18f
	KEY_YELLOW                   = 0x190
	KEY_BLUE                     = 0x191
	KEY_CHANNELUP                = 0x192
	KEY_CHANNELDOWN              = 0x193
	KEY_FIRST                    = 0x194
	KEY_LAST                     = 0x195
	KEY_AB                       = 0x196
	KEY_NEXT                     = 0x197
	KEY_RESTART                  = 0x198
	KEY_SLOW                     = 0x199
	KEY_SHUFFLE                  = 0x19a
	KEY_BREAK                    = 0x19b
	KEY_PREVIOUS                 = 0x19c
	KEY_DIGITS                   = 0x19d
	KEY_TEEN                     = 0x19e
	KEY_TWEN                     = 0x19f
	KEY_VIDEOPHONE               = 0x1a0
	KEY_GAMES                    = 0x1a1
	KEY_ZOOMIN                   = 0x1a2
	KEY_ZOOMOUT                  = 0x1a3
	KEY_ZOOMRESET                = 0x1a4
	KEY_WORDPROCESSOR            = 0x1a5
	KEY_EDITOR                   = 0x1a6
	KEY_SPREADSHEET              = 0x1a7
	KEY_GRAPHICSEDITOR           = 0x1a8
	KEY_PRESENTATION             = 0x1a9
	KEY_DATABASE                 = 0x1aa
	KEY_NEWS                     = 0x1ab
	KEY_VOICEMAIL                = 0x1ac
	KEY_ADDRESSBOOK              = 0x1ad
	KEY_MESSENGER                = 0x1ae
	KEY_DISPLAYTOGGLE            = 0x1af
	KEY_BRIGHTNESS_TOGGLE        = KEY_DISPLAYTOGGLE
	KEY_SPELLCHECK               = 0x1b0
	KEY_LOGOFF                   = 0x1b1
	KEY_DOLLAR                   = 0x1b2
	KEY_EURO                     = 0x1b3
	KEY_FRAMEBACK                = 0x1b4
	KEY_FRAMEFORWARD             = 0x1b5
	KEY_CONTEXT_MENU             = 0x1b6
	KEY_MEDIA_REPEAT             = 0x1b7
	KEY_10CHANNELSUP             = 0x1b8
	KEY_10CHANNELSDOWN           = 0x1b9
	KEY_IMAGES                   = 0x1ba
	KEY_DEL_EOL                  = 0x1c0
	KEY_DEL_EOS                  = 0x1c1
	KEY_INS_LINE                 = 0x1c2
	KEY_DEL_LINE                 = 0x1c3
	KEY_FN                       = 0x1d0
	KEY_FN_ESC                   = 0x1d1
	KEY_FN_F1                    = 0x1d2
	KEY_FN_F2                    = 0x1d3
	KEY_FN_F3                    = 0x1d4
	KEY_FN_F4                    = 0x1d5
	KEY_FN_F5                    = 0x1d6
	KEY_FN_F6                    = 0x1d7
	KEY_FN_F7                    = 0x1d8
	KEY_FN_F8                    = 0x1d9
	KEY_FN_F9                    = 0x1da
	KEY_FN_F10                   = 0x1db
	KEY_FN_F11                   = 0x1dc
	KEY_FN_F12                   = 0x1dd
	KEY_FN_1                     = 0x1de
	KEY_FN_2                     = 0x1df
	KEY_FN_D                     = 0x1e0
	KEY_FN_E                     = 0x1e1
	KEY_FN_F                     = 0x1e2
	KEY_FN_S                     = 0x1e3
	KEY_FN_B                     = 0x1e4
	KEY_BRL_DOT1                 = 0x1f1
	KEY_BRL_DOT2                 = 0x1f2
	KEY_BRL_DOT3                 = 0x1f3
	KEY_BRL_DOT4                 = 0x1f4
	KEY_BRL_DOT5                 = 0x1f5
	KEY_BRL_DOT6                 = 0x1f6
	KEY_BRL_DOT7                 = 0x1f7
	KEY_BRL_DOT8                 = 0x1f8
	KEY_BRL_DOT9                 = 0x1f9
	KEY_BRL_DOT10                = 0x1fa
	KEY_NUMERIC_0                = 0x200
	KEY_NUMERIC_1                = 0x201
	KEY_NUMERIC_2                = 0x202
	KEY_NUMERIC_3                = 0x203
	KEY_NUMERIC_4                = 0x204
	KEY_NUMERIC_5                = 0x205
	KEY_NUMERIC_6                = 0x206
	KEY_NUMERIC_7                = 0x207
	KEY_NUMERIC_8                = 0x208
	KEY_NUMERIC_9                = 0x209
	KEY_NUMERIC_STAR             = 0x20a
	KEY_NUMERIC_POUND            = 0x20b
	KEY_NUMERIC_A                = 0x20c
	KEY_NUMERIC_B                = 0x20d
	KEY_NUMERIC_C                = 0x20e
	KEY_NUMERIC_D                = 0x20f
	KEY_CAMERA_FOCUS             = 0x210
	KEY_WPS_BUTTON               = 0x211
	KEY_TOUCHPAD_TOGGLE          = 0x212
	KEY_TOUCHPAD_ON              = 0x213
	KEY_TOUCHPAD_OFF             = 0x214
	KEY_CAMERA_ZOOMIN            = 0x215
	KEY_CAMERA_ZOOMOUT           = 0x216
	KEY_CAMERA_UP                = 0x217
	KEY_CAMERA_DOWN              = 0x218
	KEY_CAMERA_LEFT              = 0x219
	KEY_CAMERA_RIGHT             = 0x21a
	KEY_ATTENDANT_ON             = 0x21b
	KEY_ATTENDANT_OFF            = 0x21c
	KEY_ATTENDANT_TOGGLE         = 0x21d
	KEY_LIGHTS_TOGGLE            = 0x21e
	BTN_DPAD_UP                  = 0x220
	BTN_DPAD_DOWN                = 0x221
	BTN_DPAD_LEFT                = 0x222
	BTN_DPAD_RIGHT               = 0x223
	KEY_ALS_TOGGLE               = 0x230
	KEY_BUTTONCONFIG             = 0x240
	KEY_TASKMANAGER              = 0x241
	KEY_JOURNAL                  = 0x242
	KEY_CONTROLPANEL             = 0x243
	KEY_APPSELECT                = 0x244
	KEY_SCREENSAVER              = 0x245
	KEY_VOICECOMMAND             = 0x246
	KEY_BRIGHTNESS_MIN           = 0x250
	KEY_BRIGHTNESS_MAX           = 0x251
	KEY_KBDINPUTASSIST_PREV      = 0x260
	KEY_KBDINPUTASSIST_NEXT      = 0x261
	KEY_KBDINPUTASSIST_PREVGROUP = 0x262
	KEY_KBDINPUTASSIST_NEXTGROUP = 0x263
	KEY_KBDINPUTASSIST_ACCEPT    = 0x264
	KEY_KBDINPUTASSIST_CANCEL    = 0x265
	KEY_RIGHT_UP                 = 0x266
	KEY_RIGHT_DOWN               = 0x267
	KEY_LEFT_UP                  = 0x268
	KEY_LEFT_DOWN                = 0x269
	KEY_ROOT_MENU                = 0x26a
	KEY_MEDIA_TOP_MENU           = 0x26b
	KEY_NUMERIC_11               = 0x26c
	KEY_NUMERIC_12               = 0x26d
	KEY_AUDIO_DESC               = 0x26e
	KEY_3D_MODE                  = 0x26f
	KEY_NEXT_FAVORITE            = 0x270
	KEY_STOP_RECORD              = 0x271
	KEY_PAUSE_RECORD             = 0x272
	KEY_VOD                      = 0x273
	KEY_UNMUTE                   = 0x274
	KEY_FASTREVERSE              = 0x275
	KEY_SLOWREVERSE              = 0x276
	KEY_DATA                     = 0x275
	BTN_TRIGGER_HAPPY            = 0x2c0
	BTN_TRIGGER_HAPPY1           = 0x2c0
	BTN_TRIGGER_HAPPY2           = 0x2c1
	BTN_TRIGGER_HAPPY3           = 0x2c2
	BTN_TRIGGER_HAPPY4           = 0x2c3
	BTN_TRIGGER_HAPPY5           = 0x2c4
	BTN_TRIGGER_HAPPY6           = 0x2c5
	BTN_TRIGGER_HAPPY7           = 0x2c6
	BTN_TRIGGER_HAPPY8           = 0x2c7
	BTN_TRIGGER_HAPPY9           = 0x2c8
	BTN_TRIGGER_HAPPY10          = 0x2c9
	BTN_TRIGGER_HAPPY11          = 0x2ca
	BTN_TRIGGER_HAPPY12          = 0x2cb
	BTN_TRIGGER_HAPPY13          = 0x2cc
	BTN_TRIGGER_HAPPY14          = 0x2cd
	BTN_TRIGGER_HAPPY15          = 0x2ce
	BTN_TRIGGER_HAPPY16          = 0x2cf
	BTN_TRIGGER_HAPPY17          = 0x2d0
	BTN_TRIGGER_HAPPY18          = 0x2d1
	BTN_TRIGGER_HAPPY19          = 0x2d2
	BTN_TRIGGER_HAPPY20          = 0x2d3
	BTN_TRIGGER_HAPPY21          = 0x2d4
	BTN_TRIGGER_HAPPY22          = 0x2d5
	BTN_TRIGGER_HAPPY23          = 0x2d6
	BTN_TRIGGER_HAPPY24          = 0x2d7
	BTN_TRIGGER_HAPPY25          = 0x2d8
	BTN_TRIGGER_HAPPY26          = 0x2d9
	BTN_TRIGGER_HAPPY27          = 0x2da
	BTN_TRIGGER_HAPPY28          = 0x2db
	BTN_TRIGGER_HAPPY29          = 0x2dc
	BTN_TRIGGER_HAPPY30          = 0x2dd
	BTN_TRIGGER_HAPPY31          = 0x2de
	BTN_TRIGGER_HAPPY32          = 0x2df
	BTN_TRIGGER_HAPPY33          = 0x2e0
	BTN_TRIGGER_HAPPY34          = 0x2e1
	BTN_TRIGGER_HAPPY35          = 0x2e2
	BTN_TRIGGER_HAPPY36          = 0x2e3
	BTN_TRIGGER_HAPPY37          = 0x2e4
	BTN_TRIGGER_HAPPY38          = 0x2e5
	BTN_TRIGGER_HAPPY39          = 0x2e6
	BTN_TRIGGER_HAPPY40          = 0x2e7
	KEY_MIN_INTERESTING          = KEY_MUTE
	KEY_MAX                      = 0x2ff
	REL_X                        = 0x00 + KEY_MAX + 1
	REL_Y                        = 0x01 + KEY_MAX + 1
	REL_Z                        = 0x02 + KEY_MAX + 1
	REL_RX                       = 0x03 + KEY_MAX + 1
	REL_RY                       = 0x04 + KEY_MAX + 1
	REL_RZ                       = 0x05 + KEY_MAX + 1
	REL_HWHEEL                   = 0x06 + KEY_MAX + 1
	REL_DIAL                     = 0x07 + KEY_MAX + 1
	REL_WHEEL                    = 0x08 + KEY_MAX + 1
	REL_MISC                     = 0x09 + KEY_MAX + 1
	REL_MAX                      = 0x0f + KEY_MAX + 1
  ECODES_MAX                   = REL_MAX + 1                    
)

