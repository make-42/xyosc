//go:build ignore

//kage:unit pixels

package main

var Strength float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	center := imageDstSize() / 2
	amount := (center - srcPos) * Strength
	var clr vec4
	clr.r = imageSrc2At(srcPos + amount).r
	clr.g = imageSrc2UnsafeAt(srcPos).g
	clr.b = imageSrc2At(srcPos - amount).b
	clr.a = (imageSrc2UnsafeAt(srcPos).a + imageSrc2At(srcPos+amount).a + imageSrc2At(srcPos-amount).a) / 3
	return clr
}
