package token

import "fmt"

// separated from token.go for readability

type TokenKind int

const (
  ANY TokenKind = -256
  EOF TokenKind = -2
	NOT_KIND TokenKind = -1 
	KEY TokenKind = 1 + iota
	REL
	BTN

	GP_BTN
	GP_AXIS

  INT
	FLOAT
  STRING

	DELAY_MS

	COMMENT
	NL

  COMMA

  OFF
  ON

  COLON
  EXCLAMATION

  // []
  LEFT_BRACKET
  RIGHT_BRACKET

  // {}
  LEFT_BRACES
  RIGHT_BRACES

  // ()
  LEFT_PARENTHESIS
  RIGHT_PARENTHESIS

  RELEASE
  HOLD

  TOGGLE

  PLUS
  MINUS
  EQUAL

  OPTION_NAME

  SKIP
  WAIT

  NONE
)

var _map_token_kind = map[TokenKind]string{
  EOF: "EOF",
  NOT_KIND: "NOT_KIND",
  KEY: "KEY",
	REL: "REL",
	BTN: "BTN",

	GP_BTN: "GP_BTN",
	GP_AXIS: "GP_AXIS",

  INT: "INT",
	FLOAT: "FLOAT",
  STRING: "STRING",

	DELAY_MS: "DELAY_MS",

	COMMENT: "COMMENT",
	NL: "NL",
  COMMA: "COMMA",

  OFF: "OFF",
  ON: "ON",

  COLON: "COLON",
  EXCLAMATION: "EXCLAMATION",
  
  LEFT_BRACKET: "LEFT_BRACKET",
  RIGHT_BRACKET: "RIGHT_BRACKET",
  LEFT_BRACES: "LEFT_BRACES",
  RIGHT_BRACES: "RIGHT_BRACES",
  LEFT_PARENTHESIS: "LEFT_PARENTHESIS",
  RIGHT_PARENTHESIS: "RIGHT_PARENTHESIS",

  RELEASE: "RELEASE",
  HOLD: "HOLD",

  TOGGLE: "TOGGLE",

  PLUS: "PLUS",
  MINUS: "MINUS",
  EQUAL: "EQUAL",

  OPTION_NAME: "OPTION_NAME",

  SKIP: "SKIP",
  WAIT: "WAIT",

  NONE: "NONE",
}

func (k TokenKind) String() string {
  s, ok := _map_token_kind[k]
  if !ok {
    return fmt.Sprintf("TokenKind(%v)", int(k))
  }
  return s
}
