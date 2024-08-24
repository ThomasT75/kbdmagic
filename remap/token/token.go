package token

import (
	"errors"
	"fmt"
	"strings"
)

// tokens what more could you want

type Token struct {
	text string
	kind TokenKind
  line int
  column int
}

func NewToken(text string, kind TokenKind, line int, column int) Token {
  return Token{
    text: text,
    kind: kind,
    line: line,
    column: column,
  }
}

type TokenSequence struct {
  cleanTokenList []Token
  unwindPoints []int
  index int 
  err error
}

func NewTokenSequence(tokenList []Token) TokenSequence {
  var cleanTokenList []Token
  
  // clean the token list before use
  var c bool = false
  for _, v := range tokenList {
    switch v.kind {
    case COMMENT: 
      c = true
      continue
    default: 
      if c == true && v.kind == NL {
        c = false
        continue
      }
      cleanTokenList = append(cleanTokenList, v)
    }
  }

  return TokenSequence{
    cleanTokenList: cleanTokenList,
    index: 0,
    err: nil,
  }
}

// Unwind() return to this point/state in the token sequence 
func (ts *TokenSequence) setUnwindPoint() {
  ts.unwindPoints = append(ts.unwindPoints, ts.index)
}

// Returns to the last point set by SetUnwindPoint()
func (ts *TokenSequence) unwind() {
  ts.index = ts.unwindPoints[len(ts.unwindPoints)-1]
  ts.discardLastUnwindPoint()
}

// Discard the last point in case you didn't needed to Unwind() 
func (ts *TokenSequence) discardLastUnwindPoint() {
  if len(ts.unwindPoints) > 0 {
    ts.unwindPoints = ts.unwindPoints[:len(ts.unwindPoints)-1]
  }
}

// Call to advance to the next token in the sequence
func (ts *TokenSequence) NextToken() {
  ts.index += 1
}

func (ts *TokenSequence) getCurrentToken() *Token {
  if ts.index < len(ts.cleanTokenList) {
    return &ts.cleanTokenList[ts.index]
  }
  return &Token{
    kind: EOF,
  }
} 

func (ts *TokenSequence) getNextToken() *Token {
  i := ts.index + 1
  if i < len(ts.cleanTokenList) {
    return &ts.cleanTokenList[i]
  }
  return &Token{
    kind: EOF,
  }
}

// Returns the current token text
func (ts *TokenSequence) GetText() string {
  return ts.getCurrentToken().text
}

// Checks if the current token matches the kind arg
func (ts *TokenSequence) Check(kind TokenKind) bool {
  if ts.getCurrentToken().kind == kind {
    return true
  }
  return false
}

// Same as Check() but for the next token 
// and without calling NextToken() first
func (ts *TokenSequence) Peak(kind TokenKind) bool {
  if ts.getNextToken().kind == kind {
    return true
  }
  return false
}

// Calls NextToken() if the current token matches the kind arg
func (ts *TokenSequence) Accept(kind TokenKind) bool {
  r := ts.Check(kind)
  if r {
    ts.NextToken()
  }
  return r
}

// Checks if the current token matches any of the kinds arg
func (ts *TokenSequence) CheckAny(kinds ...TokenKind) bool {
  for _, k := range kinds {
    if ts.Check(k) {
      return true
    }
  }
  return false
}

// Checks if the next token matches any of the kinds arg
func (ts *TokenSequence) PeakAny(kinds ...TokenKind) bool {
  for _, k := range kinds {
    if ts.Peak(k) {
      return true
    }
  }
  return false
}

// Matches the pattern using each slice as a list to match against current token 
// and advancing to the next token if a match was found if no match return false
// unless the list contains a token.ANY then continue without advancing.
// 
// This function should not change the state of TokenSequence
//
// Use token.ANY to make the slice optional
func (ts *TokenSequence) Pattern(p ...[]TokenKind) bool {
  // unwind to where we were before the funciton call
  ts.setUnwindPoint()
  defer ts.unwind()
  for _, s := range p { // for each sub-slice
    matched := false
    // token.ANY should be last because the way this loop works
    for _, k := range sortKindANYLast(s) {
      // match against the current token
      if ts.Check(k) {
        // we found a match call ts.NextToken() and break
        ts.NextToken()
        matched = true
        break
      }
      if k == ANY {
        // continue without calling ts.NextToken()
        matched = true
      }
    }
    // return if the last step didn't match
    if !matched {
      return false
    }
  }

  // success
  return true
}

// reorders a slice of TokenKind to have the ANY token last
// should maintain order of non-ANY tokens
func sortKindANYLast(s []TokenKind) (r []TokenKind) {
  r = make([]TokenKind, len(s))
  hasAny := false
  for _, k := range s {
    if k == ANY {
      hasAny = true
    } else {
      r = append(r, k)
    }
  }
  if hasAny {
    r = append(r, ANY)
  }
  return r
}

// Expects the current token to matches the kind arg
// else it puts a error in GetLastError() by calling Abort()
func (ts *TokenSequence) Expect(kind TokenKind) bool {
  if ts.getCurrentToken().kind != kind {
    ts.Abort("Expected:", kind, "got:", ts.getCurrentToken().kind)
    return false
  }
  ts.NextToken()
  return true
}

// Puts a error in GetLastError() while adding more info to the error string
// very useful for debugging the parser step. use it instead of errors.New()
func (ts *TokenSequence) Abort(msg ...any) {
  line, column := ts.getCurrentToken().line, ts.getCurrentToken().column 
  currentStr := fmt.Sprintf("%v(\"%v\")", ts.getCurrentToken().kind, ts.getCurrentToken().text)
  nextStr := fmt.Sprintf("%v(\"%v\")", ts.getNextToken().kind, ts.getNextToken().text)
  currentStr = strings.ReplaceAll(currentStr, "\n", "\\n")
  nextStr = strings.ReplaceAll(nextStr, "\n", "\\n")
  err_msg := fmt.Sprintf("file.remap:%v:%v:[%v/%v]\n\t%v", line, column, currentStr, nextStr, fmt.Sprintln(msg...))
  fmt.Println(ts.cleanTokenList[0:ts.index])
  // log.Println(err_msg)
  ts.err = errors.New(err_msg)
}

// Used to get the last error produced by Abort()
func (ts *TokenSequence) GetLastError() error {
  return ts.err
}
