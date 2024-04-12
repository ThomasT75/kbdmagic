package remap

import (
	"kbdmagic/common"
)

func GetRemapTable(s string) (common.RemapTableType, *[]error) {
  tl, err_list := lexer(s)
  if err_list != nil {
    return common.RemapTableType{}, err_list
  }

  ts := NewTokenSequence(tl)
  emi := NewEmitter()

  err := parser(ts, emi)
  if err != nil {
    return common.RemapTableType{}, &[]error{err}
  }
  // log.Println(Emitter.output)
  return emi.output, emi.GetErrorList()
}
