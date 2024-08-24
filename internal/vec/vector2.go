package vec

import "math"


//x is first y is second
type Vector2 struct {
  X float64
  Y float64
}

// how oposite are the 2 vectors
// -1.0 = oposite | 1.0 = not oposite
func Dot(a Vector2, b Vector2) float64 {
  return (a.X * b.X) + (a.Y * b.Y)
}

// magnitude without the square root step
// aka lenght²
func MagnitudeSquared(a Vector2) float64 {
  return a.X * a.X + a.Y * a.Y
}

// aka length
func Magnitude(a Vector2) float64 {
  m := MagnitudeSquared(a)
  return math.Sqrt(m)
}

// put the vector at unit length regardless 
func Normalize(a Vector2) (r Vector2) {
  m := Magnitude(a)
  r.X = a.X / m
  r.Y = a.Y / m
  return r
}

// put the vector at unit length if past the unit length
func NormalizePastUnit(a Vector2) (r Vector2) {
  m := MagnitudeSquared(a)
  if m < 1.0 {
    return a
  }
  m = math.Sqrt(m)
  r.X = a.X / m
  r.Y = a.Y / m
  return r
}

func Sum(a Vector2, b Vector2) (r Vector2) {
  r.X = a.X + b.X
  r.Y = a.Y + b.Y
  return r
}

func DivideF(a Vector2, f float64) (r Vector2) {
  r.X = a.X / f
  r.Y = a.Y / f
  return r
}

func MultiplyF(a Vector2, f float64) (r Vector2) {
  r.X = a.X * f
  r.Y = a.Y * f
  return r
}

func (v *Vector2) Div(f float64) (r Vector2) {
  return DivideF(*v, f)
}

func (v *Vector2) Mul(f float64) (r Vector2) {
  return MultiplyF(*v, f)
}

func (v *Vector2) IsZero() bool {
  return v.X == 0.0 && v.Y == 0.0 
}

func (v *Vector2) Unpack32() (x float32, y float32) {
  return float32(v.X), float32(v.Y)
}

