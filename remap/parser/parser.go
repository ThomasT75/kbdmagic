package parser

import (
	"kbdmagic/common"
	"kbdmagic/ecodes"
	"kbdmagic/remap/emitter"
	"kbdmagic/remap/token"
	"strconv"
)

// if you want to rewrite this don't.
// instead make a parser that follows rules writen in a grammar file

// Parsers a TokenSequence using the provided emitter
// return error on grammar mistakes
func Parser(ts *token.TokenSequence, emi *emitter.Emitter) error {
  if ts.Check(token.EOF) {
    ts.Abort("Empty Remap")
    return ts.GetLastError()
  }
  var OptSection bool = true 
  for {
    for ts.Check(token.NL) {
      ts.NextToken()
    }

    if ts.Check(token.EXCLAMATION) && OptSection {
      err := option(ts, emi)
      if err != nil {
        return err
      }
    } 
    if ts.Check(token.EXCLAMATION) && !OptSection {
      ts.Abort("Options Should be before any Remap")
      return ts.GetLastError()
    }

    for ts.Check(token.NL) {
      ts.NextToken()
    }

    if !ts.Check(token.EXCLAMATION) {
      if isRemap(ts) {
        err := remap(ts, emi)
        if err != nil {
          return err
        }
        OptSection = false
      } else if isBoolRemap(ts) {
        err := boolRemap(ts, emi)
        if err != nil {
          return err
        } 
        OptSection = false
      } else {
        ts.Abort("Expected a Remap or BoolRemap")
        return ts.GetLastError()
      }
    }

    for ts.Check(token.NL) {
      ts.NextToken()
    }

    if ts.Check(token.EOF) {
      break
    }
  }
  return nil
}

func option(ts *token.TokenSequence, emi *emitter.Emitter) error {
  var ok bool
  var option_name string
  var option_value string
  ok = ts.Expect(token.EXCLAMATION)
  if !ok { return ts.GetLastError() }
  // option_name
  if ts.Check(token.OPTION_NAME) {
    option_name = ts.GetText()
    ts.NextToken()
  }

  ok = ts.Expect(token.EQUAL)
  if !ok { return ts.GetLastError() }

  // option_value
  if ts.CheckAny(token.INT, token.FLOAT) {
    option_value = ts.GetText()
    ts.NextToken()
  } else {
    ts.Abort("No valid option_value Type was provided")
  }
  if ts.Check(token.DELAY_MS) {
    option_value += ts.GetText()
    ts.NextToken()
  }

  ok = ts.Expect(token.NL)
  if !ok { return ts.GetLastError() }
  emi.SubmitOption(option_name, option_value)
  return nil
}

func isRemap(ts *token.TokenSequence) bool {
  return ts.Pattern(
    []token.TokenKind{token.TOGGLE, token.ANY},
    []token.TokenKind{token.KEY, token.BTN, token.REL},
    []token.TokenKind{token.PLUS, token.MINUS, token.ANY},
    []token.TokenKind{token.COLON},
  )
}

func remap(ts *token.TokenSequence, emi *emitter.Emitter) error {
  mode := true
  if ts.Accept(token.TOGGLE) {
    emi.SetOp(common.TOGGLE_OP)
  } else {
    mode = false
  }
  if ts.CheckAny(token.KEY, token.BTN) {
    text := ts.GetText()
    emi.SetSIndexFromText(text)
    ts.NextToken()
    if ts.Expect(token.COLON) {
      if ts.Accept(token.NONE) {
        emi.SetNone()
      } else if ts.Check(token.LEFT_BRACKET) {
        err := sequence(ts, emi)
        if err != nil {
          return err
        }
      } else {
        err := gpbutton(ts, emi)
        if err != nil {
          return err
        }
      }
    } else {
      return ts.GetLastError()
    }
  } else if ts.Check(token.REL) && ts.PeakAny(token.PLUS, token.MINUS) {
    text := ts.GetText()
    emi.SetSIndexFromText(text)
    emi.SetOp(common.CLICK_OP)
    ts.NextToken()
    if ts.Accept(token.PLUS) {
      emi.SetState(true)
    } else if ts.Accept(token.MINUS) {
      emi.SetState(false)
    }
    if ts.Expect(token.COLON) {
      if ts.Accept(token.NONE) {
        emi.SetNone()
      } else if ts.Check(token.LEFT_BRACKET) {
        err := sequence(ts, emi)
        if err != nil {
          return err
        }
      } else {
        err := gpbutton(ts, emi)
        if err != nil {
          return err
        }
      }
    } else {
      return ts.GetLastError()
    }
  } else if !mode {
    if ts.Check(token.REL) {
      text := ts.GetText()
      emi.SetSIndexFromText(text)
      ts.NextToken()
      if ts.Expect(token.COLON) {
        if ts.Accept(token.NONE) {
          emi.SetNone()
        } else {
          err := gpaxis(ts, emi)
          if err != nil {
            return err
          }
        }
      } else {
        return ts.GetLastError()
      }
    }
  } else {
    ts.Abort("Expected REMAP")
    return ts.GetLastError()
  }

  ok := ts.Expect(token.NL)
  if !ok {
    return ts.GetLastError()
  }
  emi.Submit()
  return nil
}

