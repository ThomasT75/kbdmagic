package validator

import (
	"fmt"
	"kbdmagic/common"
	"kbdmagic/ecodes"
	"kbdmagic/remap/emitter"
)

// the way this deal with validating the emitter output is by creating the remap table list
// using only what we need from the emitter output and checking some values that are known
// to cause problems
//
// tbh im not happy with this solution but it is way better to maintain this than before
// like i could actually reuse this code instead of having to throw it away
//
// if i were to do this better i would do a multi pass system like
//
// > make list of bool commands
// > create remap table first
// > create bool remap table next
// > populate sequence table with the data so far
//
// right now im doing everything at the same time except for step 1 which works great
// but im thinking about adding more to this remap system and this part of the code is
// the most vulnerable to being rewriten just because of the complexity of adding more
// states
//
// Yes this is the most vulnerable code not parser.go LMAO
//
// leaving notes here for later

// {State:0 SIndex:RELEASED: KEY_2 Command:Type: Button, Code: 545 (GP_BTN_DPAD_DOWN), ReleaseOp, Force: 0, Bool: false, Delay: 0s Plus:false None:false}
func Validate(eoSlice []emitter.EmitterOutput, opts common.Options) (common.RemapTableListType, []error) {
  r := common.RemapTableListType{
    Opts: opts,
  }
  errList := []error{}
  sequence_index := 0
  isBoolList := [ecodes.MAX_STATED_INDEX]bool{}
  var last_bool_index common.StatedIndex = 0
  // list all bool indexes to prevent a race condition
  for _, eo := range eoSlice {
    if eo.Command.Bcmd {
      if !isValidSIndex(eo.SIndex) {
        errList = append(errList, fmt.Errorf("Invalid SIndex: %v", eo))
      } else {
        isBoolList[eo.SIndex] = true
      }
    }
  }

  // construct the remap table list
  for _, eo := range eoSlice {
    if eo.Command.Delay == 0 {
      eo.Command.Delay = opts.DefaultClickDelay
    }
    if eo.Command.Bcmd {
      if !isValidSIndex(eo.SIndex) {
        errList = append(errList, fmt.Errorf("Invalid SIndex: %v", eo))
      }
      last_bool_index = eo.SIndex
    }

    switch eo.State {
    case emitter.RemapState:
      if !isValidSIndex(eo.SIndex) {
        errList = append(errList, fmt.Errorf("Invalid SIndex: %v", eo))
      }
      if eo.None {
        r.Remap[eo.SIndex] = common.OutputCmd{}
        continue
      } 

      is_bool := isBoolList[eo.SIndex]

      switch eo.Command.Type {
      case common.BUTTON_CMD:
        r.Remap[eo.SIndex] = constructButtonCmd(eo, is_bool)
      case common.ABS_CMD:
        r.Remap[eo.SIndex] = constructAbsCmd(eo, is_bool)
      case common.SEQUENCE_CMD:
        r.Remap[eo.SIndex] = constructSequenceCmd(sequence_index, is_bool)
        sequence_index += 1
      default:
        if is_bool {
          emptyCmd := common.OutputCmd{}
          if r.Remap[eo.SIndex] == emptyCmd {
            r.Remap[eo.SIndex] = constructBoolCmd()
          } else {
            r.Remap[eo.SIndex].Bcmd = true
          }
          continue
        } 

        errList = append(errList, fmt.Errorf("Invalid command type for remap: %v", eo))
      }
    case emitter.BoolState:
      if !isValidSIndex(eo.SIndex) {
        errList = append(errList, fmt.Errorf("Invalid SIndex: %v", eo))
      }

      is_bool := isBoolList[eo.SIndex]
      
      switch eo.Command.Type {
      case common.BUTTON_CMD:
        r.Bool[last_bool_index] = append(r.Bool[last_bool_index], common.FromOutputCmd{
          SIdx: eo.SIndex,
          Cmd: constructButtonCmd(eo, is_bool),
        })
      case common.ABS_CMD:
        r.Bool[last_bool_index] = append(r.Bool[last_bool_index], common.FromOutputCmd{
          SIdx: eo.SIndex,
          Cmd: constructAbsCmd(eo, is_bool),
        })
      case common.SEQUENCE_CMD:
        r.Bool[last_bool_index] = append(r.Bool[last_bool_index], common.FromOutputCmd{
          SIdx: eo.SIndex,
          Cmd: constructSequenceCmd(sequence_index, is_bool),
        })
        sequence_index += 1
      default:
        errList = append(errList, fmt.Errorf("Invalid command type for bool remap: %v", eo))
      }

      // if bool state is off also write to remap table if there is nothing there
      if !ecodes.IsStateIndexOn(last_bool_index) {
        emptyCmd := common.OutputCmd{}
        if r.Remap[eo.SIndex] == emptyCmd {
          switch eo.Command.Type {
          case common.BUTTON_CMD:
            r.Remap[eo.SIndex] = constructButtonCmd(eo, is_bool)
          case common.ABS_CMD:
            r.Remap[eo.SIndex] = constructAbsCmd(eo, is_bool)
          case common.SEQUENCE_CMD:
            r.Remap[eo.SIndex] = constructSequenceCmd(sequence_index, is_bool)
          default:
            errList = append(errList, fmt.Errorf("Invalid command type for remap: %v", eo))
          }
        }
      }
    case emitter.SequenceState:
      current_sequence_index := sequence_index - 1

      if current_sequence_index > len(r.Sequence) - 1 {
        r.Sequence = append(r.Sequence, []common.SequenceCmd{})
      }
      switch eo.Command.Type {
      case common.BUTTON_CMD:
        r.Sequence[current_sequence_index] = append(r.Sequence[current_sequence_index], common.SequenceCmd{
          Plus: eo.Plus,
          Cmd: constructButtonCmd(eo, false),
        })
      case common.ABS_CMD:
        r.Sequence[current_sequence_index] = append(r.Sequence[current_sequence_index], common.SequenceCmd{
          Plus: eo.Plus,
          Cmd: constructAbsCmd(eo, false),
        })
      case common.DELAY_CMD:
        r.Sequence[current_sequence_index] = append(r.Sequence[current_sequence_index], common.SequenceCmd{
          Plus: false,
          Cmd: constructDelayCmd(eo),
        })
      default:
        errList = append(errList, fmt.Errorf("Invalid command type for sequence: %v", eo))
      }
    }
  } 

  if len(errList) > 0 {
    return r, errList
  }
  return r, nil
}

