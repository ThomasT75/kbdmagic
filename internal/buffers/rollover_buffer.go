package buffers

type RolloverBuf[T any] struct {
  buf []T
  idx int 
  size int
}

func (b *RolloverBuf[_]) Size() int {
  return b.size
}

func NewRolloverBuf[T any](size int) RolloverBuf[T] {
  return RolloverBuf[T]{
    buf: make([]T, size),
    idx: 0,
    size: size,
  }
}

func (b *RolloverBuf[T]) Add(v T) {
  b.buf[b.idx] = v 
  b.idx += 1
  b.idx %= b.size
}

func (b *RolloverBuf[T]) Sum(f func(sum T, value T) T) (s T) {
  for _, v := range b.buf {
    s = f(s, v)
  }
  return s
}
