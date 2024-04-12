package common

import (
	"kbdmagic/ecodes"
	"log"
	"os"
	"path/filepath"
	"time"

)


const REMAP_TABLE_SIZE = ecodes.ECODES_MAX*2
const REMAP_FILE_EXT = ".remap"

type RemapTableType = [REMAP_TABLE_SIZE]OutputCmd

const (
  BUTTON_CMD = 1 + iota
  ABS_CMD
  MACRO_CMD
  BOOL_CMD
)

const (
  PRESS_OP = 1 + iota
  RELEASE_OP
  CLICK_OP
)

//the values here should be evdev ready meaning no ecodes.ToEvdev
//also that function doesn't exist
type OutputCmd struct {
  Type int //event type
  Code int //event code
  Force float32 //used on ABS_CMD
  Op int //press, release, toggle operation
  Delay time.Duration // delay >= Millisecond this is a delay cmd
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
