package options

import (
  "cmp"
	"errors"
	"time"
)

// here you can se the value of a option 
// int need to be a whole number
// float64 need to be a number + "." + number 
// time.Duration need to be a number followed by "ms"

type Options struct {
  // imaginary mouse edge/limit
  // just like a stick as an edge/limit of -1.0 to 1.0
  StickPoints float64 
  MouseSense float64
  DeadZone float64
  // Store the last X pulls and take them into account
  RollSize int
  // PullRate of the analog per second
  PullRate int
  // how hard it is to move the stick at the center vs at the edge 
  // (higher number = more soft)
  WeightStart float64
  WeightEnd float64
  // in ms how long should a click hold for
  DefaultClickDelay time.Duration
  // the delay between each click in a queue
  DefaultClickQueueDelay time.Duration
}

const (
  STICK_POINTS = 1 + iota
  MOUSE_SENSE
  DEAD_ZONE
  ROLL_SIZE
  PULL_RATE
  WEIGHT_START
  WEIGHT_END
  DEFAULT_CLICK_DELAY
  DEFAULT_CLICK_QUEUE_DELAY
)

// These are the options names

var OPTIONS_MAP map[string]int = map[string]int{
  "StickPoints": STICK_POINTS,
  "MouseSense": MOUSE_SENSE,
  "DeadZone": DEAD_ZONE,
  "RollSize": ROLL_SIZE,
  "PullRate": PULL_RATE,
  "WeightStart": WEIGHT_START,
  "WeightEnd": WEIGHT_END,
  "DefaultClickDelay": DEFAULT_CLICK_DELAY,
  "DefaultClickQueueDelay": DEFAULT_CLICK_QUEUE_DELAY,
}

func Defaults() Options {
  return Options{
    StickPoints: 300,
    MouseSense: 1.0,
    DeadZone: 0.2,
    RollSize: 8,
    PullRate: 1000,
    WeightStart: 1.0,
    WeightEnd: 1.0,
    DefaultClickDelay: 50 * time.Millisecond,
    DefaultClickQueueDelay: 50 * time.Millisecond,
  }
} 

// get index of the option from text for use in Options.Set()
func GetIndexFromText(text string) (int, error) {
  idx, ok := OPTIONS_MAP[text]
  if !ok { return -1, errors.New("This Option Doesn't Exist")}
  return idx, nil
}

// return a Zero value of the type this option is
func GetType(index int) any {
  switch index {
  case STICK_POINTS, MOUSE_SENSE, DEAD_ZONE, WEIGHT_START, WEIGHT_END:
    return float64(0)
  case ROLL_SIZE, PULL_RATE:
    return int(0)
  case DEFAULT_CLICK_DELAY, DEFAULT_CLICK_QUEUE_DELAY:
    return time.Duration(0)
  default:
    return nil  
  }
}

// used to set a option to a value using a index 
func (opt *Options) Set(index int, value any) error {
  switch index {
  case STICK_POINTS:
    iv, ok := value.(float64)
    if !ok { goto ERR }
    opt.StickPoints = iv
  case MOUSE_SENSE:
    iv, ok := value.(float64)
    if !ok { goto ERR }
    opt.MouseSense = iv
  case DEAD_ZONE:
    iv, ok := value.(float64)
    if !ok { goto ERR }
    opt.DeadZone = clamp(iv, 0.0, 1.0)
  case ROLL_SIZE:
    iv, ok := value.(int)
    if !ok { goto ERR }
    opt.RollSize = max(iv, 1)
  case PULL_RATE:
    iv, ok := value.(int)
    if !ok { goto ERR }
    opt.PullRate = max(iv, 1)
  case WEIGHT_START:
    iv, ok := value.(float64)
    if !ok { goto ERR }
    opt.WeightStart = iv
  case WEIGHT_END:
    iv, ok := value.(float64)
    if !ok { goto ERR }
    opt.WeightEnd = iv
  case DEFAULT_CLICK_DELAY:
    iv, ok := value.(time.Duration)
    if !ok { goto ERR }
    opt.DefaultClickDelay = iv
  case DEFAULT_CLICK_QUEUE_DELAY:
    iv, ok := value.(time.Duration)
    if !ok { goto ERR }
    opt.DefaultClickQueueDelay = iv
  default:
    return errors.New("Unknown Option")
  }
  return nil
  ERR:
  return errors.New("Index type and value type mismatch")
}

func clamp[T cmp.Ordered](value T, minium T, maxium T) T {
  return max(minium, min(value, maxium))
}
