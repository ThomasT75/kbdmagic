package remap

import (
	"errors"
	"kbdmagic/common"
	"kbdmagic/ecodes"
	"log"
	"strconv"
)

// parser
func parser(ts *token_sequence, emi *emitter) error {
  if ts.check(EOF) {
    return errors.New("Empty Remap")
  }
  for {
    for ts.check(NL) {
      ts.next_token()
    }
    err := remap(ts, emi)
    if err != nil {
      return err
    }

    ok := ts.expect(NL)
    if !ok {
      return ts.err
    }

    for ts.check(NL) {
      ts.next_token()
    }

    emi.Submit()
    if ts.check(EOF) {
      break
    }
  }
  return nil
}

func remap(ts *token_sequence, emi * emitter) error {
  mode := true
  if ts.accept(TOGGLE) {

  } else {
    mode = false
  }
  if ts.check(KEY) || ts.check(BTN) {
    text := ts.get_current_token().text
    emi.SetEcodeFromText(text)
    ts.next_token()
    if ts.accept(COLON) {
      err := gpbutton(ts, emi)
      if err != nil {
        return err
      }
    } else {
      err := macro(ts, emi)
      if err != nil {
        return err
      }
    }
  } else if ts.check(REL) && (ts.peak(PLUS) || ts.peak(MINUS)) {
    text := ts.get_current_token().text
    emi.SetEcodeFromText(text)
    emi.SetOp(common.CLICK_OP)
    ts.next_token()
    if ts.accept(PLUS) {
      emi.SetState(true)
    } else if ts.accept(MINUS) {
      emi.SetState(false)
    }
    if ts.accept(COLON) {
      err := gpbutton(ts, emi)
      if err != nil {
        return err
      }
    } else {
      err := macro(ts, emi)
      if err != nil {
        return err
      }
    }
  } else if !mode {
    if ts.check(REL) {
      text := ts.get_current_token().text
      emi.SetEcodeFromText(text)
      ts.next_token()
      ok := ts.expect(COLON)
      if !ok {
        return ts.err
      }
      err := gpaxis(ts, emi)
      if err != nil {
        return err
      }
    }
  } else {
    ts.abort("Expected REMAP")
    return ts.err
  }

  return nil
}

func macro(ts *token_sequence, emi * emitter) error {
  ok := ts.expect(LEFT_BRACKET)
  if !ok {
    return ts.err
  }

  if ts.check(INT) && ts.peak(DELAY_MS) {

  } else {
    if ts.accept(HOLD) {

    } else if ts.accept(RELEASE) {

    }
    
    if ts.check(GP_BTN) {

    }
  }


  ok = ts.expect(RIGHT_BRACKET)
  if !ok {
    return ts.err
  }

  return nil
}

func gpbutton(ts *token_sequence, emi * emitter) error {
  if ts.check(GP_BTN) {
    text := ts.get_current_token().text
    emi.SetCode(ecodes.GP_BTN_MAP[text]) 
    emi.SetType(common.BUTTON_CMD)
    ts.next_token()
  } else if ts.check(GP_AXIS) && (ts.peak(PLUS) || ts.peak(MINUS)) {
    text := ts.get_current_token().text
    ts.next_token()
    if ts.accept(PLUS) {
      text += "_POSITIVE"
    } else if ts.accept(MINUS) {
      text += "_NEGATIVE"
    }
    emi.SetCode(ecodes.GP_AXIS_MAP[text]) 
    emi.SetType(common.ABS_CMD)
    force(ts, emi)
  } else {
    ts.abort("Expected GP_BTN")
    return ts.err
  }
  
  return nil
}

func gpaxis(ts *token_sequence, emi *emitter) error {
  if ts.check(GP_AXIS) {
    text := ts.get_current_token().text
    emi.SetCode(ecodes.GP_AXIS_MAP[text]) 
    emi.SetType(common.ABS_CMD)
    ts.next_token()
    force(ts, emi)
  } else {
    return errors.New("Expected GP_AXIS")
  }
  
  return nil
}

//optinal so it shouldn't abort
func force(ts *token_sequence, emi * emitter) {
  if ts.check(INT) {
    text := ts.get_current_token().text
    v, err := strconv.Atoi(text)
    if err != nil {
      log.Fatal("Parser: can't get int from text of kind INT for force") 
    }
    emi.SetForce(float32(v))
    ts.next_token()
  } else if ts.check(FLOAT) {
    text := ts.get_current_token().text
    v, err := strconv.ParseFloat(text, 32) //still a float64 value
    if err != nil {
      log.Fatal("Parser: can't get float32 from text of kind FLOAT for force") 
    }
    emi.SetForce(float32(v))
    ts.next_token()
  }
}
