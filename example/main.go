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

// +build example

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/pkg/browser"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/hajimehoshi/bitmapfont"
)

var (
	flagTest = flag.Bool("test", false, "test mode")
)

func run() error {
	width := 640

	// https://www.unicode.org/udhr/
	// https://omniglot.com/udhr/
	text := `en:  All human beings are born free and equal in dignity and rights.
ang: Ealle fīras sind boren frēo ond geefenlican in ār ond riht.
de:  Alle Menschen sind frei und gleich an Würde und Rechten geboren.
el:  'Ολοι οι άνθρωποι γεννιούνται ελεύθεροι και ίσοι στην αξιοπρέπεια και τα δικαιώματα.
es:  Todos los seres humanos nacen libres e iguales en dignidad y derechos y dotados.
eo:  Ĉiuj homoj estas denaske liberaj kaj egalaj laŭ digno kaj rajtoj.
fr:  Tous les êtres humains naissent libres et égaux en dignité et en droits.
got: ᚨᛚᛚᚨᛁ ᛗᚨᚾᚾᚨ ᚠᚱᛖᛁᚺᚨᛚᛋ ᛃᚨᚺ ᛋᚨᛗᚨᛚᛖᛁᚲᛟ ᛁᚾ ᚹᚨᛁᚱᚦᛁᛞᚨᛁ ᛃᚨᚺ ᚱᚨᛁᚺᛏᛖᛁᛋ ᚹᚨᚢᚱᚦᚨᚾᛋ.
hy:  Բոլոր մարդիկ ծնվում են ազատ ու հավասար՝ իրենց արժանապատվությամբ և իրավունքներով:
it:  Tutti gli esseri umani nascono liberi ed eguali in dignità e diritti.
ja:  すべての人間は、生れながらにして自由であり、かつ、尊厳と権利とについて平等である。
ka:  ყველა ადამიანი იბადება თავისუფალი და თანასწორი თავისი ღირსებითა და უფლებებით.
ko:  모든 인간은 태어날 때부터 자유로우며 그 존엄과 권리에 있어 동등하다.
mn:  Хүн бүр төрж мэндлэхэд эрх чөлөөтэй, адилхан нэр төртэй, ижил эрхтэй байдаг.
pl:  Wszyscy ludzie rodzą się wolni i równi pod względem swej godności i swych praw.
pt:  Todos os seres humanos nascem livres e iguais em dignidade e em direitos.
ru:  Все люди рождаются свободными и равными в своем достоинстве и правах.
sw:  Watu wote wamezaliwa huru, hadhi na haki zao ni sawa.
tr:  Bütün insanlar hür, haysiyet ve haklar bakımından eşit doğarlar.
uk:  Всі люди народжуються вільними і рівними у своїй гідності та правах.
vi:  Tất cả mọi người sinh ra đều được tự do và bình đẳng về nhân phẩm và quyền.
`

	const (
		offsetX = 8
		offsetY = 8
	)

	for _, s := range []int{10, 12} {
		var (
			dotX        int
			dotY        int
			glyphWidth  int
			glyphHeight int
		)
		switch s {
		case 10:
			dotX = 3
			dotY = 9
			glyphWidth = 10
			glyphHeight = 12
		case 12:
			dotX = 4
			dotY = 12
			glyphWidth = 12
			glyphHeight = 16
		}

		height := glyphHeight*len(strings.Split(strings.TrimSpace(text), "\n")) + offsetX*2
		if *flagTest {
			width = glyphWidth*256 + offsetX*2
			height = glyphHeight*256 + offsetY*2
			text = ""
			for i := 0; i < 256; i++ {
				for j := 0; j < 256; j++ {
					r := rune(i*256 + j)
					if r == '\n' {
						text += " "
						continue
					}
					text += string(r)
				}
				text += "\n"
			}
		}

		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.ZP, draw.Src)
		if *flagTest {
			gray := color.RGBA{0xcc, 0xcc, 0xcc, 0xff}
			for j := 0; j < 256; j++ {
				for i := 0; i < 256; i++ {
					if (i+j)%2 == 0 {
						continue
					}
					x := i*glyphWidth + offsetX
					y := j*glyphHeight + offsetY
					draw.Draw(dst, image.Rect(x, y, x+glyphWidth, y+glyphHeight), image.NewUniform(gray), image.ZP, draw.Src)
				}
			}
		}

		var f font.Face
		switch s {
		case 10:
			f = bitmapfont.Gothic10r
		case 12:
			f = bitmapfont.Gothic12r
		}
		d := font.Drawer{
			Dst:  dst,
			Src:  image.NewUniform(color.Black),
			Face: f,
			Dot:  fixed.P(dotX+offsetX, dotY+offsetY),
		}

		for _, l := range strings.Split(text, "\n") {
			d.DrawString(l)
			d.Dot.X = fixed.I(dotX + offsetX)
			d.Dot.Y += f.Metrics().Height
		}

		path := fmt.Sprintf("example_%d.png", s)
		if *flagTest {
			path = fmt.Sprintf("example_%d_test.png", s)
		}
		fout, err := os.Create(path)
		if err != nil {
			return err
		}
		defer fout.Close()

		if err := png.Encode(fout, d.Dst); err != nil {
			return err
		}

		if err := browser.OpenFile(path); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		panic(err)
	}
}
