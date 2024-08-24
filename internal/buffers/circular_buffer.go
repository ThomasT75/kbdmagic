package buffers

type CircularBuf[T any] struct {
  buf []T
  readIndex int
  writeIndex int
  size int
}

func NewCircularBuf[T any](size int) CircularBuf[T] {
  return CircularBuf[T]{
    buf: make([]T, size),
    size: size,
  }
}

func (c *CircularBuf[T]) Write(cmd T) {
  c.buf[c.writeIndex % c.size] = cmd
  c.writeIndex += 1 
}

func (c *CircularBuf[T]) Read() T {
  r := c.buf[c.readIndex % c.size]
  if c.readIndex < c.writeIndex {
    c.readIndex += 1 
  } 
  return r
}

func (c *CircularBuf[_]) CanRead() bool {
  return c.readIndex < c.writeIndex
}

func (c *CircularBuf[_]) SpaceLeftToWrite() int {
  return c.size - (c.writeIndex - c.readIndex)
}
