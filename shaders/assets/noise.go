//go:build ignore

//kage:unit pixels

package main

var Strength float
var Time float
var Scale float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	var clr vec4
	clr = imageSrc2At(srcPos)
	amount := abs(cos(sin(srcPos.x*Scale+Time+cos(srcPos.y*Scale+Time)*Scale)*Scale+sin(srcPos.x*Scale+Time)*Scale)) * Strength
	clr.r += amount
	clr.g += amount
	clr.b += amount
	clr.a += amount
	return clr
}
