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
	"strconv"

	"github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func parseColor(clrspec string) color.Color {
	r, err := strconv.ParseUint(clrspec[:2], 16, 8)
	assert(err)
	g, err := strconv.ParseUint(clrspec[2:4], 16, 8)
	assert(err)
	b, err := strconv.ParseUint(clrspec[4:6], 16, 8)
	assert(err)
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

func genCode(content string, level qrcode.RecoveryLevel, size int, bg, fg color.Color) image.Image {
	q, err := qrcode.New(content, level)
	assert(err)
	q.BackgroundColor = bg
	q.ForegroundColor = fg
	buf, err := q.PNG(size)
	assert(err)
	img, err := png.Decode(bytes.NewBuffer(buf))
	assert(err)
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

func addLabel(img *image.RGBA, label string, shift int) {
	x := shift
	y := img.Bounds().Dy() - shift
	p := img.At(x, y)
	r, g, b, _ := p.RGBA()
	col := color.RGBA{255 - uint8(r), 255 - uint8(g), 255 - uint8(b), 255}
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
	mshift := flag.Int("mshift", 5, "shift of label against bottom-left corner")
	bgcolor := flag.String("bgcolor", "ffffff", "background color for QR code")
	fgcolor := flag.String("fgcolor", "000000", "foreground color for QR code")
	mark := flag.String("mark", "", "mark text (always appear at bottom-left corner)")
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
	bc := parseColor(*bgcolor)
	fc := parseColor(*fgcolor)
	code := genCode(content, qrcode.RecoveryLevel(*errlvl), *size, bc, fc)
	if *bg == "" {
		assert(png.Encode(os.Stdout, code))
		return
	}
	canvas := getBackground(*bg)
	paintCode(canvas, code, *xshift, *yshift)
	if *mark != "" {
		addLabel(canvas, *mark, *mshift)
	}
	assert(png.Encode(os.Stdout, canvas))
}
