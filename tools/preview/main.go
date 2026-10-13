// Command preview renders the hello-world screen into hello.png using
// the same DrawHello code that runs on the device. It exists so the
// rendering can be verified without hardware.
//
//	go run ./tools/preview && open hello.png
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/voorz/xgp_screen/screen"
)

func main() {
	const w, h = 240, 320
	p := screen.New(w, h)
	screen.DrawHello(p, "240X320", "BPP16")

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := p.Pix[y*p.Stride+x]
			r := uint8(v>>11) & 0x1F
			g := uint8(v>>5) & 0x3F
			b := uint8(v) & 0x1F
			img.Set(x, y, color.RGBA{
				R: r<<3 | r>>2,
				G: g<<2 | g>>4,
				B: b<<3 | b>>2,
				A: 0xFF,
			})
		}
	}
	f, err := os.Create("hello.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote hello.png")
}
