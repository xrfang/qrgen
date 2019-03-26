package main

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

func main() {
	err := qrcode.WriteFile("862100000", qrcode.High, 512, "qr.png")
	if err != nil {
		fmt.Println("write error")
	}
}
