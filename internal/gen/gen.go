// Copyright 2018 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build generate
// +build generate

package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"

	"golang.org/x/text/width"

	"github.com/interfacekun/bitmapfont/v3/internal/arabic"
	"github.com/interfacekun/bitmapfont/v3/internal/baekmuk"
	"github.com/interfacekun/bitmapfont/v3/internal/fixed"
	"github.com/interfacekun/bitmapfont/v3/internal/mplus"
	"github.com/interfacekun/bitmapfont/v3/internal/unicode"
	"github.com/interfacekun/bitmapfont/v3/internal/wqi"
)

var (
	flagWidths   = flag.Bool("widths", false, "output widths infomation")
	flagOutput   = flag.String("output", "", "output file")
	flagEastAsia = flag.Bool("eastasia", false, "prefer east Asia punctuations")
	flagChinese  = flag.Bool("chinese", false, "prefer Chiinese glyphs")
)

const (
	glyphRegionWidth  = 12
	glyphRegionHeight = 16
)

type fontType int

const (
	fontTypeNone fontType = iota
	fontTypeFixed
	fontTypeMPlus
	fontTypeBaekmuk
	fontTypeArabic
	fontTypeWqi
)

func getFontType(r rune) fontType {
	// For Latin glyphs, M+ doesn't work. Use the fixed font whatever the face is.
	if unicode.IsLatin(r) {
		return fontTypeFixed
	}

	if 0x2500 <= r && r <= 0x257f {
		// Box Drawing
		// M+ defines a part of box drawing glyphs.
		// For consistency, use other font's glyphs instead.
		return fontTypeBaekmuk
	}
	if 0xff65 <= r && r <= 0xff9f {
		// Halfwidth Katakana
		return fontTypeMPlus
	}

	if width.LookupRune(r).Kind() == width.EastAsianAmbiguous {
		if *flagEastAsia {
			return fontTypeMPlus
		}
		return fontTypeFixed
	}

	if _, ok := fixed.Glyph(r, 12); ok {
		return fontTypeFixed
	}
	if *flagChinese {
		if _, ok := wqi.Glyph(r); ok {
			return fontTypeWqi
		}
	}
	if _, ok := mplus.Glyph(r, 12); ok {
		return fontTypeMPlus
	}
	if _, ok := baekmuk.Glyph(r, 12); ok {
		return fontTypeBaekmuk
	}
	if _, ok := arabic.Glyph(r); ok {
		return fontTypeArabic
	}
	return fontTypeNone
}

func getGlyph(r rune) (image.Image, bool) {
	switch getFontType(r) {
	case fontTypeNone:
		return nil, false
	case fontTypeFixed:
		g, ok := fixed.Glyph(r, 12)
		if ok {
			return &g, true
		}
	case fontTypeMPlus:
		g, ok := mplus.Glyph(r, 12)
		if ok {
			return &g, true
		}
	case fontTypeBaekmuk:
		g, ok := baekmuk.Glyph(r, 12)
		if ok {
			return &g, true
		}
	case fontTypeArabic:
		g, ok := arabic.Glyph(r)
		if ok {
			return g, true
		}
	case fontTypeWqi:
		g, ok := wqi.Glyph(r)
		if ok {
			return g, true
		}
	default:
		panic("not reached")
	}
	return nil, false
}

func addGlyphs(img draw.Image) {
	for j := 0; j < 0x100; j++ {
		for i := 0; i < 0x100; i++ {
			r := rune(i + j*0x100)
			g, ok := getGlyph(r)
			if !ok {
				continue
			}

			b := g.Bounds()
			w, h := b.Dx(), b.Dy()
			dstX := i * glyphRegionWidth
			dstY := j * glyphRegionHeight
			dstR := image.Rect(dstX, dstY, dstX+w, dstY+h)
			p := g.Bounds().Min
			draw.Draw(img, dstR, g, p, draw.Over)
		}
	}
}

func run() error {
	if *flagWidths {
		return outputWidths()
	}

	img := image.NewAlpha(image.Rect(0, 0, glyphRegionWidth*256, glyphRegionHeight*256))
	addGlyphs(img)

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	as := make([]byte, w*h/8)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			a := img.At(i, j).(color.Alpha).A
			idx := w*j + i
			if a != 0 {
				as[idx/8] |= 1 << uint(7-idx%8)
			}
		}
	}

	fout, err := os.Create(*flagOutput)
	if err != nil {
		return err
	}
	defer fout.Close()

	cw, err := gzip.NewWriterLevel(fout, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer cw.Close()

	if _, err := cw.Write(as); err != nil {
		return err
	}
	return nil
}

func outputWidths() error {
	var wideRunes []rune
	for r := rune(0); r <= 0xffff; r++ {
		img, ok := arabic.Glyph(r)
		if !ok {
			continue
		}
		if img.Bounds().Dx() == glyphRegionWidth {
			wideRunes = append(wideRunes, r)
		}
	}

	f, err := os.Create(*flagOutput)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "// Code generated by github.com/interfacekun/bitmapfont/internal/gen. DO NOT EDIT.")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "package bitmap")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "var wideRunes = map[rune]struct{}{")
	for _, r := range wideRunes {
		fmt.Fprintf(f, "\t0x%04x: {},\n", r)
	}
	fmt.Fprintln(f, "}")

	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		panic(err)
	}
}