func isBoolRemap (ts *token.TokenSequence) bool {
  return ts.Pattern(
    []token.TokenKind{token.TOGGLE, token.ANY},
    []token.TokenKind{token.KEY, token.BTN},
    []token.TokenKind{token.LEFT_BRACES},
  )
}

func boolRemap(ts *token.TokenSequence, emi *emitter.Emitter) error {
  var bool_text string
  var use_bool_op bool = false
  var bool_op common.CmdOp
  var first_state bool
  if ts.Accept(token.TOGGLE) {
    use_bool_op = true
    bool_op = common.TOGGLE_OP
  }
  if ts.CheckAny(token.KEY, token.BTN) {
    bool_text = ts.GetText()
    ts.NextToken()
  } else {
    ts.Abort("Expected key or btn for bool remap trigger")
    return ts.GetLastError()
  }

  //expect
  ok := ts.Expect(token.LEFT_BRACES)
  if !ok { return ts.GetLastError() }
  ok = ts.Expect(token.NL)
  if !ok { return ts.GetLastError() }
  if ts.Check(token.OFF) {
    ts.NextToken()
    first_state = false
  } else if ts.Check(token.ON) {
    ts.NextToken()
    first_state = true
  } else {
    ts.Abort("Expected either ON or OFF in a bool remap")
    return ts.GetLastError()
  }
  ok = ts.Expect(token.NL)
  if !ok { return ts.GetLastError() }

  //first state
  emi.SetSIndexFromText(bool_text)
  emi.SetBoolCmd()
  emi.SetState(first_state)
  if use_bool_op {
    emi.SetOp(bool_op)
  }
  emi.Submit()
  emi.EnterBoolState()

  for isRemap(ts) {
    err := remap(ts, emi)
    if err != nil {
      return err
    }
  }
  
  emi.ExitState()
  
  if !ts.Check(token.RIGHT_BRACES) {
    //expect
    if first_state {
      ok = ts.Expect(token.OFF)
      if !ok { return ts.GetLastError() }
    } else {
      ok = ts.Expect(token.ON)
      if !ok { return ts.GetLastError() }
    }
    ok = ts.Expect(token.NL)
    if !ok { return ts.GetLastError() }

    //second state
    emi.SetSIndexFromText(bool_text)
    emi.SetBoolCmd()
    emi.SetState(!first_state)
    if use_bool_op {
      emi.SetOp(bool_op)
    }
    emi.Submit()
    emi.EnterBoolState()

    for isRemap(ts) {
      err := remap(ts, emi)
      if err != nil {
        return err
      }
    }

    emi.ExitState()
  }
  //expect
  ok = ts.Expect(token.RIGHT_BRACES)
  if !ok { return ts.GetLastError() }
  ok = ts.Expect(token.NL)
  if !ok { return ts.GetLastError() }

  return nil
}

