//go:build ignore

//kage:unit pixels

package main

var Strength float // do not put 0.
var MidPoint float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	clr := imageSrc2UnsafeAt(srcPos)

	clr_l := (clr.r + clr.g + clr.b) / 3
	var clr_out vec4
	if clr_l == 0 {
		clr_out = clr
	} else {
		mult := ((exp((clr.a*clr_l-MidPoint)*Strength) / (1.0 + exp((clr.a*clr_l-MidPoint)*Strength))) - (exp(-MidPoint*Strength) / (1.0 + exp(-MidPoint*Strength)))) / ((exp((1.0-MidPoint)*Strength) / (1.0 + exp((1.0-MidPoint)*Strength))) - (exp(-MidPoint*Strength) / (1.0 + exp(-MidPoint*Strength))))
		clr_out.a = clr.a
		clr_out.r = clr.r * mult / clr_l
		clr_out.g = clr.g * mult / clr_l
		clr_out.b = clr.b * mult / clr_l
	}

	return clr_out
}
