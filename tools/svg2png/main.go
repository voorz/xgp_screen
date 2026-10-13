// Command svg2png converts SVG files to PNG using oksvg.
//
//	go run ./tools/svg2png <input.svg> <output.png> <size>
package main

import (
	"image"
	"image/png"
	"log"
	"os"
	"strconv"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatal("usage: svg2png <input.svg> <output.png> <size>")
	}
	inPath, outPath := os.Args[1], os.Args[2]
	size, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}

	in, err := os.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	icon, err := oksvg.ReadIconStream(in)
	if err != nil {
		log.Fatal(err)
	}
	icon.SetTarget(0, 0, float64(size), float64(size))

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	raster := rasterx.NewDasher(size, size, scanner)
	icon.Draw(raster, 1.0)

	out, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s (%dx%d)", outPath, size, size)
}
