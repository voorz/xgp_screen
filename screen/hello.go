package screen

// DrawHello renders the milestone-1 screen: big HELLO WORLD plus two
// diagnostic lines built from the real framebuffer parameters (info and
// sub, e.g. "320X240" / "BPP16"). Layout adapts to the actual pixmap size:
// the big block is centered slightly above middle, diagnostics are anchored
// to the bottom with padding. Works for both 240x320 portrait and 320x240
// landscape panels.
func DrawHello(p *Pixmap, info, sub string) {
	p.Fill(Navy)
	p.Border(4, White)

	const big = 5
	bigH := 7 * big
	gapBig := 12

	// Center the HELLO/WORLD block slightly above vertical middle.
	blockH := bigH*2 + gapBig
	y0 := (p.H-blockH)/2 - 10
	if y0 < 8 {
		y0 = 8
	}

	w := TextWidth("HELLO", big)
	p.DrawText((p.W-w)/2, y0, "HELLO", big, White)
	w = TextWidth("WORLD", big)
	p.DrawText((p.W-w)/2, y0+bigH+gapBig, "WORLD", big, White)

	const small = 2
	smallH := 7 * small
	gapSmall := 8
	bottomPad := 16

	// Diagnostics anchored to the bottom.
	if sub != "" {
		w := TextWidth(sub, small)
		p.DrawText((p.W-w)/2, p.H-bottomPad-smallH, sub, small, Green)
	}
	if info != "" {
		w := TextWidth(info, small)
		p.DrawText((p.W-w)/2, p.H-bottomPad-smallH-gapSmall-smallH, info, small, Green)
	}
}
