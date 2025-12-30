// Package oklab implements the Oklab color space, as described at https://bottosson.github.io/posts/oklab/
package oklab

// L: 0.000000–1.000000
// A: -0.233888–0.276216
// B: -0.311528–0.198570

// L: 0.000000–1.000000
// C: 0.000000–0.322491
// H: -3.141592–3.141592

import (
	"image/color"
	"math"
)

type Oklab struct {
	L float64 // Perceived lightness
	A float64 // How green/red the color is
	B float64 // How blue/yellow the color is
}

type Oklch struct {
	L float64 // Perceived lightness
	C float64 // Chroma
	H float64 // Hue
}

var OklabModel = color.ModelFunc(oklabModel)
var OklchModel = color.ModelFunc(oklchModel)

// See image.Color.
func (c Oklab) RGBA() (uint32, uint32, uint32, uint32) {
	r, g, b := c.SRGB()
	r, g, b = clampf(r), clampf(g), clampf(b)
	return uint32(0xffff * r), uint32(0xffff * g), uint32(0xffff * b), 0xffff
}

// Convert to linear sRGB.
// See https://bottosson.github.io/posts/oklab/
func (c Oklab) LinearSRGB() (float64, float64, float64) {
	l_ := c.L + 0.3963377774*c.A + 0.2158037573*c.B
	m_ := c.L - 0.1055613458*c.A - 0.0638541728*c.B
	s_ := c.L - 0.0894841775*c.A - 1.2914855480*c.B

	l := l_ * l_ * l_
	m := m_ * m_ * m_
	s := s_ * s_ * s_

	r := 4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	g := -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	b := -0.0041960863*l - 0.7034186147*m + 1.7076147010*s

	return r, g, b
}

// Convert to sRGB.
// See https://bottosson.github.io/posts/colorwrong/#what-can-we-do%3F
func (c Oklab) SRGB() (float64, float64, float64) {
	r, g, b := c.LinearSRGB()
	return linearSrgbToSrgb(r), linearSrgbToSrgb(g), linearSrgbToSrgb(b)
}

// Convert to LCh, which is Oklab in polar.
func (c Oklab) Oklch() Oklch {
	return Oklch{
		L: c.L,
		C: math.Sqrt(c.A*c.A + c.B*c.B),
		H: math.Atan2(c.B, c.A),
	}
}

// See image.Color.
func (c Oklch) RGBA() (uint32, uint32, uint32, uint32) {
	return c.Oklab().RGBA()
}

// Convert to Oklab.
func (c Oklch) Oklab() Oklab {
	return Oklab{
		L: c.L,
		A: c.C * math.Cos(c.H),
		B: c.C * math.Sin(c.H),
	}
}

func oklabModel(c color.Color) color.Color {
	r8, g8, b8, a8 := c.RGBA()
	r := float64(r8) / float64(a8)
	g := float64(g8) / float64(a8)
	b := float64(b8) / float64(a8)

	r, g, b = srgbToLinearSrgb(r), srgbToLinearSrgb(g), srgbToLinearSrgb(b)

	l := 0.4122214708*r + 0.5363325363*g + 0.0514459929*b
	m := 0.2119034982*r + 0.6806995451*g + 0.1073969566*b
	s := 0.0883024619*r + 0.2817188376*g + 0.6299787005*b

	l_ := math.Cbrt(l)
	m_ := math.Cbrt(m)
	s_ := math.Cbrt(s)

	return Oklab{
		L: 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_,
		A: 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_,
		B: 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_,
	}
}

func oklchModel(c color.Color) color.Color {
	return oklabModel(c).(Oklab).Oklch()
}

// Convert a linear sRGB color component to sRGB.
// See https://bottosson.github.io/posts/colorwrong/#what-can-we-do%3F
func linearSrgbToSrgb(x float64) float64 {
	if x >= 0.0031308 {
		return 1.055*math.Pow(x, 1.0/2.4) - 0.055
	}
	return 12.92 * x
}

// Convert an sRGB color component to linear sRGB.
// See https://bottosson.github.io/posts/colorwrong/#what-can-we-do%3F
func srgbToLinearSrgb(x float64) float64 {
	if x >= 0.04045 {
		return math.Pow((x+0.055)/(1+0.055), 2.4)
	}
	return x / 12.92
}

func clampf(x float64) float64 {
	if x < 0 {
		return 0
	} else if x > 1 {
		return 1
	}
	return x
}
