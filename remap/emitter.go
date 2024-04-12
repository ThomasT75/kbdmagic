package remap

// emitter

import (
	"errors"
	"kbdmagic/common"
	"kbdmagic/ecodes"
	"log"
)

type outCmd = common.OutputCmd

type emitter struct {
  ecodeText string
  ecode int
  current outCmd
  output common.RemapTableType
  state bool
  bothStates bool
  manualOp bool
  usedEcodes []int
  errList []error
}

func NewEmitter() *emitter {
  emi := &emitter{
    errList: nil,
  }
  emi.reset()

  return emi
}

func (e *emitter) reset() {
  e.ecodeText = ""
  e.ecode = -1
  e.current = outCmd{
    Type: -1,
    Code: -1,
    Op: -1,
    Force: 1,
  }
  e.bothStates = true
  e.manualOp = false
}

func (e *emitter) GetErrorList() *[]error {
  if len(e.errList) > 0 {
    return &e.errList
  }
  return nil
}

//state being false for negative/release and true for positive/pressed
func (e *emitter) SetState(state bool) {
  e.state = state
  e.bothStates = false
}

func (e *emitter) SetEcodeFromText(text string) {
  var ecode int
  var etype uint16 = 1
  ecode, ok := ecodes.KEY_MAP[text]
  if ok {
    goto SET
  }
  ecode, ok = ecodes.BTN_MAP[text]
  if ok {
    goto SET
  }
  ecode, ok = ecodes.REL_MAP[text]
  etype = 2
  if !ok {
    err_msg := "Emitter: failed to get ecode from text"
    log.Panicln(err_msg)
    e.errList = append(e.errList, errors.New(err_msg))
  }

  SET:
  e.ecodeText = text
  e.setEcode(uint16(ecode), etype, e.state)
}

func (e *emitter) setEcode(ecode uint16, etype uint16, state bool) {
  i := ecodes.FromEvdev(ecode, etype)
  var evalue int32 = 0
  if state { evalue = 1 }
  i = ecodes.ToStateIndex(i, evalue)
  e.ecode = i
}

func (e *emitter) SetType(t int) {
  e.current.Type = t
}

func (e *emitter) SetCode(c int) {
  e.current.Code = c
}

func (e *emitter) SetForce(f float32) {
  e.current.Force = f
}
func (e *emitter) SetOp(o int) {
  e.current.Op = o
  e.manualOp = true
}

func (e *emitter) Submit() {
  e.SetEcodeFromText(e.ecodeText)
  if e.ecode == -1 {
    err_msg := "Emitter: Trying to submit to a invalid ecode"
    log.Panicln(err_msg)
    e.errList = append(e.errList, errors.New(err_msg))
  }
  if e.current.Type == -1 {
    err_msg := "Emitter: Type of event not set before trying to submit"
    log.Panicln(err_msg)
    e.errList = append(e.errList, errors.New(err_msg))
  }
  if e.current.Code == -1 {
    err_msg := "Emitter: Event Code Was not set before submit cmd"
    log.Panicln(err_msg)
    e.errList = append(e.errList, errors.New(err_msg))
  }
  if !e.manualOp {
    if e.state {
      e.current.Op = common.PRESS_OP
    } else {
      e.current.Op = common.RELEASE_OP
    }
  }
  for _, v := range e.usedEcodes {
    if v == e.ecode {
      err_msg := "Emitter: Tried to set the same ecode twice"
      log.Panicln(err_msg)
      e.errList = append(e.errList, errors.New(err_msg))
    }
  }
  e.output[e.ecode] = e.current
  e.usedEcodes = append(e.usedEcodes, e.ecode) 
  // log.Println(e.ecodeText, e.ecode, e.current)
  if e.bothStates {
    e.SetState(!e.state)
    e.SetEcodeFromText(e.ecodeText)
    e.Submit()
  } else {
    e.reset()
  }
}







