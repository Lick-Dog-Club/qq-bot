package text2png

import (
	_ "embed"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

//go:embed AaBuKeYan-2.ttf
var fontBytes []byte

func handleLines(lines []string) (res []string) {
	var newLines []string
	for _, line := range lines {
		newLines = append(newLines, strings.Split(line, "\n")...)
	}
	return newLines
}

func Draw(lines []string, out string) error {
	lines = handleLines(lines)
	f, _ := truetype.Parse(fontBytes)
	face := truetype.NewFace(f, &truetype.Options{
		Size: 26,
	})
	const W = 1024
	var H = 35 * len(lines)
	dc := gg.NewContext(W, H)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	dc.SetFontFace(face)
	const h = 30
	for i, line := range lines {
		y := H/2 - h*len(lines)/2 + i*h
		dc.DrawStringWrapped(line, W, float64(y), 0.9, 0, W, 5, 0)
	}

	return dc.SavePNG(out)
}
