package screen

// DrawHello renders the milestone-1 screen: big HELLO WORLD plus two
// diagnostic lines built from the real framebuffer parameters (info and
// sub, e.g. "240X320" / "BPP16"). Everything is centered for the panel.
func DrawHello(p *Pixmap, info, sub string) {
	p.Fill(Navy)
	p.Border(4, White)

	const big = 5
	w := TextWidth("HELLO", big)
	p.DrawText((p.W-w)/2, 70, "HELLO", big, White)
	w = TextWidth("WORLD", big)
	p.DrawText((p.W-w)/2, 70+7*big+12, "WORLD", big, White)

	const small = 2
	if info != "" {
		w := TextWidth(info, small)
		p.DrawText((p.W-w)/2, 232, info, small, Green)
	}
	if sub != "" {
		w := TextWidth(sub, small)
		p.DrawText((p.W-w)/2, 232+7*small+8, sub, small, Green)
	}
}
