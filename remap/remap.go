package remap

import (
	"errors"
	"kbdmagic/common"
	"kbdmagic/remap/emitter"
	"kbdmagic/remap/lexer"
	"kbdmagic/remap/parser"
	"kbdmagic/remap/validator"
)

// compiles source into a remapTableList
func GetRemapTable(source string) (common.RemapTableListType, []error) {
  // lexer steo: separate words and into tokens
  ts, lexerErrList := lexer.Lexer(source)
  if lexerErrList != nil {
    stepErr := errors.New("Lexer Step Failed")
    return common.RemapTableListType{}, append([]error{stepErr}, lexerErrList...)
  }

  // make an emitter
  emi := emitter.NewEmitter()

  // parse the token sequence and emit them using emitter
  parserErr := parser.Parser(&ts, &emi)
  if parserErr != nil {
    stepErr := errors.New("Parser Step Failed")
    return common.RemapTableListType{}, append([]error{stepErr}, parserErr)
  }

  // check emitter for errors
  emitterErrList := emi.GetErrorList()
  if emitterErrList != nil {
    stepErr := errors.New("Emitter Step Failed")
    return common.RemapTableListType{}, append([]error{stepErr}, emitterErrList...)
  }

  // get emitter output
  eos, opts := emi.GetOutput()

  // very nice way of debugging the emitter but i don't have loging to file setup yet
  // and this takes a lot of terminal screen space
  // var ds string = "\n"
  // for _, e := range eos {
  //   ds += fmt.Sprintf("%+v\n", e)
  // }
  // log.Fatal(ds, emi.GetErrorList())

  // validate the emitter output (and construct into a remapTableList)
  remapTableList, validatorErrList := validator.Validate(eos, opts)
  if validatorErrList != nil {
    stepErr := errors.New("validation Step Failed")
    return common.RemapTableListType{}, append([]error{stepErr}, validatorErrList...)
  }

  // pack it, zip it and ship it 
  return remapTableList, nil
}
