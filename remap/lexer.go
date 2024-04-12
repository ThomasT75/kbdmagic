package remap

import (
	"errors"
	"fmt"
	"kbdmagic/ecodes"
	"log"
	"strings"
	"unicode/utf8"
)

// lexer
const (
  BREAK_ON_IDK int = iota
  
  BREAK_ON_NL
)

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
    } else {
      fmt.Println(ecodes.BTN_MAP)
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

func lexer(source_code string) ([]token, *[]error) {
  var tokens []token
  var error_list []error
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
      err_msg := fmt.Sprint("At line ", line, " Invalid token: ", strings.ReplaceAll(text, "\n", "\\n"))
      error_list = append(error_list, errors.New(err_msg))
      log.Println(err_msg)
    } else {
      tokens = append(tokens, token{
        kind: kind,
        text: text,
        line: line,
      })
    }

    rune_text = []rune{}
  }

  if len(error_list) > 0 {
    return tokens, &error_list
  }
  return tokens, nil
}