func constructBoolCmd() common.OutputCmd {
  cmd := common.OutputCmd{
    Bcmd: true,
  }
  return cmd
}

func constructButtonCmd(eo emitter.EmitterOutput, bCmd bool) common.OutputCmd {
  cmd := common.OutputCmd{
    Type: common.BUTTON_CMD,
    Code: eo.Command.Code,
    Op: eo.Command.Op,
    Bcmd: bCmd,
    Delay: eo.Command.Delay,
  }

  return cmd
}

func constructAbsCmd(eo emitter.EmitterOutput, bCmd bool) common.OutputCmd {
  cmd := common.OutputCmd{
    Type: common.ABS_CMD,
    Code: eo.Command.Code,
    Force: eo.Command.Force,
    Op: eo.Command.Op,
    Bcmd: bCmd,
    Delay: eo.Command.Delay,
  }

  return cmd
}

func constructSequenceCmd(sequence_index int, bCmd bool) common.OutputCmd {
  cmd := common.OutputCmd{
    Type: common.SEQUENCE_CMD,
    Code: sequence_index,
    Bcmd: bCmd,
  }

  return cmd
}

func constructDelayCmd(eo emitter.EmitterOutput) common.OutputCmd {
  cmd := common.OutputCmd{
    Type: common.DELAY_CMD,
    Delay: eo.Command.Delay,
  }
  return cmd
}

func isValidSIndex(si common.StatedIndex) bool {
  return si > 0
}

