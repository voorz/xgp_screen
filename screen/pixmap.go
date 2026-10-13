// Package screen provides a minimal RGB565 software renderer for the
// XGP v3 GC9307 TFT panel (240x320, accessed via /dev/fb0).
//
// The drawing layer works on an abstract Pixmap so it can be verified
// without hardware (see tools/preview); the Linux framebuffer backend
// (fb.go) maps the real device into the same Pixmap shape.
package screen

// Common RGB565 colors.
const (
	Black = 0x0000
	White = 0xFFFF
	Red   = 0xF800
	Green = 0x07E0
	Blue  = 0x001F
	Navy  = 0x0010
)

// RGB565 converts 8-bit R/G/B to a 16-bit RGB565 pixel.
func RGB565(r, g, b uint8) uint16 {
	return uint16(r>>3)<<11 | uint16(g>>2)<<5 | uint16(b>>3)
}

// Pixmap is a 16-bit RGB565 pixel buffer. Stride is pixels per row and may
// be larger than W (framebuffer line_length padding).
type Pixmap struct {
	W, H   int
	Stride int
	Pix    []uint16
}

// New allocates a W*H pixmap with tight stride.
func New(w, h int) *Pixmap {
	return &Pixmap{W: w, H: h, Stride: w, Pix: make([]uint16, w*h)}
}

// Set writes one pixel, silently clipping out-of-bounds coordinates.
func (p *Pixmap) Set(x, y int, c uint16) {
	if x < 0 || y < 0 || x >= p.W || y >= p.H {
		return
	}
	p.Pix[y*p.Stride+x] = c
}

// Fill paints the whole visible area.
func (p *Pixmap) Fill(c uint16) {
	for y := 0; y < p.H; y++ {
		row := p.Pix[y*p.Stride : y*p.Stride+p.W]
		for x := range row {
			row[x] = c
		}
	}
}

// FillRect paints a rectangle, clipped to the visible area.
func (p *Pixmap) FillRect(x, y, w, h int, c uint16) {
	if w <= 0 || h <= 0 {
		return
	}
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x+w > p.W {
		w = p.W - x
	}
	if y+h > p.H {
		h = p.H - y
	}
	if w <= 0 || h <= 0 {
		return
	}
	for yy := y; yy < y+h; yy++ {
		row := p.Pix[yy*p.Stride+x : yy*p.Stride+x+w]
		for xx := range row {
			row[xx] = c
		}
	}
}

// Border draws a frame of the given thickness around the visible area.
func (p *Pixmap) Border(thick int, c uint16) {
	p.FillRect(0, 0, p.W, thick, c)
	p.FillRect(0, p.H-thick, p.W, thick, c)
	p.FillRect(0, 0, thick, p.H, c)
	p.FillRect(p.W-thick, 0, thick, p.H, c)
}