func sequence(ts *token.TokenSequence, emi *emitter.Emitter) error {
  ok := ts.Expect(token.LEFT_BRACKET)
  if !ok { return ts.GetLastError() }

  emi.SetType(common.SEQUENCE_CMD)
  emi.SetState(true)
  emi.Submit()

  emi.EnterSequenceState()
  // setting this to false will disable the check
  var isDelayFirst = true
  // used to check if the previous cmd was plus 
  var prevPlus = false
  for !ts.Check(token.RIGHT_BRACKET) && !ts.Check(token.EOF) {
    if ts.Check(token.INT) && ts.Peak(token.DELAY_MS) {
      if isDelayFirst {
        ts.Abort("Delay as the first element in a Sequence is not allowed")
        return ts.GetLastError()
      }
      if prevPlus {
        ts.Abort("Can't combo a Button + Delay")
      }
      emi.SetType(common.DELAY_CMD)
      delay(ts, emi)
      emi.Submit()
    } else {
      isDelayFirst = false
      //operador
      hadOperator := true
      if ts.Accept(token.HOLD) {
        emi.SetOp(common.PRESS_OP)
      } else if ts.Accept(token.RELEASE) {
        emi.SetOp(common.RELEASE_OP)
      } else {
        hadOperator = false
        emi.SetOp(common.CLICK_OP)
      }

      // button
      err := gpbutton(ts, emi)
      if err != nil {
        return err
      }

      if ts.Accept(token.LEFT_PARENTHESIS) {
        if hadOperator {
          ts.Abort("Can't use a operador + an explicit click delay")
          return ts.GetLastError()
        } 
        delay(ts, emi)
        ok = ts.Expect(token.RIGHT_PARENTHESIS)
        if !ok { return ts.GetLastError() }
      }

      // combo check
      if ts.Accept(token.PLUS) {
        prevPlus = true
        emi.SetSequencePlus()
      } else {
        prevPlus = false
      }

      emi.Submit()
      if prevPlus {
        continue
      }
    }

    if ts.Check(token.RIGHT_BRACKET) {
      break
    }

    ok := ts.Expect(token.COMMA)
    if !ok { return ts.GetLastError() }

    ts.Accept(token.NL)
  }
  emi.ExitState()

  ok = ts.Expect(token.RIGHT_BRACKET)
  if !ok { return ts.GetLastError() }

  return nil
}

// sets the type to common.BUTTON_CMD or common.ABS_CMD
func gpbutton(ts *token.TokenSequence, emi *emitter.Emitter) error {
  if ts.Check(token.GP_BTN) {
    text := ts.GetText()
    emi.SetCode(ecodes.GP_BTN_MAP[text]) 
    emi.SetType(common.BUTTON_CMD)
    ts.NextToken()
  } else if ts.Check(token.GP_AXIS) && ts.PeakAny(token.PLUS, token.MINUS) {
    text := ts.GetText()
    ts.NextToken()
    if ts.Accept(token.PLUS) {
      text += "_POSITIVE"
    } else if ts.Accept(token.MINUS) {
      text += "_NEGATIVE"
    }
    emi.SetCode(ecodes.GP_AXIS_MAP[text]) 
    emi.SetType(common.ABS_CMD)
    err := force(ts, emi)
    if err != nil {
      return err
    }
  } else {
    ts.Abort("Expected GP_BTN")
    return ts.GetLastError()
  }
  
  return nil
}

// sets the type to common.ABS_CMD
func gpaxis(ts *token.TokenSequence, emi *emitter.Emitter) error {
  if ts.Check(token.GP_AXIS) {
    text := ts.GetText()
    emi.SetCode(ecodes.GP_AXIS_MAP[text]) 
    emi.SetType(common.ABS_CMD)
    ts.NextToken()
    err := force(ts, emi)
    if err != nil {
      return err
    }
  } else {
    ts.Abort("Expected GP_AXIS")
    return ts.GetLastError()
  }
  
  return nil
}

func force(ts *token.TokenSequence, emi *emitter.Emitter) error {
  if ts.Check(token.INT) {
    text := ts.GetText()
    v, err := strconv.Atoi(text)
    if err != nil {
      ts.Abort("Parser: can't get int from text of kind INT for force") 
      return ts.GetLastError()
    }
    emi.SetForce(float32(v))
    ts.NextToken()
  } else if ts.Check(token.FLOAT) {
    text := ts.GetText()
    v, err := strconv.ParseFloat(text, 32) //still a float64 value
    if err != nil {
      ts.Abort("Parser: can't get float32 from text of kind FLOAT for force") 
      return ts.GetLastError()
    }
    emi.SetForce(float32(v))
    ts.NextToken()
  }

  return nil
}

// doesn't set type to common.DELAY_CMD
func delay(ts *token.TokenSequence, emi *emitter.Emitter) error {
  if ts.Check(token.INT) && ts.Peak(token.DELAY_MS) {
    text := ts.GetText()
    text += "ms"
    emi.SetDelayFromText(text)
    ok := ts.Expect(token.INT)
    if !ok { return ts.GetLastError() }
    ok = ts.Expect(token.DELAY_MS)
    if !ok { return ts.GetLastError() }
  } else {
    ts.Abort("Expected Delay")
    return ts.GetLastError()
  }

  return nil
}
