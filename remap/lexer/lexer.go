package lexer

import (
	"errors"
	"fmt"
	"kbdmagic/common/options"
	"kbdmagic/ecodes"
	"kbdmagic/remap/token"
	"strings"
	"unicode/utf8"
)

// lexer
const (
  BREAK_ON_IDK int = iota
  
  BREAK_ON_NL
)

func removeLastInSlice[T any](s []T) []T {
  if len(s) > 0 {
    s = s[:len(s)-1]
  }
  return s
} 

func isNumber(text string) bool {
  // loop for each letter in text
  for _, c := range text {
    isNum := false
    // check if this letter is a number 
    for _, n := range "0123456789" {
      if c == n {
        isNum = true
        break
      }
    }
    // if this letter isn't a number return false
    if !isNum {
      return false
    }
  }
  // if we only saw numbers return true
  return true
}

// return a token kind that matches the text given
func getKindFromText(text string) token.TokenKind {
  _, ok := options.OPTIONS_MAP[text]
  if ok {
    return token.OPTION_NAME
  }
	if strings.HasPrefix(text, "KEY_") {
    _, ok := ecodes.KEY_MAP[text]
    if ok {
      return token.KEY
    }
  }
  if strings.HasPrefix(text, "REL_") {
    _, ok := ecodes.REL_MAP[text]
    if ok {
      return token.REL
    }
  }
  if strings.HasPrefix(text, "BTN_") {
    _, ok := ecodes.BTN_MAP[text]
    if ok {
      return token.BTN
    } 
  }
  //now do these
  if strings.HasPrefix(text, "GP_BTN_") {
    _, ok := ecodes.GP_BTN_MAP[text]
    if ok {
      return token.GP_BTN
    }
  }
  if strings.HasPrefix(text, "GP_AXIS_") {
    _, ok := ecodes.GP_AXIS_MAP[text]
    if ok {
      return token.GP_AXIS
    }
  }
  if strings.HasPrefix(text, "//") {
    return token.COMMENT
  }
  if isNumber(text) {
    return token.INT
  }
  if strings.Contains(text, ".") {
    before, after, _ := strings.Cut(text, ".")
    if isNumber(before) && isNumber(after) {
      return token.FLOAT
    }
  }
	switch text {
  case "hold": return token.HOLD
  case "release": return token.RELEASE
  case "toggle": return token.TOGGLE
  case ":": return token.COLON
  case "!": return token.EXCLAMATION
  case "[": return token.LEFT_BRACKET
  case "]": return token.RIGHT_BRACKET
  case "{": return token.LEFT_BRACES
  case "}": return token.RIGHT_BRACES
  case "(": return token.LEFT_PARENTHESIS
  case ")": return token.RIGHT_PARENTHESIS
  case ",": return token.COMMA
  case "ON": return token.ON 
  case "OFF": return token.OFF
  case "ms": return token.DELAY_MS
  case "-": return token.MINUS
  case "+": return token.PLUS
  case "=": return token.EQUAL
  case "skip": return token.SKIP
  case "wait": return token.WAIT
  case "none": return token.NONE
	}
  is_nl := true
  for _, v := range text {
    if v != '\n' {
      is_nl = false
      break
    }
  }
  if is_nl {
    return token.NL
  }
  return token.NOT_KIND
}

// breaks a string down into a token sequence
func Lexer(source_code string) (token.TokenSequence, []error) {
  var tokens []token.Token
  var error_list []error
  var rune_text []rune
  var break_on int = BREAK_ON_IDK
  var line int = 1
  var column int = 0
  //support utf8 + reread is what this loop does 
  //if you want to reread a value next iteration set w = 0 
  for i, w := 0, 0; i < len(source_code); i += w {
    var r rune
    if w > 0 {
      column += 1
    }
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
        if rune_text[0] != '\n' {
          rune_text = removeLastInSlice(rune_text)
          line -= 1
          w = 0 
        }
      case ' ', '\t':
        if peak() == ' ' || peak() == '\t' {
          rune_text = removeLastInSlice(rune_text)
          continue
        }
        rune_text = removeLastInSlice(rune_text)
      case ',', ':', '!', '[', ']', '{', '}', '(', ')', '-', '+', '=':
        if len(rune_text) > 1 {
          rune_text = removeLastInSlice(rune_text)
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
        if isNumber(string(rune_text[0])) {
          if isNumber(string(peak())) || peak() == '.' {
            continue
          }
        } else {
          continue
        }
      default: continue
      }
    case BREAK_ON_NL:
      if r != '\n' {
        continue
      }
      rune_text = removeLastInSlice(rune_text)
      w = 0
      break_on = BREAK_ON_IDK
    }

    text := string(rune_text)
    kind := getKindFromText(text)
    if text == "" {
      continue
    }
    if kind == token.NOT_KIND {
      err_msg := fmt.Sprintf("At line:column %v:%v, Invalid token: %v", line, column, strings.ReplaceAll(text, "\n", "\\n"))
      error_list = append(error_list, errors.New(err_msg))
    } else {
      tokens = append(tokens, token.NewToken(text, kind, line, column))
    }

    rune_text = []rune{}
  }

  ts := token.NewTokenSequence(tokens)
  if len(error_list) > 0 {
    return ts, error_list
  }
  return ts, nil
}
