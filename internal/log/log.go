package log

import (
	"fmt"
	"log"
	"os"

)

const (
  colorReset = "\033[0m"
  colorRed = "\033[31m"
  colorGreen = "\033[32m"
  colorYellow = "\033[33m"
  colorBlue = "\033[34m"
  colorPurple = "\033[35m"
  colorCyan = "\033[36m"
  colorWhite = "\033[37m"
)

var logger_normal *log.Logger = nil
var logger_info *log.Logger = nil
var logger_warn *log.Logger = nil
var logger_err *log.Logger = nil
var logger_fatal *log.Logger = nil
var debug = true

func init() {
  logger_normal = log.New(os.Stderr, "", 0)
  logger_info = log.New(os.Stderr, colorCyan+"[INFO] "+colorReset, log.Ltime | log.Lshortfile)
  logger_warn = log.New(os.Stderr, colorYellow+"[WARN] "+colorReset, log.Ltime | log.Lshortfile)
  logger_err = log.New(os.Stderr, colorRed+"[ERR] "+colorReset, log.Ltime | log.Lshortfile)
  logger_fatal = log.New(os.Stderr, colorPurple+"[FATAL] "+colorReset, log.Ltime | log.Lshortfile)
}


func logging(l *log.Logger, calldepth int, v ...any) {
  l.Output(calldepth, fmt.Sprintln(v...))
}

func Normal(v ...any) {
  if debug {
    logging(logger_normal, 3, v...)
  }
}

func Info(v ...any) {
  if debug {
    logging(logger_info, 3, v...)
  }
}

func Warn(v ...any) {
  if debug {
    logging(logger_warn, 3, v...)
  }
}

func Error(v ...any) {
  if debug {
    logging(logger_err, 3, v...)
  }
}

func Fatal(v ...any) {
  logging(logger_fatal, 3, v...)
  os.Exit(1)
}

