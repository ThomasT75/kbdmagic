package remap

import (
	"errors"
	"fmt"
	"log"
)

const (
  EOF int = -2
	NOT_KIND int = -1 
	KEY int = 1 + iota
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

type token_sequence struct {
  token_sequence []token
  index int 
  err error
}

func NewTokenSequence(token_list []token) *token_sequence {
  var clean_token_list []token
  for _, v := range token_list {
    switch v.kind {
    case COMMENT: continue
    default: clean_token_list = append(clean_token_list, v)
    }
  }
  return &token_sequence{
    token_sequence: clean_token_list,
    index: 0,
    err: nil,
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

func (ts *token_sequence) expect(kind int) bool {
  if ts.get_current_token().kind != kind {
    ts.abort("Expected:", kind, "got:", ts.get_current_token().kind)
    return false
  }
  ts.next_token()
  return true
}

func (ts *token_sequence) abort(msg ...any) {
  v := []any{"At line", ts.get_current_token().line}
  v = append(v, msg...)
  err_msg := fmt.Sprint(v)
  log.Println(err_msg)
  ts.err = errors.New(err_msg)
}
