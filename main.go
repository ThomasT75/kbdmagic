package main

import (
	"cmp"
	"log"

	"math"
	"sync"
	"time"

	"strings"

	"github.com/bendahl/uinput"

	"github.com/grafov/evdev"

	"kbdmagic/ecodes"
)

//TODO
//bindable sense ajudment
//solf press on keys like a ramp based on time the more you hold the faster it goes
//keys that affect trigger/stick forces
//if the recenter vector and the diferent rel vector is similiar we could instant skip to the edge of the deadzone  plus the rel diference
//diference vec if this vec is returning to center with some speed do skip to the edge of the deadzone or add the diference an go past the deadzone
//some vector maths need better 0 checking

// good mice to analog recepie
// edge ring
//    basically clamp how far away the mouse can go from center
// recenter
//    instead of having to bring the mouse to the deadzone do it automatic
// recenter slower the closer you are to the deadzone
//    so that finer movements are translated well
// don't recenter if inside the deadzone
//    if you do that you might throw a manual recenter into a false-positive input
// VECTOR MATH
//    if you recenter each axis alone you will lose diagonal movement

// example of flag usage
// MatchM := flag.String("mouse", "", "force the use of a specify mouse ex: -mouse \"Mouse Name\" ")
// flag.String("keyboard", "", "force the use of a specify keyboard ex: -keyboard \"Keyboard Name\"")
// flag.Parse()
// args := flag.Args()
// if len(args) == 1 {
//
// }

var RemapTable = map[uint16]int {
  evdev.KEY_0: uinput.ButtonMode,
  evdev.KEY_SPACE: uinput.ButtonSouth,
  evdev.KEY_E: uinput.ButtonNorth,
  evdev.KEY_UP: uinput.ButtonWest,
  evdev.KEY_DOWN: uinput.ButtonSouth,
  evdev.KEY_LEFT: uinput.ButtonNorth,
  evdev.KEY_RIGHT: uinput.ButtonEast,
  evdev.KEY_TAB: uinput.ButtonStart,
  evdev.KEY_GRAVE: uinput.ButtonSelect,
  evdev.BTN_LEFT: uinput.ButtonTriggerRight,
  evdev.BTN_RIGHT: uinput.ButtonTriggerLeft,
  evdev.BTN_EXTRA: uinput.ButtonWest,
  evdev.BTN_SIDE: uinput.ButtonEast,
  evdev.BTN_MIDDLE: uinput.ButtonThumbRight,
  evdev.KEY_F: uinput.ButtonEast,
  evdev.KEY_X: uinput.ButtonDpadDown,
  evdev.KEY_K: uinput.ButtonDpadDown,
  evdev.KEY_R: uinput.ButtonDpadUp,
  evdev.KEY_I: uinput.ButtonDpadUp,
  evdev.KEY_Z: uinput.ButtonDpadLeft,
  evdev.KEY_C: uinput.ButtonDpadRight,
  evdev.KEY_LEFTSHIFT: uinput.ButtonBumperRight,
  evdev.KEY_LEFTCTRL: uinput.ButtonBumperLeft,
}


//diferent remap tables because rel and key share the same int range aka REL_WHEEL == KEY_0
//meaning if you roll your mouse wheel it would trigger KEY_0 and vice versa
//key and btn don't share the same int range so we can assing BTN_* to this map
var KeyRemapTable2 [evdev.KEY_MAX]outputCmd
var RelRemapTable2 [evdev.REL_MAX]outputCmd

var RemapTable2 [ecodes.ECODES_MAX]outputCmd

func init() {
  i := ecodes.FromEvdev(evdev.KEY_F, evdev.EV_KEY)
  RemapTable2[i] = outputCmd{
    Type: BUTTON_CMD,
    Code: ecodes.GP_BTN_MAP["GP_BTN_A"],
  }
  i = ecodes.FromEvdev(evdev.REL_X, evdev.EV_REL)
  RemapTable2[i] = outputCmd{
    Type: ABS_CMD,
    Code: ecodes.GP_AXIS_MAP["GP_AXIS_RX"],
  }
  i = ecodes.FromEvdev(evdev.REL_Y, evdev.EV_REL)
  RemapTable2[i] = outputCmd{
    Type: ABS_CMD,
    Code: ecodes.GP_AXIS_MAP["GP_AXIS_RY"],
  }
  i = ecodes.FromEvdev(evdev.KEY_LEFTCTRL, evdev.EV_KEY)
  RemapTable2[i] = outputCmd{
    Type: BOOL_CMD,
    Code: 0,
  }
}

const (
  BUTTON_CMD = iota
  ABS_CMD
  MACRO_CMD
  BOOL_CMD
)

//the values here should be evdev ready meaning no ecodes.ToEvdev
//also that function doesn't exist
type outputCmd struct {
  Type int //event type
  Code int //event code
  Op int
  delay time.Duration // delay >= Millisecond this is a delay cmd
}

var MacroTable [][]outputCmd

var BoolRemapTable [ecodes.ECODES_MAX*2][]outputCmd

var PressedKeys [evdev.KEY_MAX]int
  
