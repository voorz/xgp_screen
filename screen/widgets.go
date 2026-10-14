package screen

// Widgets: progress bar and sparkline chart for the dashboard.

// DrawBar paints a horizontal progress bar at (x, y, w, h).
// frac is 0.0..1.0 (clamped). The track is drawn in trackC, the filled
// portion in fillC. If border is true, a 1px border in borderC is drawn.
// Both track and fill use rounded corners (pill shape, radius = h/2).
// The fill's right end is rounded only when full; otherwise flat.
func (p *Pixmap) DrawBar(x, y, w, h int, frac float64, fillC, trackC, borderC uint16, border bool) {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	ix, iy, iw, ih := x, y, w, h
	if border {
		// Border.
		p.FillRect(x, y, w, 1, borderC)
		p.FillRect(x, y+h-1, w, 1, borderC)
		p.FillRect(x, y, 1, h, borderC)
		p.FillRect(x+w-1, y, 1, h, borderC)
		// Track inset by border.
		ix, iy, iw, ih = x+1, y+1, w-2, h-2
	}
	if iw <= 0 || ih <= 0 {
		return
	}
	// Track: full pill.
	p.FillRoundedRect(ix, iy, iw, ih, ih/2, trackC)
	// Fill: pill shape (rounded on both ends), even when partial.
	fw := int(float64(iw) * frac)
	if fw <= 0 {
		return
	}
	r := ih / 2
	if fw >= iw {
		p.FillRoundedRect(ix, iy, iw, ih, r, fillC)
		return
	}
	if fw <= ih {
		// Very short: circle with diameter = fill width.
		p.DrawCircle(ix+fw/2, iy+ih/2, fw/2, fillC)
		return
	}
	// Partial pill: rounded on both ends.
	p.FillRoundedRect(ix, iy, fw, ih, r, fillC)
}

// FillRoundedRect paints a filled rounded rectangle.
// Radius is clamped to half of the smaller dimension.
func (p *Pixmap) FillRoundedRect(x, y, w, h, r int, c uint16) {
	if w <= 0 || h <= 0 {
		return
	}
	if r <= 0 {
		p.FillRect(x, y, w, h, c)
		return
	}
	maxR := h / 2
	if w/2 < maxR {
		maxR = w / 2
	}
	if r > maxR {
		r = maxR
	}
	// Center rect (full width minus corner circles, full height).
	p.FillRect(x+r, y, w-2*r, h, c)
	// Left and right side rects (between top and bottom corners).
	if h > 2*r {
		p.FillRect(x, y+r, r, h-2*r, c)
		p.FillRect(x+w-r, y+r, r, h-2*r, c)
	}
	// Four corner circles.
	p.DrawCircle(x+r, y+r, r, c)
	p.DrawCircle(x+w-r-1, y+r, r, c)
	p.DrawCircle(x+r, y+h-r-1, r, c)
	p.DrawCircle(x+w-r-1, y+h-r-1, r, c)
}

