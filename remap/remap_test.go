package remap

import (
	"bytes"
	"log"
	"os"
	"testing"
)

var sc = `

//KINDA HATE THIS AUTO TAB INSIDE HERE


KEY_LEFTCTRL {
  OFF:
    REL_WHEEL+ [hold GP_BTN_LB, GP_BTN_X] 
    REL_WHEEL- [hold GP_BTN_LB, GP_BTN_B]
  ON:
    REL_WHEEL+ [hold GP_BTN_LB, GP_BTN_Y] 
    REL_WHEEL- [hold GP_BTN_LB, GP_BTN_A]
}

KEY_LEFTCTRL{
OFF
REL_WHEEL+[hold GP_BTN_LB,GP_BTN_X] 
REL_WHEEL-[hold GP_BTN_LB,
GP_BTN_B]
ON
REL_WHEEL+[hold GP_BTN_LB,GP_BTN_Y] 
REL_WHEEL-[hold GP_BTN_LB,GP_BTN_A]
}

KEY_W:GP_AXIS_X-1.0
KEY_S:GP_AXIS_X+1
REL_X:GP_AXIS_RX

`

func Test(t *testing.T) {
  var buf bytes.Buffer
  log.SetOutput(&buf)
  defer func() {
      log.SetOutput(os.Stderr)
  }()
  getRemapTable(sc)
  t.Log(buf.String())
}
