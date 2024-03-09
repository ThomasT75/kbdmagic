package remap

import (
	"errors"
	"log"
	"strings"
	"unicode/utf8"

  "kbdmagic/ecodes"
)

const (
  EOF int = -2
	NOT_KIND int = -1 
	KEY int = iota
	REL
	BTN

	GP_BTN
	GP_AXIS

  INT
	FLOAT

	DELAY_MS

	COMMENT
	NL

  COMMA

  OFF
  ON

  COLON

  LEFT_BRACKET
  RIGHT_BRACKET

  LEFT_BRACES
  RIGHT_BRACES

  RELEASE
  HOLD

  TOGGLE

  PLUS
  MINUS
)

type token struct {
	text string
	kind int
  line int
  column int
}

// lexer
const (
  BREAK_ON_IDK int = iota
  
  BREAK_ON_NL
)

func lexer(source_code string) []token {
  var tokens []token
  var rune_text []rune
  var break_on int = BREAK_ON_IDK
  var line int = 1
  // var column int = 0
  //support utf8 + reread is what this loop does 
  //if you want to reread a value next iteration set w = 0 
  for i, w := 0, 0; i < len(source_code); i += w {
    var r rune
    r, w = utf8.DecodeRuneInString(source_code[i:])
    
    peak := func() rune {
      var ret rune = utf8.RuneError
      if i+w < len(source_code) {
        ret, _ = utf8.DecodeRuneInString(source_code[i+w:])
      }
      return ret
    }

    rune_text = append(rune_text, r)
    switch break_on {
    case BREAK_ON_IDK:
      switch r {
      case '\n':
        line += 1
        if rune_text[0] == '\n' {
          if peak() == '\n' {
            continue
          } 
        } else {
          rune_text = remove_last(rune_text)
          line -= 1
          w = 0 
        }
      case ' ', '\t':
        if peak() == ' ' || peak() == '\t' {
          rune_text = remove_last(rune_text)
          continue
        }
        rune_text = remove_last(rune_text)
      case ',', ':', '[', ']', '{', '}', '-', '+':
        if len(rune_text) > 1 {
          rune_text = remove_last(rune_text)
          w = 0 
        }
      case '/':
        if len(rune_text) == 2 {
          if rune_text[len(rune_text)-2] == '/' {
            break_on = BREAK_ON_NL
          }  
        }
        continue
      case '1', '2', '3', '4', '5',
           '6', '7', '8', '9', '0':
        if is_number(string(rune_text[0])) {
          if is_number(string(peak())) || peak() == '.' {
            continue
          }
        }
      default: continue
      }
    case BREAK_ON_NL:
      if r != '\n' {
        continue
      }
      rune_text = remove_last(rune_text)
      w = 0
      break_on = BREAK_ON_IDK
    }

    text := string(rune_text)
    kind := get_kind(text)
    if text == "" {
      continue
    }
    if kind == NOT_KIND {
      log.Println("At line", line, "Invalid token:", strings.ReplaceAll(text, "\n", "\\n"))
    } else {
      tokens = append(tokens, token{
        kind: kind,
        text: text,
        line: line,
      })
    }

    rune_text = []rune{}
  }

  return tokens
}

func remove_last[T any](s []T) []T {
  if len(s) > 0 {
    s = s[:len(s)-1]
  }
  return s
} 

func is_number(text string) bool {
  is_it := true
  for _, c := range text {
    is_num := false
    for _, n := range "0123456789" {
      if c == n {
        is_num = true
        break
      }
    }
    if !is_num {
      is_it = false
      break
    }
  }
  return is_it
}

func get_kind(text string) int {
	if strings.HasPrefix(text, "KEY_") {
    _, ok := ecodes.KEY_MAP[text]
    if ok {
      return KEY
    }
  }
  if strings.HasPrefix(text, "REL_") {
    _, ok := ecodes.REL_MAP[text]
    if ok {
      return REL
    }
  }
  if strings.HasPrefix(text, "BTN_") {
    _, ok := ecodes.BTN_MAP[text]
    if ok {
      return BTN
    }
  }
  //now do these
  if strings.HasPrefix(text, "GP_BTN_") {
    _, ok := ecodes.GP_BTN_MAP[text]
    if ok {
      return GP_BTN
    }
  }
  if strings.HasPrefix(text, "GP_AXIS_") {
    _, ok := ecodes.GP_AXIS_MAP[text]
    if ok {
      return GP_AXIS
    }
  }
  if strings.HasPrefix(text, "//") {
    return COMMENT
  }
  if is_number(text) {
    return INT
  }
  if strings.HasPrefix(text, "0.") || strings.HasPrefix(text, "1.") {
    s := strings.Split(text, ".")
    if len(s) == 2 {
      if is_number(s[len(s)-1]) {
        return FLOAT
      }
    }
  }
	switch text {
  case "hold": return HOLD
  case "release": return RELEASE
  case "toogle": return TOGGLE
  case ":": return COLON
  case "[": return LEFT_BRACKET
  case "]": return RIGHT_BRACKET
  case "{": return LEFT_BRACES
  case "}": return RIGHT_BRACES
  case ",": return COMMA
  case "ON": return ON 
  case "OFF": return OFF
  case "ms": return DELAY_MS
  case "-": return MINUS
  case "+": return PLUS
	}
  is_nl := true
  for _, v := range text {
    if v != '\n' {
      is_nl = false
      break
    }
  }
  if is_nl {
    return NL
  }
  return NOT_KIND
}

