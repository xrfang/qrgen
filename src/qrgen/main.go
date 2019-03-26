package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

func genCode(content string, level qr.ErrorCorrectionLevel, size int) image.Image {
	img, _ := qr.Encode(content, level, qr.Auto)
	img, _ = barcode.Scale(img, size, size)
	return img
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

func paintCode(canvas *image.RGBA, code image.Image, xshift, yshift int) {
	b := canvas.Bounds()
	f := code.Bounds()
	if f.Dx() > b.Dx() || f.Dy() > b.Dy() {
		panic("QR Code larger than background")
	}
	x := int((b.Dx()-f.Dx())/2) + xshift
	y := int((b.Dy()-f.Dy())/2) + yshift
	r := image.Rect(x, y, x+f.Dx(), y+f.Dy())
	draw.Draw(canvas, r, code, image.Point{0, 0}, draw.Over)
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
	code := genCode(content, qr.ErrorCorrectionLevel(*errlvl), *size)
	if *bg == "" {
		assert(png.Encode(os.Stdout, code))
		return
	}
	canvas := getBackground(*bg)
	paintCode(canvas, code, *xshift, *yshift)
	assert(png.Encode(os.Stdout, canvas))
}
