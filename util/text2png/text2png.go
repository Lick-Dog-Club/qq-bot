package text2png

import (
	_ "embed"
	"math"
	"strings"
	"unicode"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

//go:embed FangZhengShuSongJianTi-1.ttf
var FontBytes []byte

func handleLines(lines []string) (res []string) {
	var newLines []string
	for _, line := range lines {
		newLines = append(newLines, strings.Split(line, "\n")...)
	}
	return newLines
}

// CharCount 汉字一个约等于 2 个字母，返回的是字母数量
func CharCount(s string) int {
	var count int
	for _, c := range s {
		if unicode.Is(unicode.Han, c) {
			count += 2
		} else {
			count += 1
		}
	}
	return count
}

func Draw(lines []string, out string) error {
	lines = handleLines(lines)
	var max float64
	const fontSize = 28

	for _, ll := range lines {
		max = math.Max(max, float64(CharCount(ll)))
	}
	f, _ := truetype.Parse(FontBytes)
	face := truetype.NewFace(f, &truetype.Options{
		Size: fontSize,
	})
	var W float64 = float64(max) * fontSize * 0.6
	var H = 40 * len(lines)
	dc := gg.NewContext(int(W), H)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	dc.SetFontFace(face)
	const h = 38
	for i, line := range lines {
		y := H/2 - h*len(lines)/2 + i*h
		dc.DrawStringWrapped(line, W, float64(y), 0.95, 0, W, 1.5, gg.AlignLeft)
	}

	return dc.SavePNG(out)
}