// DrawSparkline paints a line chart of vals inside (x, y, w, h).
// vals[0] is the oldest sample, vals[len-1] the newest (right edge).
// Values are normalized to [min, max]; if max <= min the line is flat
// at mid height.
//
// Style (standard chart look):
//   - 3 horizontal grid lines
//   - area under the curve filled with a dimmed version of c
//   - 2px thick smooth line in c
//   - bright dot on the newest point
func (p *Pixmap) DrawSparkline(x, y, w, h int, vals []float64, c uint16) {
	if w < 4 || h < 4 || len(vals) == 0 {
		return
	}
	// Find range.
	min, max := vals[0], vals[0]
	for _, v := range vals[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max <= min {
		max = min + 1
	}
	n := len(vals)
	px := func(i int) int {
		if n == 1 {
			return x + w - 1
		}
		return x + i*(w-1)/(n-1)
	}
	py := func(v float64) int {
		f := (v - min) / (max - min)
		return y + h - 1 - int(f*float64(h-1))
	}

	// Grid lines (3 horizontal).
	gridC := RGB565(0x30, 0x30, 0x40)
	for g := 1; g <= 3; g++ {
		gy := y + g*h/4
		for gx := x; gx < x+w; gx += 3 {
			p.Set(gx, gy, gridC)
		}
	}

	// Area fill under the curve (dimmed color).
	r5 := (c >> 11) & 0x1F
	g6 := (c >> 5) & 0x3F
	b5 := c & 0x1F
	fillC := uint16(r5>>2)<<11 | uint16(g6>>2)<<5 | uint16(b5>>2)
	for i := 0; i+1 < n; i++ {
		x0, x1 := px(i), px(i+1)
		y0 := py(vals[i])
		// Fill vertical strip from curve down to bottom.
		for xx := x0; xx <= x1; xx++ {
			t := float64(xx-x0) / float64(x1-x0+1)
			yy := int(float64(y0)*(1-t) + float64(py(vals[i+1]))*t)
			for yyy := yy; yyy < y+h; yyy++ {
				p.Set(xx, yyy, fillC)
			}
		}
	}

	// 2px thick line.
	for i := 0; i+1 < n; i++ {
		p.drawThickLine(px(i), py(vals[i]), px(i+1), py(vals[i+1]), c)
	}
	// Newest point dot.
	lx, ly := px(n-1), py(vals[n-1])
	p.FillRect(lx-2, ly-2, 5, 5, White)
	p.FillRect(lx-1, ly-1, 3, 3, c)
}

// drawThickLine paints a 2px line.
func (p *Pixmap) drawThickLine(x0, y0, x1, y1 int, c uint16) {
	p.drawLine(x0, y0, x1, y1, c)
	p.drawLine(x0+1, y0, x1+1, y1, c)
	p.drawLine(x0, y0+1, x1, y1+1, c)
}

// DrawDualSparkline paints two line charts sharing one scale.
// vals1 (c1) is drawn first, vals2 (c2) on top. No area fill to avoid
// visual clutter; 2px lines with newest-point dots.
func (p *Pixmap) DrawDualSparkline(x, y, w, h int, vals1, vals2 []float64, c1, c2 uint16) {
	if w < 4 || h < 4 || len(vals1) == 0 || len(vals2) == 0 {
		return
	}
	// Shared range across both series.
	min, max := vals1[0], vals1[0]
	for _, v := range vals1[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	for _, v := range vals2 {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max <= min {
		max = min + 1
	}
	// Grid lines (3 horizontal).
	gridC := RGB565(0x30, 0x30, 0x40)
	for g := 1; g <= 3; g++ {
		gy := y + g*h/4
		for gx := x; gx < x+w; gx += 3 {
			p.Set(gx, gy, gridC)
		}
	}
	drawOne := func(vals []float64, c uint16) {
		n := len(vals)
		px := func(i int) int {
			if n == 1 {
				return x + w - 1
			}
			return x + i*(w-1)/(n-1)
		}
		py := func(v float64) int {
			f := (v - min) / (max - min)
			return y + h - 1 - int(f*float64(h-1))
		}
		for i := 0; i+1 < n; i++ {
			p.drawThickLine(px(i), py(vals[i]), px(i+1), py(vals[i+1]), c)
		}
		// Newest point dot.
		lx, ly := px(n-1), py(vals[n-1])
		p.FillRect(lx-2, ly-2, 5, 5, White)
		p.FillRect(lx-1, ly-1, 3, 3, c)
	}
	// Draw the larger-range series first so the smaller stays visible.
	drawOne(vals1, c1)
	drawOne(vals2, c2)
}

// drawLine paints a 1px line using Bresenham, clipped by Set.
func (p *Pixmap) drawLine(x0, y0, x1, y1 int, c uint16) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		p.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// DrawCircle paints a filled circle centered at (cx, cy) with radius r.
func (p *Pixmap) DrawCircle(cx, cy, r int, c uint16) {
	if r <= 0 {
		return
	}
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				p.Set(cx+x, cy+y, c)
			}
		}
	}
}

// DrawCircleOutline paints a 1px circle outline.
func (p *Pixmap) DrawCircleOutline(cx, cy, r int, c uint16) {
	if r <= 0 {
		return
	}
	x, y := r, 0
	err := 0
	for x >= y {
		p.Set(cx+x, cy+y, c)
		p.Set(cx+y, cy+x, c)
		p.Set(cx-y, cy+x, c)
		p.Set(cx-x, cy+y, c)
		p.Set(cx-x, cy-y, c)
		p.Set(cx-y, cy-x, c)
		p.Set(cx+y, cy-x, c)
		p.Set(cx+x, cy-y, c)
		y++
		if err <= 0 {
			err += 2*y + 1
		} else {
			x--
			err -= 2*x + 1
		}
	}
}
