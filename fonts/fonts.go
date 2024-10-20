package fonts

import (
	"unicode"
	"xyosc/utils"

	"github.com/flopp/go-findfont"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func getAllChars(table *unicode.RangeTable) []rune {
	res := make([]rune, 0)

	for _, r := range table.R16 {
		for c := r.Lo; c <= r.Hi; c += r.Stride {
			res = append(res, rune(c))
		}
	}
	for _, r := range table.R32 {
		for c := r.Lo; c <= r.Hi; c += r.Stride {
			res = append(res, rune(c))
		}
	}

	return res
}

var FontIosevka32 rl.Font
var FontIosevka16 rl.Font

func Init() {
	// NOTE: Textures/Fonts MUST be loaded after Window initialization (OpenGL context is required)
	fontPath, err := findfont.Find("SourceHanSansJP-Heavy.otf")
	utils.CheckError(err)
	runes := make([]rune, 0)
	runes = append(runes, getAllChars(unicode.Katakana)...)
	runes = append(runes, getAllChars(unicode.Hiragana)...)
	runes = append(runes, getAllChars(unicode.Latin)...)
	runes = append(runes, getAllChars(unicode.Digit)...)
	runes = append(runes, getAllChars(unicode.Punct)...)
	runes = append(runes, getAllChars(unicode.Han)...)
	runes = append(runes, getAllChars(unicode.Symbol)...)
	FontIosevka32 = rl.LoadFontEx(fontPath, 32, runes)
	rl.GenTextureMipmaps(&FontIosevka32.Texture)
	rl.SetTextureFilter(FontIosevka32.Texture, rl.FilterPoint)
	FontIosevka16 = rl.LoadFontEx(fontPath, 16, runes)
	rl.GenTextureMipmaps(&FontIosevka16.Texture)
	rl.SetTextureFilter(FontIosevka16.Texture, rl.FilterPoint)
}
