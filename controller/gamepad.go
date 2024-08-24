package controller

import (
	"github.com/ThomasT75/uinput"
)

type Gamepad struct {
  gamepad uinput.Gamepad
}

// only used as OS translation layer thing
// meaning no advanced logic or exposing struct fields
// TODO check for uinput in multiple places
func NewGamepad(Name string) (*Gamepad, error) {
  controller, err := uinput.CreateGamepad("/dev/uinput", []byte(Name), 0xDEAD, 0xBEEF)
  if err != nil {
    return nil, err
  }
  return &Gamepad{
    gamepad: controller,
  }, nil
}

func (g *Gamepad) ButtonDown(key int) error {
  return g.gamepad.ButtonDown(key)
}

func (g *Gamepad) ButtonUp(key int) error {
  return g.gamepad.ButtonUp(key)
}

func (g *Gamepad) LeftStickMove(x, y float32) error {
  return g.gamepad.LeftStickMove(x, y)
}

func (g *Gamepad) RightStickMove(x, y float32) error {
  return g.gamepad.RightStickMove(x, y)
}

func (g *Gamepad) LeftTriggerForce(value float32) error {
  return g.gamepad.LeftTriggerForce(value)
}

func (g *Gamepad) RightTriggerForce(value float32) error {
  return g.gamepad.RightTriggerForce(value)
}

func (g *Gamepad) Close() error {
  return g.gamepad.Close()
}