var MatchK = "Evision RGB Keyboard"
var MatchM = "USB Gaming Mouse"
  
var kbd_grabed = false
var mice_grabed = false
var stop = false


func main() {
  devices, err := evdev.ListInputDevices()
  if err != nil {
    log.Fatal(err.Error())
  }
  

  var kbd *evdev.InputDevice
  var mice *evdev.InputDevice
  // var ports []string
  for _, dev := range devices {
    //input0 seems to be the real device
    if !strings.Contains(dev.Phys, "input0") {
      continue
    }
    if strings.Contains(dev.Name, MatchK) && kbd == nil {
      log.Println("kbd found")
      kbd = dev 
      err = kbd.Grab()
      if err != nil {
        log.Fatal(err.Error())
      }
      defer kbd.Release()
      kbd_grabed = true
      log.Println("kbd grabed")
    }
    if strings.Contains(dev.Name, MatchM) && mice == nil {
      log.Println("mouse found")
      mice = dev 
      err = mice.Grab()
      if err != nil {
        log.Fatal(err.Error())
      }
      defer mice.Release()
      mice_grabed = true
      log.Println("mouse grabed")
    }
  }


  // controller, err := uinput.CreateGamepad("/dev/uinput", []byte("Sony Computer Entertainment Wireless Controller"), 0x1356, 0x1476)
  controller, err := uinput.CreateGamepad("/dev/uinput", []byte("gamepadkbdmagic"), 0xDEAD, 0xBEEF)
  if err != nil {
    log.Fatal(err.Error())
  }
  defer controller.Close()
  controller.LeftTriggerForce(-1)
  controller.RightTriggerForce(-1)

  lstick := Stick{}
  rstick := Stick{}

  cmutex := sync.Mutex{}

  tgrab := make(chan int)

  go RunInput(kbd, tgrab, controller, &cmutex, &lstick, &rstick)
  go RunInput(mice, tgrab, controller, &cmutex, &lstick, &rstick)
  go UpdateController(controller, &cmutex, &lstick, &rstick)


  for {
    var d int
    d = <-tgrab // this will block
    if d == 2 {
      break
    }

    var err error 

    if kbd_grabed {
      err = kbd.Release()
    } else {
      err = kbd.Grab()
    }
    if err != nil {
      log.Fatal(err.Error())
    }
    kbd_grabed = !kbd_grabed

    if mice_grabed {
      err = mice.Release()
    } else {
      err = mice.Grab()
    }
    if err != nil {
      log.Fatal(err.Error())
    }
    mice_grabed = !mice_grabed
  }
}

func clamp[T cmp.Ordered](value T, minium T, maxium T) T {
  return max(minium, min(value, maxium))
}

func VecMag(x float32, y float32) float32 {
  s := float64(x * x + y * y)
  if s < 1.0 {
    return 1
  }
  return float32(math.Sqrt(float64(x * x + y * y)))
}

const StickPoints float32 = 320.0
const ReCenterSpeed float32 = 0.1 // 0-1 
const RecenterPoints float32 = StickPoints * ReCenterSpeed
const DeadZone float32 = 0.14 
const DeadZonePoints float32 = StickPoints * DeadZone

type Stick struct {
  x float32 //real value sent to controller
  y float32
  relx float32 //used as semi-direct axis translation
  rely float32
  bx float32 //forces max/min value
  by float32 
  m sync.Mutex
}

//might need to add triggers to this 
type GamepadState struct {
  lstick *Stick
  rstick *Stick
  pressed *map[int]int
}


func UpdateController(controller uinput.Gamepad, cmutex *sync.Mutex, lstick *Stick, rstick *Stick) {
  sticks := []*Stick{lstick, rstick}
  for {
    //sticks
    for _, stick := range sticks {
      stick.m.Lock()
      // Edge Ring
      stick.relx = clamp(stick.relx, -StickPoints, StickPoints) 
      stick.rely = clamp(stick.rely, -StickPoints, StickPoints)

      // Translate
      stick.x = stick.relx / StickPoints
      stick.y = stick.rely / StickPoints

      // Recenter
      relmag := VecMag(stick.relx, stick.rely) 
      if relmag > DeadZonePoints {
        stick.relx -= RecenterPoints * (float32(math.Abs(float64(stick.relx))) / StickPoints) * (stick.relx / relmag)
        stick.rely -= RecenterPoints * (float32(math.Abs(float64(stick.rely))) / StickPoints) * (stick.rely / relmag)
      }
      
      // Tanslate Boolean Keys to Analog
      if stick.bx != 0 {
        stick.x = stick.bx
      }
      if stick.by != 0 {
        stick.y = stick.by
      }

      // Making Stick Axis An Unit Vector
      // Might Be Unnecesery
      if stick.x != 0.0 && stick.y != 0.0 {
        mg := VecMag(stick.x, stick.y) 
        stick.x = (stick.x / mg)
        stick.y = (stick.y / mg)
      }

      stick.m.Unlock()
    }

    cmutex.Lock()
    controller.LeftStickMove(lstick.x, lstick.y)
    controller.RightStickMove(rstick.x, rstick.y)
    cmutex.Unlock()
    
    time.Sleep(8 * time.Millisecond)
  }
}

