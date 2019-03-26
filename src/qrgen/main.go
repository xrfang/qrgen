package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func genCode(content string, level qrcode.RecoveryLevel, size int) image.Image {
	buf, err := qrcode.Encode(content, level, size)
	assert(err)
	img, err := png.Decode(bytes.NewBuffer(buf))
	assert(err)
	return img
}

func saveResult(fn string, img image.Image) {
	f, err := os.Create(fn)
	assert(err)
	defer f.Close()
	assert(png.Encode(f, img))
}

func getBackground(fn string) *image.RGBA {
	f, err := os.Open(fn)
	assert(err)
	defer f.Close()
	img, err := png.Decode(f)
	assert(err)
	canvas := image.NewRGBA(img.Bounds())
	draw.Draw(canvas, canvas.Bounds(), img, image.Point{0, 0}, draw.Src)
	return canvas
}

func paintCode(canvas *image.RGBA, code image.Image, x, y int) {
	b := canvas.Bounds()
	f := code.Bounds()
	if f.Dx() > b.Dx() || f.Dy() > b.Dy() {
		panic("QR Code is larger than background")
	}
}

func addLabel(img *image.RGBA, label string) {
	x := 1
	y := 20
	col := color.RGBA{32, 32, 32, 255}
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

func main() {
	var dbg bool
	defer func() {
		if e := recover(); e != nil {
			msg := trace("ERROR: %v", e)
			if dbg {
				fmt.Println(msg.Error())
			} else {
				fmt.Println(msg[0])
			}
			os.Exit(2)
		}
	}()
	flag.Usage = func() {
		fmt.Printf("QR Code Generator V%s.%s\n\n", _G_REVS, _G_HASH)
		fmt.Printf("USAGE: %s [options] <code-content>\n\n", filepath.Base(os.Args[0]))
		fmt.Println("OPTIONS:\n")
		flag.PrintDefaults()
		fmt.Println("\nNOTE: result will be written to STDOUT")
	}
	bg := flag.String("bg", "", "background image")
	size := flag.Int("size", 0, "size of QR code")
	errlvl := flag.Int("level", 1, "error tolerance level (0~3)")
	xshift := flag.Int("xshift", 0, "x-shift away from center")
	yshift := flag.Int("yshift", 0, "y-shift away from center")
	mark := flag.String("mark", "", "mark text (always appear at lower-left corner)")
	flag.BoolVar(&dbg, "debug", false, "show stack trace on error")
	flag.Parse()
	if *errlvl < 0 || *errlvl > 3 {
		fmt.Println("invalid error tolerance level (-level)")
		fmt.Println("allowed: 0 (Level-L), 1 (Level-M), 2 (Level-Q), 3 (Level-H)")
		os.Exit(1)
	}
	if *size <= 0 {
		fmt.Println("size of code invalid or not specified (-size)")
		os.Exit(1)
	}
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	content := flag.Arg(0)
	code := genCode(content, qrcode.RecoveryLevel(*errlvl), *size)
	if *bg == "" {
		assert(png.Encode(os.Stdout, code))
		return
	}
	canvas := getBackground(*bg)
	paintCode(canvas, code, *xshift, *yshift)
	addLabel(canvas, *mark)
	assert(png.Encode(os.Stdout, canvas))
}
