package screen

// TTF font rendering with anti-aliasing (replaces the 5x7 pixel font).
// Uses embedded DejaVu Sans, a clean sans-serif in the spirit of MiSans.

import (
	_ "embed"
	"image"
	"image/draw"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed DejaVuSans.ttf
var ttfData []byte

var (
	ttfFont *opentype.Font
	faces   = make(map[int]font.Face)
	facesMu sync.Mutex
)

func init() {
	var err error
	ttfFont, err = opentype.Parse(ttfData)
	if err != nil {
		panic("parse font: " + err.Error())
	}
}

// faceFor returns a cached Face at the given pixel height.
func faceFor(px int) font.Face {
	facesMu.Lock()
	defer facesMu.Unlock()
	if f, ok := faces[px]; ok {
		return f
	}
	f, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{
		Size:    float64(px),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic("new face: " + err.Error())
	}
	faces[px] = f
	return f
}

// scaleToPx maps the old bitmap scale to pixel height.
// Old: 5x7 bitmap * scale. New: 7*scale px height (scale=2 → 14px).
func scaleToPx(scale int) int {
	if scale < 1 {
		scale = 1
	}
	return scale * 7
}

// DrawText draws s at (x, y) with anti-aliased TTF rendering.
// scale maps to pixel height (scale=2 → 14px, matches old bitmap size).
// (x, y) is the top-left of the text bounding box.
func (p *Pixmap) DrawText(x, y int, s string, scale int, c uint16) {
	px := scaleToPx(scale)
	face := faceFor(px)

	// Measure to position.
	d := &font.Drawer{Face: face}
	// Draw to a temporary RGBA, then blend to RGB565.
	// Estimate bounds.
	bounds, _ := d.BoundString(s)
	w := (bounds.Max.X - bounds.Min.X).Ceil() + 2
	h := (bounds.Max.Y - bounds.Min.Y).Ceil() + 2
	if w <= 0 || h <= 0 {
		return
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// White text on transparent; we'll tint via blending.
	d.Dst = img
	d.Src = image.White
	d.Dot = fixed.P(-bounds.Min.X.Ceil()+1, -bounds.Min.Y.Ceil()+1)
	d.DrawString(s)

	// Blend into Pixmap with the target color.
	r5 := (c >> 11) & 0x1F
	g6 := (c >> 5) & 0x3F
	b5 := c & 0x1F
	// Expand to 8-bit.
	r8 := uint8(r5<<3 | r5>>2)
	g8 := uint8(g6<<2 | g6>>4)
	b8 := uint8(b5<<3 | b5>>2)

	for iy := 0; iy < h; iy++ {
		for ix := 0; ix < w; ix++ {
			a := img.RGBAAt(ix, iy).A
			if a == 0 {
				continue
			}
			pxX, pxY := x+ix, y+iy
			if pxX < 0 || pxY < 0 || pxX >= p.W || pxY >= p.H {
				continue
			}
			if a == 255 {
				p.Set(pxX, pxY, c)
				continue
			}
			// Alpha blend.
			bg := p.Pix[pxY*p.Stride+pxX]
			bgR := uint8((bg>>11)&0x1F)<<3 | uint8((bg>>11)&0x1F)>>2
			bgG := uint8((bg>>5)&0x3F)<<2 | uint8((bg>>5)&0x3F)>>4
			bgB := uint8(bg&0x1F)<<3 | uint8(bg&0x1F)>>2
			af := float64(a) / 255.0
			rf := float64(r8)*af + float64(bgR)*(1-af)
			gf := float64(g8)*af + float64(bgG)*(1-af)
			bf := float64(b8)*af + float64(bgB)*(1-af)
			p.Set(pxX, pxY, RGB565(uint8(rf), uint8(gf), uint8(bf)))
		}
	}
}

// TextWidth returns the pixel width of s at the given scale.
func TextWidth(s string, scale int) int {
	px := scaleToPx(scale)
	face := faceFor(px)
	d := &font.Drawer{Face: face}
	bounds, _ := d.BoundString(s)
	return (bounds.Max.X - bounds.Min.X).Ceil() + 2
}

// Ensure the old font.go doesn't conflict: we keep DrawText/TextWidth here.
// The bitmap data in font.go is now unused but kept for reference.
var _ = draw.Src