func DoBoolCmd(c int, evalue int32) {
  i := clamp(int(evalue), 0, 1) + c
  for k, v := range BoolRemapTable[i] {
     

  }
}

func DoButton(btn int, evalue int32, controller uinput.Gamepad, cmutex *sync.Mutex) {
  //assuming evalue is either 0 or 1
  p := PressedKeys[btn] 
  PressedKeys[btn] = max(p + int(evalue*2-1), 0)
  value := min(PressedKeys[btn], 1)

  switch btn {
  case uinput.ButtonTriggerLeft: 
  cmutex.Lock()
  controller.LeftTriggerForce(float32(value*2-1))
  cmutex.Unlock()
  case uinput.ButtonTriggerRight:
  cmutex.Lock()
  controller.RightTriggerForce(float32(value*2-1))
  cmutex.Unlock()
  } 

  if value == 1 {
    cmutex.Lock()
    controller.ButtonDown(btn)
    cmutex.Unlock()
  } else {
    cmutex.Lock()
    controller.ButtonUp(btn)
    cmutex.Unlock()
  }
}

func DoAxis(abs int, value int32, lstick *Stick, rstick *Stick) {
  axis_value := float32(value*2-1)
  daxis_value := float32(value)
  
  switch abs {
  //Left Stick Axis
  case ecodes.ABS_X:
  lstick.m.Lock()
  lstick.relx += daxis_value
  lstick.m.Unlock()
  case ecodes.ABS_Y:
  lstick.m.Lock()
  lstick.rely += daxis_value
  lstick.m.Unlock()

  //Right Stick Axis
  case ecodes.ABS_RX:
  rstick.m.Lock()
  rstick.relx += daxis_value
  rstick.m.Unlock()
  case ecodes.ABS_RY:
  rstick.m.Lock()
  rstick.rely += daxis_value
  rstick.m.Unlock()

  //Left X
  case ecodes.ABS_X_POSITIVE:
  lstick.m.Lock()
  lstick.bx = max(lstick.bx + axis_value, 0)
  lstick.m.Unlock()
  case ecodes.ABS_X_NEGATIVE:
  lstick.m.Lock()
  lstick.bx = min(lstick.bx - axis_value, 0)
  lstick.m.Unlock()

  //Left Y
  case ecodes.ABS_Y_POSITIVE:
  lstick.m.Lock()
  lstick.by = max(lstick.by + axis_value, 0)
  lstick.m.Unlock()
  case ecodes.ABS_Y_NEGATIVE:
  lstick.m.Lock()
  lstick.by = min(lstick.by - axis_value, 0)
  lstick.m.Unlock()

  //Right X
  case ecodes.ABS_RX_POSITIVE:
  rstick.m.Lock()
  rstick.bx = max(rstick.bx + axis_value, 0)
  rstick.m.Unlock()
  case ecodes.ABS_RX_NEGATIVE:
  rstick.m.Lock()
  rstick.bx = min(rstick.bx - axis_value, 0)
  rstick.m.Unlock()

  //Right Y
  case ecodes.ABS_RY_POSITIVE:
  rstick.m.Lock()
  rstick.by = max(rstick.by + axis_value, 0)
  rstick.m.Unlock()
  case ecodes.ABS_RY_NEGATIVE:
  rstick.m.Lock()
  rstick.by = min(rstick.by - axis_value, 0)
  rstick.m.Unlock()

  }
} 

func RunInput(id *evdev.InputDevice, tgrab chan int, controller uinput.Gamepad, cmutex *sync.Mutex, lstick *Stick, rstick *Stick) {
  for {
    event, err := id.ReadOne()
    if err != nil {
      log.Fatal(err.Error())
    }
    // fmt.Println(event.String())
    if event.Code == evdev.KEY_PAUSE && event.Value == 1 { 
      tgrab <- 1
      stop = !stop
    }
    if event.Code == evdev.KEY_DELETE && event.Value == 1 {
      tgrab <- 2
      stop = !stop
    }
    if stop {
      continue
    }
    //remove key repeat without writing to the device file
    if event.Type == 1 && event.Value == 2 { 
      continue
    }
    //some weird event that can cause problems
    if event.Type == 0 {
      continue
    }
    //it can't return a error other than nil so yea
    index := ecodes.FromEvdev(event.Code, event.Type)
    var outCmd outputCmd
    //this is so a event can't trigger another with a similar code value
    //but diferent event type
    //like KEY_0 and REL_WHEEL
    // if event.Type == evdev.EV_KEY {
    //   outCmd = KeyRemapTable2[event.Code]
    // } else if event.Type == evdev.EV_REL {
    //   outCmd = RelRemapTable2[event.Code]
    // }
    outCmd = RemapTable2[index]
    if outCmd.Type == BUTTON_CMD {
      DoButton(outCmd.Code, event.Value, controller, cmutex)
    } else if outCmd.Type == ABS_CMD {
      DoAxis(outCmd.Code, event.Value, lstick, rstick)
    } else if outCmd.Type == BOOL_CMD {
      DoBoolCmd(outCmd.Code, event.Value)
    }
  }
}
