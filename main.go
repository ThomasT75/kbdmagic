package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"sync"
	"time"

	"strings"

	"github.com/grafov/evdev"

	"kbdmagic/common"
	"kbdmagic/controller"
	"kbdmagic/ecodes"
	"kbdmagic/internal/buffers"
	"kbdmagic/internal/log"
	"kbdmagic/internal/vec"
	"kbdmagic/remap"
)

// TODO
// better pretty print in general
// interface for a stick so it can change behavior at remap time
// log to file (internal logging module)
// dump remap table to log file hotkey (pretty print)
// rewrite validator to be an actual validator (use notes in validator.go)
// move evdev behind a interface as a test for the future
// more control over the state of a certain key (simpler way of handling toggle)

const (
  VERSION_MAJOR = 0
  VERSION_MINOR = 1
  VERSION_PATCH = 0
)

func GetVersion() string {
  return fmt.Sprintf("v%v.%v.%v", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
}

var RemapTable common.RemapTableType
var RemapTableMutex sync.Mutex

var SequenceQueue []buffers.CircularBuf[common.SequenceCmd]
var SequenceTable common.SequenceTableType

var Options common.Options

//every time BOOL_CMD change they state they need to edit the remaptable using this array
var BoolRemapTable common.BoolRemapTableType

//basically an is_pressed() function but in array form
var StateTable [common.REMAP_TABLE_SIZE]int32 

var StateToggleTable [common.REMAP_TABLE_SIZE]bool

var PressedKeys [evdev.KEY_MAX]int

var ClickButtonQueue [ecodes.GP_INDEX_MAX]buffers.CircularBuf[common.OutputCmd]

var RemapUndo common.RemapTableType
  
var keyboard_grabed = false
var mouse_grabed = false
var stop = true
var show_inputs = false

func main() {
  // flag init
  flags_should_quit := false // if set quit after flag.Parse()
  remap_name := "Default"

  // flags
  gamepad_name := flag.String("gamepad_name", "gamepadkbdmagic", "the name that will be given to the virtual gamepad") 
  mouse_match_string := flag.String("mouse", "Mouse", "force the use of a specify mouse ex: -mouse \"Mouse Name\".")
  keyboard_match_string := flag.String("keyboard", "Keyboard", "force the use of a specify keyboard ex: -keyboard \"Keyboard Name\".")
  // show inputs 
  flag.BoolVar(&show_inputs, "show_inputs", false, "prints into the cli the events that can be processed")
  // list devices names
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
  // list remap names in portable dir flag
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
  // show version 
  flag.BoolFunc("version", "Show the program version", func(s string) error {
    println(GetVersion())
    flags_should_quit = true
    return nil
  })

  // flag parsing
  flag.Parse()
  if flags_should_quit {
    return
  }
  // file parsing
  args := flag.Args()
  if len(args) == 1 {
    // this is for args after the flags
    remap_name = args[0]
  } else if len(args) > 1 {
    log.Fatal("Too many arguments expected Remap name")
  } 
  log.Info("Trying To Use", remap_name, "Remap file")

  remaps_dir := common.GetRemapsDir()

  remap_filename := remap_name + common.REMAP_FILE_EXT
    
  // reading file
  remap_sc_b, err := os.ReadFile(filepath.Join(remaps_dir, remap_filename))
  if err != nil {
    log.Fatal(err.Error())
  }
  remap_sc := string(remap_sc_b)

  // compiling
  log.Info("Found Remap Compiling...")
  var errList []error
  compileTimeStart := time.Now()

  // send the code for compiling
  RemapTableList, errList := remap.GetRemapTable(remap_sc)
  // check for errors in the compilation step
  if errList != nil {
    for _, e := range errList {
      log.Error(e)
    }
    log.Fatal("Errors Were Reported While Compiling")
  }

  // map RemapTableList to variables
  RemapTable = RemapTableList.Remap
  BoolRemapTable = RemapTableList.Bool
  SequenceTable = RemapTableList.Sequence
  SequenceQueue = make([]buffers.CircularBuf[common.SequenceCmd], len(SequenceTable))
  
  Options = RemapTableList.Opts

  log.Info("Remap Compile Successful, Took:", time.Now().Sub(compileTimeStart).String()) 
  log.Info("Options:", Options)

  // get device list
  devices, err := evdev.ListInputDevices()
  if err != nil {
    log.Fatal(err.Error())
  }
  
  // picking devices from list
  var keyboard *evdev.InputDevice
  var mouse *evdev.InputDevice
  for _, dev := range devices {
    //input0 seems to be the real device
    if !strings.Contains(dev.Phys, "input0") {
      continue
    }
    if strings.Contains(dev.Name, *keyboard_match_string) && keyboard == nil {
      log.Info(dev.Name, "Found as a Keyboard")
      keyboard = dev 
      defer keyboard.Release()
    }
    if strings.Contains(dev.Name, *mouse_match_string) && mouse == nil {
      log.Info(dev.Name, "Found as a Mouse")
      mouse = dev 
      defer mouse.Release()
    }
  }
  if mouse == nil {
    log.Fatal("Couldn't find a mouse")
  }
  if keyboard == nil {
    log.Fatal("Couldn't find a keyboard")
  }

  // controller creation
  controller, err := controller.NewGamepad(*gamepad_name) 
  if err != nil {
    log.Fatal(err.Error())
  }
  defer controller.Close()
  // controller init
  controller.LeftTriggerForce(-1)
  controller.RightTriggerForce(-1)

  lstick := Stick{
    rollBuf: buffers.NewRolloverBuf[vec.Vector2](Options.RollSize),
  }
  rstick := Stick{
    rollBuf: buffers.NewRolloverBuf[vec.Vector2](Options.RollSize),
  }

  // device grab channel
  tgrab := make(chan int)

  // execution
  go RunInput(keyboard, tgrab, controller, &lstick, &rstick)
  go RunInput(mouse, tgrab, controller, &lstick, &rstick)
  go UpdateController(controller, &lstick, &rstick)

  // main thread
  for {
    var d int
    d = <-tgrab // this will block
    // handle program quiting
    if d == 2 {
      break
    }

    // handle grab and ungrab
    var err error 
    if keyboard_grabed {
      err = keyboard.Release()
    } else {
      err = keyboard.Grab()
    }
    if err != nil {
      log.Fatal(err.Error())
    }
    keyboard_grabed = !keyboard_grabed

    if mouse_grabed {
      err = mouse.Release()
    } else {
      err = mouse.Grab()
    }
    if err != nil {
      log.Fatal(err.Error())
    }
    mouse_grabed = !mouse_grabed
  }
}

// linerar interpolation floating
func lerpf(from float64, to float64, weight float64) float64 {
  return from * (1.0 - weight) + (to * weight)
}

type Stick struct {
  // real stick value used by UpdateController
  pos vec.Vector2 
  // mouse relative position accumulation
  velPos vec.Vector2 
  // buffer to hold past velPos values used by UpdateController
  rollBuf buffers.RolloverBuf[vec.Vector2] 
  // bool/absolute position of the stick this takes piority over velPos and rollBuf
  boolPos vec.Vector2
  m sync.Mutex
}

// Update *** Queue functions are very simple just loop over each buffer 
// but only execute that buffer if that queue state time is before time.Now()
// if it is after is because the queue set it's state time to be after x amount of time 
// 
// done this way to only spawn 1 go routine for each use case

func UpdateClickQueue(controller controller.GamepadInterface, lstick *Stick, rstick *Stick) {
  for i := range ClickButtonQueue {
    ClickButtonQueue[i] = buffers.NewCircularBuf[common.OutputCmd](128)
  }
  var clickState []time.Time = make([]time.Time, len(ClickButtonQueue))
  for i := range len(clickState) {
    clickState[i] = time.Now()
  }
  for {
    for i := range len(ClickButtonQueue) {
      cBuf := &ClickButtonQueue[i]
      for cBuf.CanRead() {
        if clickState[i].After(time.Now()) {
          break
        }
        cmd := cBuf.Read()
        switch cmd.Type {
        case common.BUTTON_CMD:
          DoButton(cmd, 0, controller)
        case common.ABS_CMD:
          DoAxis(cmd, 1, lstick, rstick)
        case common.DELAY_CMD:
          clickState[i] = time.Now().Add(cmd.Delay)
          break
        }
      }
    }
    time.Sleep(time.Second / time.Duration(Options.PullRate))
  }
}

func UpdateSequenceQueue(controller controller.GamepadInterface, lstick *Stick, rstick *Stick) {
  for i := range len(SequenceQueue) {
    SequenceQueue[i] = buffers.NewCircularBuf[common.SequenceCmd](256)
  }
  var sequenceState []time.Time = make([]time.Time, len(SequenceQueue))
  for i := range len(sequenceState) {
    sequenceState[i] = time.Now()
  }
  for {
    for i := range len(SequenceQueue) {
      cBuf := &SequenceQueue[i]
      shouldWait := false
      for cBuf.CanRead() {
        if sequenceState[i].After(time.Now()) {
          break
        }
        sCmd := cBuf.Read()
        cmd := sCmd.Cmd
        timeToWait := time.Duration(0)
        switch cmd.Type {
        case common.BUTTON_CMD:
          DoButton(cmd, 0, controller)
          if cmd.Op == common.CLICK_OP {
            shouldWait = true
            timeToWait = max(timeToWait, cmd.Delay)
          }
          if sCmd.Plus {
            continue
          }
        case common.ABS_CMD:
          DoAxis(cmd, 1, lstick, rstick)
          if cmd.Op == common.CLICK_OP {
            shouldWait = true
            timeToWait = max(timeToWait, cmd.Delay)
          }
          if sCmd.Plus {
            continue
          }
        case common.DELAY_CMD:
          sequenceState[i] = time.Now().Add(cmd.Delay)
          break
        }
        if shouldWait {
          sequenceState[i] = time.Now().Add(timeToWait)
          break
        }
      }
    }
    time.Sleep(time.Second / time.Duration(Options.PullRate))
  }
}

func UpdateStickPos(controller controller.GamepadInterface, lstick *Stick, rstick *Stick) {
  sticks := []*Stick{lstick, rstick}
  for {
    for _, stick := range sticks {
      stick.m.Lock()

      stick.rollBuf.Add(stick.velPos)

      //roll buf 
      velo := vec.DivideF(stick.rollBuf.Sum(vec.Sum), float64(stick.rollBuf.Size()))

      StickPoints := Options.StickPoints / float64(stick.rollBuf.Size())

      veloMagSquared := vec.MagnitudeSquared(velo)
      veloMagClamped := common.Clamp(veloMagSquared / (StickPoints * StickPoints), 0.0, 1.0)

      // Translate
      stick.pos = vec.MultiplyF(velo.Div(StickPoints), lerpf(Options.WeightStart, Options.WeightEnd, veloMagClamped))

      // map it to the non deadzoned region
      stick.pos = stick.pos.Mul(1.0 - Options.DeadZone)
      if !stick.pos.IsZero() {
        normStickPos := vec.Normalize(stick.pos)
        stick.pos = vec.Sum(stick.pos, normStickPos.Mul(Options.DeadZone))
      }

      // consume the velocity
      stick.velPos = vec.Vector2{}

      // Tanslate Boolean Keys to Analog
      if !stick.boolPos.IsZero() {
        stick.pos = stick.boolPos
      }
      
      stick.pos = vec.NormalizePastUnit(stick.pos)

      stick.m.Unlock()
    }

    controller.LeftStickMove(lstick.pos.Unpack32())
    controller.RightStickMove(rstick.pos.Unpack32())

    time.Sleep(time.Second / time.Duration(Options.PullRate))
  }
}

func UpdateController(controller controller.GamepadInterface, lstick *Stick, rstick *Stick) {
  go UpdateClickQueue(controller, lstick, rstick)
  go UpdateSequenceQueue(controller, lstick, rstick)
  go UpdateStickPos(controller, lstick, rstick)
}

func DoSequence(cmd common.OutputCmd) {
  sW := len(SequenceTable[cmd.Code])
  if SequenceQueue[cmd.Code].SpaceLeftToWrite() > sW {
    for _, sCmd := range SequenceTable[cmd.Code] {
      SequenceQueue[cmd.Code].Write(sCmd)
    }
  } else {
    log.Warn("Droped SequenceCmd because it couldn't fit into the queue")
  }
}

func DoBoolCmd(cmd common.OutputCmd, sIndex common.StatedIndex, controller controller.GamepadInterface, lstick *Stick, rstick *Stick) {
  // a mutex is needed here because 2 bool remaps can write to the same remap in diferent devices/goroutines
  // idk how much the above case can be useful so i didn't made the emitter block that case so we need mutexes 
  var rewindList []common.StatedIndex

  RemapTableMutex.Lock()
  for _, v := range BoolRemapTable[sIndex] {
    RemapTable[v.SIdx] = v.Cmd

    if StateTable[v.SIdx] != 0 && ecodes.IsStateIndexOn(v.SIdx) && v.Cmd.Op != common.CLICK_OP {
      rewindList = append(rewindList, v.SIdx)
    }
  }
  RemapTableMutex.Unlock()

  for _, si := range rewindList {
    var rwCmd common.OutputCmd
    //undo before bool's overwrite
    rwCmd = RemapUndo[si-1]
    switch rwCmd.Type {
    case common.BUTTON_CMD:
      DoButton(rwCmd, 0, controller)
    case common.ABS_CMD:
      DoAxis(rwCmd, 0, lstick, rstick)
    }
    //mark bool's undo
    RemapUndo[si-1] = RemapTable[si-1]
    //fastforward bool's press
    rwCmd = RemapTable[si]
    switch rwCmd.Type {
    case common.BUTTON_CMD:
      DoButton(rwCmd, StateTable[si], controller)
    case common.ABS_CMD:
      DoAxis(rwCmd, StateTable[si], lstick, rstick)
    }
  }
}

func DoClickCMD(cmd common.OutputCmd) {
  idx := ecodes.MapGPToIndex(cmd.Code)
  // press
  pCmd := cmd
  pCmd.Op = common.PRESS_OP
  ClickButtonQueue[idx].Write(pCmd)

  // wait
  dCmd := common.OutputCmd{}
  dCmd.Delay = cmd.Delay
  dCmd.Type = common.DELAY_CMD
  ClickButtonQueue[idx].Write(dCmd)
  
  // release
  rCmd := cmd
  rCmd.Op = common.RELEASE_OP
  ClickButtonQueue[idx].Write(rCmd)

  // click queue delay
  dCmd = common.OutputCmd{}
  dCmd.Delay = Options.DefaultClickQueueDelay
  dCmd.Type = common.DELAY_CMD
  ClickButtonQueue[idx].Write(dCmd)
}

func DoButton(cmd common.OutputCmd, evalue int32, controller controller.GamepadInterface) {
  //assuming evalue is either 0 or 1
  switch cmd.Op {
  case common.PRESS_OP:
    evalue = 1
  case common.RELEASE_OP:
    evalue = 0
  case common.CLICK_OP:
    DoClickCMD(cmd)
    return
  case common.TOGGLE_OP:
    log.Error("can't run toggle op inside button function")
    return
  }
  btn := cmd.Code
  //this lets multiple presses to the same gp_btn only releasing when all are done
  p := PressedKeys[btn] 
  //either 0 or more than 1 and evalue either 1 or -1 
  PressedKeys[btn] = max(p + int(evalue*2-1), 0)
  //clamp this back down to 1 or 0
  value := min(PressedKeys[btn], 1)

  switch btn {
  case ecodes.BTN_TL2: 
  controller.LeftTriggerForce(float32(value*2-1))
  case ecodes.BTN_TR2:
  controller.RightTriggerForce(float32(value*2-1))
  } 

  if value == 1 {
    controller.ButtonDown(btn)
  } else {
    controller.ButtonUp(btn)
  }
}

func DoAxis(cmd common.OutputCmd, value int32, lstick *Stick, rstick *Stick) {
  daxis_value := float64(value) * float64(cmd.Force) * Options.MouseSense
  
  switch cmd.Code {
  //Left Stick Axis
  case ecodes.ABS_X:
  lstick.m.Lock()
  lstick.velPos.X += daxis_value
  lstick.m.Unlock()
  case ecodes.ABS_Y:
  lstick.m.Lock()
  lstick.velPos.Y += daxis_value
  lstick.m.Unlock()

  //Right Stick Axis
  case ecodes.ABS_RX:
  rstick.m.Lock()
  rstick.velPos.X += daxis_value
  rstick.m.Unlock()
  case ecodes.ABS_RY:
  rstick.m.Lock()
  rstick.velPos.Y += daxis_value
  rstick.m.Unlock()

  default:
    switch cmd.Op {
    case common.PRESS_OP:
      value = 1
    case common.RELEASE_OP:
      value = 0
    case common.CLICK_OP:
      DoClickCMD(cmd)
      return
    case common.TOGGLE_OP:
      log.Error("can't run toggle op inside button function")
      return
    }
    axis_value := float64(value*2-1) * float64(cmd.Force)

    switch cmd.Code {
    //Left X
    case ecodes.ABS_X_POSITIVE:
    lstick.m.Lock()
    lstick.boolPos.X += axis_value
    lstick.m.Unlock()
    case ecodes.ABS_X_NEGATIVE:
    lstick.m.Lock()
    lstick.boolPos.X -= axis_value
    lstick.m.Unlock()

    //Left Y
    case ecodes.ABS_Y_POSITIVE:
    lstick.m.Lock()
    lstick.boolPos.Y += axis_value
    lstick.m.Unlock()
    case ecodes.ABS_Y_NEGATIVE:
    lstick.m.Lock()
    lstick.boolPos.Y -= axis_value
    lstick.m.Unlock()

    //Right X
    case ecodes.ABS_RX_POSITIVE:
    rstick.m.Lock()
    rstick.boolPos.X += axis_value
    rstick.m.Unlock()
    case ecodes.ABS_RX_NEGATIVE:
    rstick.m.Lock()
    rstick.boolPos.X -= axis_value
    rstick.m.Unlock()

    //Right Y
    case ecodes.ABS_RY_POSITIVE:
    rstick.m.Lock()
    rstick.boolPos.Y += axis_value
    rstick.m.Unlock()
    case ecodes.ABS_RY_NEGATIVE:
    rstick.m.Lock()
    rstick.boolPos.Y -= axis_value
    rstick.m.Unlock()
    }
  }
} 

func RunInput(id *evdev.InputDevice, tgrab chan int, controller controller.GamepadInterface, lstick *Stick, rstick *Stick) {
  for {
    events, err := id.Read()
    if err != nil {
      log.Fatal(err.Error())
    }
    var stateIndexList []common.StatedIndex = []common.StatedIndex{}
    for _, event := range events {
      if event.Code == evdev.KEY_PAUSE && event.Value == 0 { 
        // yes it is inverted
        if stop {
          log.Normal("Gamepad Mode: ON")
        } else {
          log.Normal("Gamepad Mode: OFF")
        }
        tgrab <- 1
        stop = !stop
      }
      if event.Code == evdev.KEY_DELETE && event.Value == 0 {
        log.Normal("Exiting...")
        tgrab <- 2
        stop = !stop
      }
      // if event.Code == evdev.KEY_HOME && event.Value == 0 {
      //   log.Info(SequenceTable)
      // }
      if stop {
        continue
      }

      // only events we support
      if event.Type != evdev.EV_KEY && event.Type != evdev.EV_REL {
        continue
      }

      //remove key repeat without writing to the device file
      if event.Type == 1 && event.Value == 2 { 
        continue
      }

      index := ecodes.NormalizeFromInputSys(event.Code, event.Type)
    
      stateIndex := ecodes.ToStateIndex(index, event.Value)

      stateIndexList = append(stateIndexList, stateIndex) 
      if event.Type == evdev.EV_REL {
        StateTable[stateIndex] = event.Value
      } else {
        StateTable[stateIndex+stateIndex%2] = event.Value
        StateToggleTable[stateIndex] = !StateToggleTable[stateIndex]
      }

      if show_inputs {
        log.Normal(event.String(), "\n\tcommand:", RemapTable[stateIndex].String())
      }
    }

    for _, sIndex := range stateIndexList {
      var outCmd common.OutputCmd
      outCmd = RemapTable[sIndex]
      if outCmd.Op != common.CLICK_OP && outCmd.Op != common.TOGGLE_OP {
        if ecodes.IsStateIndexOn(sIndex) {
          RemapUndo[sIndex - 1] = RemapTable[sIndex - 1] 
        } else {
          outCmd = RemapUndo[sIndex]
        }
      } else if outCmd.Op == common.TOGGLE_OP {
        if ecodes.IsStateIndexOn(sIndex) {
          if StateToggleTable[sIndex] {
            RemapUndo[sIndex - 1] = RemapTable[sIndex - 1] 
            outCmd.Op = common.PRESS_OP
          } else {
            outCmd = RemapUndo[sIndex - 1]
            sIndex = sIndex - 1
            outCmd.Op = common.RELEASE_OP
          }
        } else {
          continue
        }
      }
      switch outCmd.Type {
      case common.BUTTON_CMD:
        DoButton(outCmd, StateTable[sIndex], controller)
      case common.ABS_CMD:
        DoAxis(outCmd, StateTable[sIndex], lstick, rstick)
      case common.SEQUENCE_CMD:
        DoSequence(outCmd)
      }

      if outCmd.Bcmd {
        DoBoolCmd(outCmd, sIndex, controller, lstick, rstick)
      }
    }
  }
}

//part of the old stick logic keeping here for reference or skill issue
// this relied on some old behavior (that i programed) like 
// returning 1 if the squared mag was less than 1 else do the square root operation
/*
// more trash
// const ReCenterSpeed float64 = 0.075 // 0-1 
// const ReCenterPoints float64 = StickPoints * ReCenterSpeed
// const ReCenterZone float64 = 0.5
// const DeadZonePoints float64 = StickPoints * DeadZone
// const SlingshotZone float64 = 0.195
// const SlingshotZonePoints float64 = StickPoints * SlingshotZone 
  sticks := []*Stick{lstick, rstick}
  for {
    start := time.Now()


    //sticks
    for _, stick := range sticks {
      stick.m.Lock()

      // Edge Ring
      // stick.relx = clamp(stick.relx, -StickPoints, StickPoints) 
      // stick.rely = clamp(stick.rely, -StickPoints, StickPoints)

      // cvx := stick.relx - stick.vx
      // cvy := stick.rely - stick.vy
      

      vx := RoolRead(&stick.rax)
      vy := RoolRead(&stick.ray)

      // Translate
      stick.x = vx / StickPoints
      stick.y = vy / StickPoints

      // Recenter
      // relmag := VecMagZeroSafe(stick.x, stick.y) 
      // if relmag > DeadZonePoints {
      //   // if stick.x > ReCenterZone || vx == 0 || stick.x < -ReCenterZone {
      //   log.Println(ReCenterPoints * (float32(math.Abs(float64(stick.relx))) / StickPoints) * (stick.x / (relmag + 1)))
      //     stick.relx -= ReCenterPoints * (float32(math.Abs(float64(stick.relx))) / StickPoints) * (stick.x / (relmag + 2))
      //   // }
      //   // if stick.y > ReCenterZone || vy == 0 || stick.y < -ReCenterZone {
      //     stick.rely -= ReCenterPoints * (float32(math.Abs(float64(stick.rely))) / StickPoints) * (stick.y / (relmag + 2))
      //   // }
      // }

      // if stick.x > DeadZone  {
      //   stick.relx -= ReCenterPoints * MathAbs(stick.x - DeadZone)
      // } else if stick.x < -DeadZone {
      //   stick.relx += ReCenterPoints * MathAbs(stick.x + DeadZone)
      // }
      // if stick.y > DeadZone  {
      //   stick.rely -= ReCenterPoints * MathAbs(stick.x - DeadZone)
      // } else if stick.y < -DeadZone {
      //   stick.rely += ReCenterPoints * MathAbs(stick.x + DeadZone)
      // }
      //
      // } else if vx > 0 || vy > 0 {
      //   nvx, nvy := VecNormalize(vx, vy)
      //   stick.relx = nvx * SlingshotZonePoints
      //   stick.rely = nvy * SlingshotZonePoints
      // }
      //
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
*/
