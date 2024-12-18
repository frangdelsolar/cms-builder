package tools

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

func GenerateQR() {
	err := qrcode.WriteFile("https://example.org", qrcode.Medium, 256, "qr.png")
	if err != nil {
		fmt.Println(err)
	}
}
