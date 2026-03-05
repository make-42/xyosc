package i3

type Vec struct {
	X int
	Y int
	Z int
}

func (a Vec) Add(b Vec) Vec { return Vec{X: a.X + b.X, Y: a.Y + b.Y, Z: a.Z + b.Z} }
func (a Vec) Sub(v Vec) Vec { return Vec{X: a.X - v.X, Y: a.Y - v.Y, Z: a.Z - v.Z} }

func (a Vec) AddScalar(v int) Vec { return Vec{X: a.X + v, Y: a.Y + v, Z: a.Z + v} }
func (a Vec) MulScalar(v int) Vec { return Vec{X: a.X * v, Y: a.Y * v, Z: a.Z * v} }
func (a Vec) DivScalar(v int) Vec { return Vec{X: a.X / v, Y: a.Y / v, Z: a.Z / v} }

func (a Vec) ShiftRight(lo int) Vec { return Vec{X: a.X >> lo, Y: a.Y >> lo, Z: a.Z >> lo} }
func (a Vec) ShiftLeft(hi int) Vec  { return Vec{X: a.X << hi, Y: a.Y << hi, Z: a.Z << hi} }

func (a Vec) AndScalar(b int) Vec    { return Vec{X: a.X & b, Y: a.Y & b, Z: a.Z & b} }
func (a Vec) OrScalar(b int) Vec     { return Vec{X: a.X | b, Y: a.Y | b, Z: a.Z | b} }
func (a Vec) XorScalar(b int) Vec    { return Vec{X: a.X ^ b, Y: a.Y ^ b, Z: a.Z ^ b} }
func (a Vec) AndnotScalar(b int) Vec { return Vec{X: a.X &^ b, Y: a.Y &^ b, Z: a.Z &^ b} }
