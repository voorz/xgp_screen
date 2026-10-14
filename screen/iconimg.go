package screen

// Icon rendering for connectivity status (Baidu/Google).
// PNGs are embedded; drawn with alpha blending. Grayscale version is
// computed on the fly for the "disconnected" state.

import (
	"bytes"
	"embed"
	"image"
	"image/png"
	"sync"
)

//go:embed icons/baidu.png icons/google.png icons/clash.png
var iconFS embed.FS

var (
	iconCache   = make(map[string]*iconImg)
	iconCacheMu sync.Mutex
)

type iconImg struct {
	img    image.Image // RGBA
	w, h   int
	gray   []uint16 // precomputed grayscale RGB565
	native []uint16 // precomputed native RGB565 (for opaque pixels)
}

// loadIcon loads and caches an icon by name ("baidu", "google").
func loadIcon(name string) *iconImg {
	iconCacheMu.Lock()
	defer iconCacheMu.Unlock()
	if ic, ok := iconCache[name]; ok {
		return ic
	}
	data, err := iconFS.ReadFile("icons/" + name + ".png")
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	ic := &iconImg{img: img, w: w, h: h}
	ic.gray = make([]uint16, w*h)
	ic.native = make([]uint16, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			// r,g,b are 16-bit (0-65535).
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			ic.native[y*w+x] = RGB565(r8, g8, b8)
			// Grayscale via luminance.
			lum := uint8((299*int(r8) + 587*int(g8) + 114*int(b8)) / 1000)
			ic.gray[y*w+x] = RGB565(lum, lum, lum)
		}
	}
	iconCache[name] = ic
	return ic
}

// DrawIconImage draws the named icon centered at (cx, cy) with the given
// diameter. If color is true, draws in native colors; otherwise grayscale.
// Alpha blending is applied from the PNG's alpha channel.
func (p *Pixmap) DrawIconImage(cx, cy int, name string, diameter int, color bool) {
	ic := loadIcon(name)
	if ic == nil {
		// Fallback: draw a circle.
		if color {
			p.DrawCircle(cx, cy, diameter/2, Blue)
		} else {
			p.DrawCircle(cx, cy, diameter/2, colGray)
		}
		return
	}
	// Scale icon to diameter.
	// Simple nearest-neighbor scaling.
	srcW, srcH := ic.w, ic.h
	for dy := -diameter / 2; dy < diameter/2; dy++ {
		for dx := -diameter / 2; dx < diameter/2; dx++ {
			px, py := cx+dx, cy+dy
			if px < 0 || py < 0 || px >= p.W || py >= p.H {
				continue
			}
			// Map to source.
			sx := (dx + diameter/2) * srcW / diameter
			sy := (dy + diameter/2) * srcH / diameter
			if sx < 0 || sx >= srcW || sy < 0 || sy >= srcH {
				continue
			}
			// Get alpha.
			_, _, _, a := ic.img.At(sx, sy).RGBA()
			alpha := uint8(a >> 8)
			if alpha == 0 {
				continue
			}
			var c uint16
			if color {
				c = ic.native[sy*srcW+sx]
			} else {
				c = ic.gray[sy*srcW+sx]
			}
			if alpha == 255 {
				p.Set(px, py, c)
			} else {
				// Blend.
				bg := p.Pix[py*p.Stride+px]
				bgR := uint8((bg>>11)&0x1F)<<3 | uint8((bg>>11)&0x1F)>>2
				bgG := uint8((bg>>5)&0x3F)<<2 | uint8((bg>>5)&0x3F)>>4
				bgB := uint8(bg&0x1F)<<3 | uint8(bg&0x1F)>>2
				fgR := uint8((c>>11)&0x1F)<<3 | uint8((c>>11)&0x1F)>>2
				fgG := uint8((c>>5)&0x3F)<<2 | uint8((c>>5)&0x3F)>>4
				fgB := uint8(c&0x1F)<<3 | uint8(c&0x1F)>>2
				af := float64(alpha) / 255.0
				rf := float64(fgR)*af + float64(bgR)*(1-af)
				gf := float64(fgG)*af + float64(bgG)*(1-af)
				bf := float64(fgB)*af + float64(bgB)*(1-af)
				p.Set(px, py, RGB565(uint8(rf), uint8(gf), uint8(bf)))
			}
		}
	}
}
