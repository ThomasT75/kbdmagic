package main

import (
	"cmp"
	"flag"
	"log"
	"os"
	"path/filepath"

	"math"
	"sync"
	"time"

	"strings"

	"github.com/ThomasT75/uinput"

	"github.com/grafov/evdev"

	"kbdmagic/common"
	"kbdmagic/ecodes"
	"kbdmagic/remap"
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

type outputCmd = common.OutputCmd

var RemapTable common.RemapTableType

var MacroTable [][]outputCmd

var BoolRemapTable [common.REMAP_TABLE_SIZE][]outputCmd

var PressedKeys [evdev.KEY_MAX]int
  
var kbd_grabed = false
var mice_grabed = false
var stop = false
var show_inputs = false

func main() {
  // flag init
  flags_should_quit := false // if set quit after flag.Parse()
  remap_name := "Default"

  // flags
  MatchM := flag.String("mouse", "Mouse", "force the use of a specify mouse ex: -mouse \"Mouse Name\" ")
  MatchK := flag.String("keyboard", "Keyboard", "force the use of a specify keyboard ex: -keyboard \"Keyboard Name\"")
  flag.BoolVar(&show_inputs, "show_inputs", false, "prints into the cli the events that are going to be processed")
  flag.BoolFunc("ld", "list devices names for use in -mouse & -keyboard", func(string) error {
    devices, err := evdev.ListInputDevices()
    if err != nil {
      return err
    }
    for _, dev := range devices {
      //input0 seems to be the real device
      if !strings.Contains(dev.Phys, "input0") {
        continue
      }
      println(dev.Name)
    }
    flags_should_quit = true
    return nil
  })
  flag.BoolFunc("lr", "list remap names in the ./portable/remap directory", func(string) error {
    remaps_dir := common.GetRemapsDir()
    remap_files_ls, err := os.ReadDir(remaps_dir)
    if err != nil {
      return err
    }
    for _, entry := range remap_files_ls {
      if entry.IsDir() {
        continue
      }
      if filepath.Ext(entry.Name()) == common.REMAP_FILE_EXT {
        n := strings.TrimSuffix(entry.Name(), common.REMAP_FILE_EXT)
        println(n)
      }
    }
    flags_should_quit = true
    return nil
  })

  // flag parsing
  flag.Parse()
  if flags_should_quit {
    return
  }
  args := flag.Args()
  if len(args) == 1 {
    // this is for args after the flags
    remap_name = args[0]
  } else if len(args) > 1 {
    log.Fatal("Too many arguments expected Remap name")
  } else {
    log.Println("Using Default Remap file")
  }

  // execution
  remaps_dir := common.GetRemapsDir()

  remap_filename := remap_name + common.REMAP_FILE_EXT

  remap_sc_b, err := os.ReadFile(filepath.Join(remaps_dir, remap_filename))
  if err != nil {
    log.Fatal(err.Error())
  }
  remap_sc := string(remap_sc_b)

  var err_list *[]error
  RemapTable, err_list = remap.GetRemapTable(remap_sc)
  if err_list != nil {
    for _, err := range *err_list {
      log.Println(err)
    }
    log.Fatal("Errors Were Reported")
  }

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
    if strings.Contains(dev.Name, *MatchK) && kbd == nil {
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
    if strings.Contains(dev.Name, *MatchM) && mice == nil {
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
  if mice == nil {
    log.Fatal("Couldn't find a mouse")
  }
  if kbd == nil {
    log.Fatal("Couldn't find a keyboard")
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

func VecMagZeroSafe(x float32, y float32) float32 {
  s := float64(x * x + y * y)
  if s < 1.0 {
    return 1
  }
  return float32(math.Sqrt(s))
}

func VecMag(x float32, y float32) float32 {
  s := float64(x * x + y * y) 
  return float32(math.Sqrt(s))
}

func VecNormalize(x float32, y float32) (ux float32, uy float32) {
  m := VecMag(x, y)
  ux = x / m
  uy = y / m
  return ux, uy
}

func VecNormalizeZeroSafe(x float32, y float32) (ux float32, uy float32) {
  m := VecMagZeroSafe(x, y)
  ux = x / m
  uy = y / m
  return ux, uy
}

const StickPoints float32 = 320.0
const ReCenterSpeed float32 = 0.075 // 0-1 
const ReCenterPoints float32 = StickPoints * ReCenterSpeed
const DeadZone float32 = 0.20
const DeadZonePoints float32 = StickPoints * DeadZone
const SlingshotZone float32 = 0.195
const SlingshotZonePoints float32 = StickPoints * SlingshotZone 

type Stick struct {
  x float32 //real value sent to controller
  y float32
  relx float32 //used as semi-direct axis translation
  rely float32
  vx float32
  vy float32
  rax [roolsize]float32 // rolling account
  ray [roolsize]float32 
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

func SquareToCircleMap(x float32, y float32) (cx float32, cy float32) {
  cx = x * float32(math.Sqrt(float64(1.0 - 0.5 * y * y)))
  cy = y * float32(math.Sqrt(float64(1.0 - 0.5 * x * x)))
  return cx, cy
}

func MathAbs(x float32) float32 {
  return float32(math.Abs(float64(x)))
}

func VecDot(x1 float32, y1 float32, x2 float32, y2 float32) float32 {
  return (x1 * x2) + (y1 * y2)
}

const roolsize = 16

func RoolAdd(s *[roolsize]float32, add float32) {
  idx := int(s[0])
  if idx > roolsize - 1 {
    idx = 1
  }
  if idx < 0 {
    idx = 1
  }
  s[0] = float32(idx + 1)
  s[idx] = add
}

func RoolRead(s *[roolsize]float32) float32 {
  var sum float32
  for _, v := range(s[1:]) {
    sum += v
  }
  return sum
}

// maybe add a fifo queue to analog movement based on the velocity of the movement
// or rethink how the problem is dealt with 
// > deadzone 
// > read pass that 
// > convert to input then recenter based on the same velocity of the input 
// > then do what we do now with slingshowzone
func UpdateController(controller uinput.Gamepad, cmutex *sync.Mutex, lstick *Stick, rstick *Stick) {
  sticks := []*Stick{lstick, rstick}
  for {
    start := time.Now()
    //sticks
    for _, stick := range sticks {
      stick.m.Lock()

      // Edge Ring
      stick.relx = clamp(stick.relx, -StickPoints, StickPoints) 
      stick.rely = clamp(stick.rely, -StickPoints, StickPoints)

      // cvx := stick.relx - stick.vx
      // cvy := stick.rely - stick.vy
      
      vx := RoolRead(&stick.rax)
      vy := RoolRead(&stick.ray)

      // Translate
      stick.x = stick.relx / StickPoints
      stick.y = stick.rely / StickPoints

      // Recenter
      relmag := VecMagZeroSafe(stick.relx, stick.rely) 
      if relmag > DeadZonePoints {
        stick.relx -= ReCenterPoints * (float32(math.Abs(float64(stick.relx))) / StickPoints) * (stick.relx / (relmag + 1))
        stick.rely -= ReCenterPoints * (float32(math.Abs(float64(stick.rely))) / StickPoints) * (stick.rely / (relmag + 1))
      } else if vx > 0 || vy > 0 {
        nvx, nvy := VecNormalize(vx, vy)
        stick.relx = nvx * SlingshotZonePoints
        stick.rely = nvy * SlingshotZonePoints
      }

      vmag := VecMagZeroSafe(vx, vy)
      if vmag > 1.5 {

        if vx != 0 || vy != 0 {
          nvx, nvy := VecNormalize(vx, vy)
          d := VecDot(nvx, nvy, stick.x, stick.y)
          
          if d < -0.5 {
            stick.relx = nvx * SlingshotZonePoints
            stick.rely = nvy * SlingshotZonePoints
          }

        }

      }
      // if relmag < DeadZonePoints {
      //   ux, uy := VecNormalizeZeroSafe(cvx, cvy)
      //   stick.relx = ux * SlingshotZonePoints
      //   stick.rely = uy * SlingshotZonePoints
      // }
      
      // Tanslate Boolean Keys to Analog
      if stick.bx != 0 {
        stick.x = stick.bx
      }
      if stick.by != 0 {
        stick.y = stick.by
      }

      
      stick.x = clamp(stick.x, -1.0, 1.0)
      stick.y = clamp(stick.y, -1.0, 1.0)

      // stick.x, stick.y = SquareToCircleMap(stick.x, stick.y)
      
      stick.x, stick.y = VecNormalizeZeroSafe(stick.x, stick.y)

      RoolAdd(&stick.rax, 0)
      RoolAdd(&stick.ray, 0)
      stick.m.Unlock()
    }

    cmutex.Lock()
    controller.LeftStickMove(lstick.x, lstick.y)
    controller.RightStickMove(rstick.x, rstick.y)
    cmutex.Unlock()

    delta := time.Now().Sub(start)
    // log.Println(delta)
    time.Sleep(8 * time.Millisecond - delta)
  }
}

// func DoBoolCmd(c int, evalue int32) {
//   i := clamp(int(evalue), 0, 1) + c
//   for k, v := range BoolRemapTable[i] {
//      
//
//   }
// }

func DoButton(cmd outputCmd, evalue int32, controller uinput.Gamepad, cmutex *sync.Mutex) {
  //assuming evalue is either 0 or 1
  if cmd.Op == common.PRESS_OP {
    evalue = 1
  } else if cmd.Op == common.RELEASE_OP {
    evalue = 0
  } else if cmd.Op == common.CLICK_OP {
    go func() { //hack
      cmutex.Lock()
      controller.ButtonDown(cmd.Code)
      cmutex.Unlock()
      time.Sleep(50 * time.Millisecond)
      cmutex.Lock()
      controller.ButtonUp(cmd.Code)
      cmutex.Unlock()
    }()
    return
  }
  btn := cmd.Code
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

func DoAxis(cmd outputCmd, value int32, lstick *Stick, rstick *Stick) {
  axis_value := float32(value*2-1) * cmd.Force
  daxis_value := float32(value) * cmd.Force
  
  switch cmd.Code {
  //Left Stick Axis
  case ecodes.ABS_X:
  lstick.m.Lock()
  lstick.relx += daxis_value
  lstick.vx += daxis_value
  RoolAdd(&lstick.rax, daxis_value)
  lstick.m.Unlock()
  case ecodes.ABS_Y:
  lstick.m.Lock()
  lstick.rely += daxis_value
  lstick.vy += daxis_value
  RoolAdd(&lstick.ray, daxis_value)
  lstick.m.Unlock()

  //Right Stick Axis
  case ecodes.ABS_RX:
  rstick.m.Lock()
  rstick.relx += daxis_value
  rstick.vx += daxis_value
  RoolAdd(&rstick.rax, daxis_value)
  rstick.m.Unlock()
  case ecodes.ABS_RY:
  rstick.m.Lock()
  rstick.rely += daxis_value
  rstick.vy += daxis_value
  RoolAdd(&rstick.ray, daxis_value)
  rstick.m.Unlock()

  //Left X
  case ecodes.ABS_X_POSITIVE:
  lstick.m.Lock()
  lstick.bx = lstick.bx + axis_value
  lstick.m.Unlock()
  case ecodes.ABS_X_NEGATIVE:
  lstick.m.Lock()
  lstick.bx = lstick.bx - axis_value
  lstick.m.Unlock()

  //Left Y
  case ecodes.ABS_Y_POSITIVE:
  lstick.m.Lock()
  lstick.by = lstick.by + axis_value
  lstick.m.Unlock()
  case ecodes.ABS_Y_NEGATIVE:
  lstick.m.Lock()
  lstick.by = lstick.by - axis_value
  lstick.m.Unlock()

  //Right X
  case ecodes.ABS_RX_POSITIVE:
  rstick.m.Lock()
  rstick.bx = rstick.bx + axis_value
  rstick.m.Unlock()
  case ecodes.ABS_RX_NEGATIVE:
  rstick.m.Lock()
  rstick.bx = rstick.bx - axis_value
  rstick.m.Unlock()

  //Right Y
  case ecodes.ABS_RY_POSITIVE:
  rstick.m.Lock()
  rstick.by = rstick.by + axis_value
  rstick.m.Unlock()
  case ecodes.ABS_RY_NEGATIVE:
  rstick.m.Lock()
  rstick.by = rstick.by - axis_value
  rstick.m.Unlock()

  }
} 

func RunInput(id *evdev.InputDevice, tgrab chan int, controller uinput.Gamepad, cmutex *sync.Mutex, lstick *Stick, rstick *Stick) {
  for {
    event, err := id.ReadOne()
    if err != nil {
      log.Fatal(err.Error())
    }
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

    // only events we support
    if event.Type != 1 && event.Type != 2 {
      continue
    }

    if show_inputs {
      log.Println(event.String())
    }

    index := ecodes.FromEvdev(event.Code, event.Type)
    stateIndex := ecodes.ToStateIndex(index, event.Value)

    var outCmd outputCmd
    outCmd = RemapTable[stateIndex]
    switch outCmd.Type {
    case common.BUTTON_CMD:
      DoButton(outCmd, event.Value, controller, cmutex)
    case common.ABS_CMD:
      DoAxis(outCmd, event.Value, lstick, rstick)
    case common.BOOL_CMD:
      // DoBoolCmd(outCmd.Code, event.Value)
    }
  }
}
