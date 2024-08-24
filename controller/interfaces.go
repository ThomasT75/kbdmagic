package controller

import (
	"io"
)

type GamepadInterface interface {
	// ButtonDown will send a button-press event to an existing gamepad device.
	// The key can be any of the predefined keycodes from keycodes.go.
	ButtonDown(key int) error

	// ButtonUp will send a button-release event to an existing gamepad device.
	// The key can be any of the predefined keycodes from keycodes.go.
	ButtonUp(key int) error

	// LeftStickMove moves the left stick along the x and y-axis
	LeftStickMove(x, y float32) error
	// RightStickMove moves the right stick along the x and y-axis
	RightStickMove(x, y float32) error

  // LeftTriggerForce performs a trigger-axis-z event with a given force
  LeftTriggerForce(value float32) error
  // RightTriggerForce performs a trigger-axis-rz event with a given force
  RightTriggerForce(value float32) error

	io.Closer
}

