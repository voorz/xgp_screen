// Command preview renders the dashboard into dash.png (320x240 landscape).
//
//	go run ./tools/preview && open dash.png
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/voorz/xgp_screen/screen"
)

func main() {
	const w, h = 320, 240
	p := screen.New(w, h)

	upHist := make([]float64, 40)
	downHist := make([]float64, 40)
	for i := range upHist {
		upHist[i] = 1.0 + math.Sin(float64(i)*0.3)*0.5 + math.Sin(float64(i)*0.1)*0.3
		downHist[i] = 2.5 + math.Cos(float64(i)*0.25)*0.8 + math.Sin(float64(i)*0.07)*0.4
	}

	d := &screen.DashboardData{
		RouterIP: "192.168.1.1",
		MemUsed:  46 * 1024 * 1024,
		MemTotal: 100 * 1024 * 1024,
		CPUFrac:  0.23,
		UpRate:   1250000,
		DownRate: 3400000,
		UpHist:   upHist,
		DownHist: downHist,
		BaiduOk:  true,
		GoogleOk: false,
	}
	screen.DrawDashboard(p, d)

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
	f, err := os.Create("dash.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote dash.png (320x240)")
}
