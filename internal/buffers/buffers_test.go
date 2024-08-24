package buffers_test

import (
	"kbdmagic/internal/buffers"
	"testing"
)

func TestCircularBufferCanRead(t *testing.T) {
  bc := buffers.CircularBuf[int]{}
  if bc.CanRead() {
    t.Error("CircularBuf shouldn't be readable on creation")
  }
  bc.Write(0)
  if !bc.CanRead() {
    t.Error("CircularBuf should be readable after write")
  }
  bc.Read()
  if bc.CanRead() {
    t.Error("CircularBuf shouldn't be readable after read")
  }
}
