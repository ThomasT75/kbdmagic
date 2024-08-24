package common

import (
	"cmp"
	"fmt"
	"kbdmagic/common/options"
	"kbdmagic/ecodes"
	"log"
	"os"
	"path/filepath"
	"time"
)


const REMAP_TABLE_SIZE StatedIndex = StatedIndex(ecodes.ECODES_MAX*2)
const REMAP_FILE_EXT = ".remap"

type RemapTableType [REMAP_TABLE_SIZE]OutputCmd
type BoolRemapTableType [REMAP_TABLE_SIZE][]FromOutputCmd
type SequenceTableType [][]SequenceCmd

func (stt SequenceTableType) String() string {
  var s string = "\n"
  for i, v := range stt {
    s += fmt.Sprint(i, ":\n")
    for _, sC := range v {
      s += fmt.Sprintf("\t%v\n", sC)
    }
    s += fmt.Sprint("\n")
  }
  return s
}

type RemapTableListType struct {
  Remap RemapTableType
  Bool BoolRemapTableType
  Sequence SequenceTableType
  Opts Options
} 

type NormalizedIndex = ecodes.NormalizedIndex

type StatedIndex = ecodes.StatedIndex

type Options = options.Options

type CmdType int

const (
  BUTTON_CMD = CmdType(1 + iota)
  ABS_CMD
  DELAY_CMD
  SEQUENCE_CMD
)

var _MAP_TYPE = map[CmdType]string{
  0: "NoType",
  BUTTON_CMD: "Button",
  ABS_CMD: "Abs",
  DELAY_CMD: "Delay",
  SEQUENCE_CMD: "Sequence",
}

type CmdOp int

const (
  PRESS_OP = CmdOp(1 + iota)
  RELEASE_OP
  CLICK_OP
  TOGGLE_OP
)

var _MAP_OP = map[CmdOp]string{
  0: "NoOp",
  PRESS_OP: "PressOp",
  RELEASE_OP: "ReleaseOp",
  CLICK_OP: "ClickOp",
  TOGGLE_OP: "ToggleOp",
}

type FromOutputCmd struct {
  SIdx StatedIndex //state index
  Cmd OutputCmd
}

type SequenceCmd struct {
  Plus SequencePlusType // should combine with next
  Cmd OutputCmd 
}

type SequencePlusType = bool

//the values here should be evdev ready meaning no ecodes.ToEvdev
//also that function doesn't exist (lies)
type OutputCmd struct {
  Type CmdType //event type
  Code int //event code
  Op CmdOp //press, release, toggle operation
  Force float32 //used on ABS_CMD
  Delay time.Duration // delay >= Millisecond this is a delay cmd
  Bcmd bool // if set also acts as a boolRemap
}


func (cmd OutputCmd) String() string {
  var sCode string = "None"
  switch cmd.Type {
  case ABS_CMD:
    for k, c := range ecodes.GP_AXIS_MAP {
      if c == cmd.Code {
        sCode = k
        break
      }
    }
  case BUTTON_CMD:
    s, ok := ecodes.GP_STRING(cmd.Code)
    if ok {
      sCode = s
    }
  }
  return fmt.Sprintf("Type: %+v, Code: %+v (%+v), %+v, Force: %+v, Bool: %+v, Delay: %+v", _MAP_TYPE[cmd.Type], cmd.Code, sCode, _MAP_OP[cmd.Op], cmd.Force, cmd.Bcmd, cmd.Delay)
}

//TODO unify this function instead of having copies of it 
func Clamp[T cmp.Ordered](value T, minium T, maxium T) T {
  return max(minium, min(value, maxium))
}

func GetPortableDir() string {
  execpath, err := os.Executable()
  if err != nil {
    log.Println(err.Error())
    log.Fatal("can't get Executable path")
  }

  execpath, err = filepath.EvalSymlinks(execpath)
  if err != nil {
    log.Println(err.Error())
    log.Fatal("couldn't EvalSymlinks of Executable path")
  }

  execdir := filepath.Dir(execpath)

  portable_dir := filepath.Join(execdir, "portable")
  err = os.Mkdir(portable_dir, 0755)
  if err != nil && !os.IsExist(err) {
    log.Println(err.Error())
    log.Fatal("couldn't create portable dir")
  }

  return portable_dir
}

func GetRemapsDir() string {
  portable_dir := GetPortableDir()

  remaps_dir := filepath.Join(portable_dir, "remaps")
  err := os.Mkdir(remaps_dir, 0755)
  if err != nil && !os.IsExist(err) {
    log.Fatal("couldn't create remaps dir")
  }

  return remaps_dir
}

//wip
func GetVibSoundFile() string {
  portable_dir := GetPortableDir()
  return filepath.Join(portable_dir, "VibSound.mp3")
}
