//go:build ignore

//kage:unit pixels

package main

var Strength float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	center := imageDstSize() / 2
	amount := (center - srcPos) * Strength
	var clr vec4
	clr.r = imageSrc2At(srcPos+amount).r*3/4 + (imageSrc2At(srcPos+amount/2).r+imageSrc2At(srcPos+amount/2).g)/4
	clr.g = imageSrc2UnsafeAt(srcPos).g/2 + (imageSrc2At(srcPos+amount/2).r+imageSrc2At(srcPos+amount/2).g)/4 + (imageSrc2At(srcPos-amount/2).g+imageSrc2At(srcPos-amount/2).b)/4
	clr.b = imageSrc2At(srcPos-amount).b*3/4 + (imageSrc2At(srcPos-amount/2).g+imageSrc2At(srcPos-amount/2).b)/4
	clr.a = (imageSrc2UnsafeAt(srcPos).a + imageSrc2At(srcPos+amount).a + imageSrc2At(srcPos-amount).a + imageSrc2At(srcPos+amount/2).a + imageSrc2At(srcPos-amount/2).a) / 5
	return clr
}
