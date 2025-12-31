// A Kage port of https://www.shadertoy.com/view/Ms23DR
//
// The original license comment is:
//	Loosely based on postprocessing shader by inigo quilez,
//	License Creative Commons Attribution-NonCommercial-ShareAlike 3.0 Unported License.
//	https://creativecommons.org/licenses/by-nc-sa/3.0/deed.en
// Credit: https://github.com/Zyko0/kage-shaders/blob/main/crt/mattias/crt.kage

//go:build ignore

package main

//kage:unit pixels
var Strength float // do not put 0.

func curve(uv vec2) vec2 {
	uv = (uv - 0.5) * 2
	uv *= 1.1
	uv.x *= (1 + pow((abs(uv.y)/5), 2/Strength))
	uv.y *= (1 + pow((abs(uv.x)/4), 2/Strength))
	uv = uv/2 + 0.5
	uv = uv*0.92 + 0.04

	return uv
}

func Fragment(dst vec4, src vec2, color vec4) vec4 {
	origin, size := imageSrcRegionOnTexture()
	q := (src - origin) / size
	uv := q
	uv = curve(uv)

	return imageSrc0At(vec2(uv.x, uv.y)*size + origin)
}