// parser
type token_sequence struct {
  token_sequence []token
  index int 
}

func new_token_sequence(token_list []token) token_sequence {
  var clean_token_list []token
  for _, v := range token_list {
    switch v.kind {
    case COMMENT: continue
    default: clean_token_list = append(clean_token_list, v)
    }
  }
  return token_sequence{
    token_sequence: clean_token_list,
    index: 0,
  }
}

func (ts *token_sequence) next_token() {
  ts.index += 1
}

func (ts *token_sequence) get_current_token() *token {
  if ts.index < len(ts.token_sequence) {
    return &ts.token_sequence[ts.index]
  }
  return &token{
    kind: EOF,
  }
} 

func (ts *token_sequence) get_next_token() *token {
  i := ts.index + 1
  if i < len(ts.token_sequence) {
    return &ts.token_sequence[i]
  }
  return &token{
    kind: EOF,
  }
}

func (ts *token_sequence) check(kind int) bool {
  if ts.get_current_token().kind == kind {
    return true
  }
  return false
}

func (ts *token_sequence) peak(kind int) bool {
  if ts.get_next_token().kind == kind {
    return true
  }
  return false
}

func (ts *token_sequence) accept(kind int) bool {
  r := ts.check(kind)
  if r {
    ts.next_token()
  }
  return r
}

func (ts *token_sequence) expect(kind int) {
  if ts.get_current_token().kind != kind {
    ts.abort("Expected:", kind, "got:", ts.get_current_token().kind)
  }
  ts.next_token()
}

func (ts *token_sequence) abort(msg ...any) {
  v := []any{"At line", ts.get_current_token().line}
  v = append(v, msg...)
  log.Fatal(v)
}

func parser(ts token_sequence) {
    
}

func remap(ts token_sequence) {
  mode := true
  if ts.accept(TOGGLE) {

  } else {
    mode = false
  }
  if ts.check(KEY) {

    ts.next_token()
    if ts.accept(COLON) {
      gpbutton(ts)
    } else {
      macro(ts)
    }
  } else if ts.check(REL) && (ts.peak(PLUS) || ts.peak(MINUS)) {
    // r := ts.get_current_token().text

    ts.next_token()
    if ts.accept(PLUS) {

    } else if ts.accept(MINUS) {

    }
    if ts.accept(COLON) {
      gpbutton(ts)
    } else {
      macro(ts)
    }
  } else if !mode {
    if ts.check(REL) {

      ts.next_token()
      ts.expect(COLON)
      gpaxis(ts)
    }
  } else {
    ts.abort("Expected REMAP")
  }
}

func macro(ts token_sequence) {
  ts.expect(LEFT_BRACKET)

  if ts.check(INT) && ts.peak(DELAY_MS) {

  } else {
    if ts.accept(HOLD) {

    } else if ts.accept(RELEASE) {

    }
    
    if ts.check(GP_BTN) {

    }
  }


  ts.expect(RIGHT_BRACKET)

}

func gpbutton(ts token_sequence) {
  if ts.check(GP_BTN) {
    
    ts.next_token()
  } else if ts.check(GP_AXIS) && (ts.peak(PLUS) || ts.peak(MINUS)) {
    ts.next_token()
    if ts.accept(PLUS) {

    } else if ts.accept(MINUS) {

    }
    force(ts)
  } else {
    ts.abort("Expected GP_BTN")
  }
}

func gpaxis(ts token_sequence) error {
  if ts.check(GP_AXIS) {

    ts.next_token()
    force(ts)
  } else {
    return errors.New("Expected GP_AXIS")
  }
  return nil
}

//optinal so it shouldn't abort
func force(ts token_sequence)  {
  if ts.check(INT) {

    ts.next_token()
  } else if ts.check(FLOAT) {

    ts.next_token()
  }
}

// emitter

//todo write the emitter before the parse or continue the real program 
//and add remap bool separate rel/key/btn 
