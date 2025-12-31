//go:build ignore

//kage:unit pixels

package main

var Strength float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	clr := imageSrc2UnsafeAt(srcPos)

	samples := [...]float{
		0, 1, 2, 3, 4, 5, 6, 7, 8,
	}
	weights := [...]float{
		0.398942280401, 0.386668116803, 0.352065326764, 0.301137432155, 0.241970724519, 0.182649085389, 0.129517595666, 0.0862773188265, 0.0539909665132,
	}
	sum := clr
	for bx := 0; bx < 3; bx++ {
		for by := 0; by < 3; by++ {
			for i := 0; i < len(samples); i++ {
				sum += imageSrc2At(srcPos+vec2(float(bx-1), float(by-1))*samples[i]) * weights[i]
			}
		}
	}
	sum = (1-Strength)*clr + Strength*sum

	return sum
}
