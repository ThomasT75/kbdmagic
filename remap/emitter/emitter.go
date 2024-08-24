package emitter

// emitter

import (
	"errors"
	"kbdmagic/common"
	"kbdmagic/common/options"
	"kbdmagic/ecodes"
	"kbdmagic/internal/log"
	"strconv"
	"time"
)

type EmitterState int 

type EmitterOutput struct {
  State EmitterState
  SIndex common.StatedIndex
  Command common.OutputCmd
  Plus common.SequencePlusType
  None bool
} 

const (
  RemapState = EmitterState(iota)
  BoolState
  SequenceState
)

type Emitter struct {
  // ecodes
  sIndexText string
  sIndex common.StatedIndex
  // states 
  state bool
  autoSetBothStates bool
  autoSetOp bool

  // finite state machine with a stack
  fsm EmitterState
  fsmStack []EmitterState

  // options
  opts common.Options
  // output
  output []EmitterOutput
  // current should go into output
  current EmitterOutput

  // ocurred errors
  errList []error
}

func NewEmitter() Emitter {
  emi := Emitter{
    errList: nil,
    fsm: RemapState,
    opts: options.Defaults(),
  }
  emi.reset()
  return emi
}

func (e *Emitter) reset() {
  e.sIndexText = ""
  e.sIndex = 0
  e.state = false
  e.autoSetBothStates = true
  e.autoSetOp = true
  e.current = DefaultEmitterOutput()
}

func DefaultEmitterOutput() EmitterOutput {
  return EmitterOutput{
    Command: common.OutputCmd{
      Force: 1,
    },
    None: false, //not needed but safe
  }
}

// used to check if the emitter output is empty 
func isEmitterOutputEmpty(eo EmitterOutput) bool {
  eo.State = 0
  if eo == DefaultEmitterOutput() {
    return true
  }
  return false
}

func (e *Emitter) GetErrorList() []error {
  if len(e.errList) > 0 {
    return e.errList
  }
  return nil
}

func (e *Emitter) SetSIndexFromText(text string) {
  var ecode int
  var etype uint16 = 1
  ecode, ok := ecodes.KEY_MAP[text]
  if ok {
    goto GOAL
  }
  ecode, ok = ecodes.BTN_MAP[text]
  if ok {
    goto GOAL
  } 
  ecode, ok = ecodes.REL_MAP[text]
  if ok {
    etype = 2
    goto GOAL
  } 

  goto ERR

  GOAL:
  e.sIndexText = text
  e.setSIndex(uint16(ecode), etype, e.state)
  return
  ERR:
  var err_msg string
  err_msg = "Emitter: failed to get ecode from text:" + text
  if text == "" {
    err_msg = "Emiiter: ecode text is empty"
  } 
  log.Fatal(err_msg)
  e.errList = append(e.errList, errors.New(err_msg))
}

func (e *Emitter) setSIndex(ecode uint16, etype uint16, state bool) {
  ni := ecodes.NormalizeFromInputSys(ecode, etype)
  var evalue int32 = 0
  if state { evalue = 1 }
  si := ecodes.ToStateIndex(ni, evalue)
  e.sIndex = si
}

func (e *Emitter) enterState(state EmitterState) {
  e.fsmStack = append(e.fsmStack, e.fsm)
  e.fsm = state
}

func (e *Emitter) ExitState() {
  s := e.fsmStack[len(e.fsmStack)-1]
  if len(e.fsmStack) > 1 {
    e.fsmStack = e.fsmStack[:len(e.fsmStack)-1]
  }
  e.fsm = s
}

func (e *Emitter) SetNone() {
  e.current.None = true
}

//state being false for negative/release and true for positive/pressed
func (e *Emitter) SetState(state bool) {
  e.state = state
  e.autoSetBothStates = false
  // state is related to the "ecode" you will get
  e.SetSIndexFromText(e.sIndexText)
}

func (e *Emitter) SetType(t common.CmdType) {
  e.current.Command.Type = t
}

func (e *Emitter) SetCode(c int) {
  e.current.Command.Code = c
}

func (e *Emitter) SetForce(f float32) {
  e.current.Command.Force = f
}

func (e *Emitter) SetDelayFromText(text string) error {
  dur, err := time.ParseDuration(text)
  if err != nil {
    return err
  }
  e.setDelay(dur)
  return nil
}

func (e *Emitter) setDelay(d time.Duration) {
  e.current.Command.Delay = d
}

func (e *Emitter) SetOp(o common.CmdOp) {
  e.current.Command.Op = o
  e.autoSetOp = false
}

func (e *Emitter) Submit() {
  if e.autoSetOp && e.current.Command.Type == common.BUTTON_CMD {
    if e.state {
      e.current.Command.Op = common.PRESS_OP
    } else {
      e.current.Command.Op = common.RELEASE_OP
    }
  }

  e.current.State = e.fsm
  e.current.SIndex = e.sIndex
  if !isEmitterOutputEmpty(e.current) {
    e.output = append(e.output, e.current)
  }

  if e.autoSetBothStates && e.sIndexText != "" { 
    e.SetState(!e.state)
    e.Submit()
  } else {
    e.reset()
  }
}

func (e *Emitter) SubmitOption(name string, value string) error {
  idx, err := options.GetIndexFromText(name)
  if err != nil {
    return err
  } 
  var tValue any
  aType := options.GetType(idx)
  switch aType.(type) {
  case float64:
    tValue, err = strconv.ParseFloat(value, 64)
    if err != nil {
      return err
    }
  case int:
    tValue, err = strconv.ParseInt(value, 10, 0)
    tValue = int(tValue.(int64))
    if err != nil {
      return err
    }
  case time.Duration:
    tValue, err = time.ParseDuration(value)
    if err != nil {
      return err
    }
  }
  e.opts.Set(idx, tValue)
  return nil
}

func (e *Emitter) GetOutput() ([]EmitterOutput, common.Options) {
  return e.output, e.opts
}


// call to enter bool remap mode
func (e *Emitter) EnterBoolState() {
  e.enterState(BoolState)
}

func (e *Emitter) SetBoolCmd() {
  e.current.Command.Bcmd = true
}

func (e *Emitter) EnterSequenceState() {
  e.enterState(SequenceState)
}

func (e *Emitter) SetSequencePlus() {
  e.current.Plus = true
}
