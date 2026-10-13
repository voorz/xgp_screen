package screen

// Hand-drawn 5x7 bitmap font. Each glyph is 7 rows of 5 pixels,
// '#' = pixel on, '.' = off. Only the characters needed for the
// hello-world milestone plus diagnostics are included; unknown
// characters render as blank space.
var fontSrc = map[byte][7]string{
	'H': {"#...#", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'E': {"#####", "#....", "#....", "####.", "#....", "#....", "#####"},
	'L': {"#....", "#....", "#....", "#....", "#....", "#....", "#####"},
	'O': {".###.", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."},
	'W': {"#...#", "#...#", "#...#", "#.#.#", "#.#.#", "#.#.#", ".#.#."},
	'R': {"####.", "#...#", "#...#", "####.", "#.#..", "#..#.", "#...#"},
	'D': {"####.", "#...#", "#...#", "#...#", "#...#", "#...#", "####."},
	'G': {".###.", "#...#", "#....", "#.###", "#...#", "#...#", ".###."},
	'B': {"####.", "#...#", "#...#", "####.", "#...#", "#...#", "####."},
	'P': {"####.", "#...#", "#...#", "####.", "#....", "#....", "#...."},
	'X': {"#...#", "#...#", ".#.#.", "..#..", ".#.#.", "#...#", "#...#"},
	'0': {".###.", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."},
	'1': {"..#..", ".##..", "..#..", "..#..", "..#..", "..#..", ".###."},
	'2': {".###.", "#...#", "....#", "...#.", "..#..", ".#...", "#####"},
	'3': {"#####", "...#.", "..#..", "...#.", "....#", "#...#", ".###."},
	'4': {"...#.", "..##.", ".#.#.", "#..#.", "#####", "...#.", "...#."},
	'5': {"#####", "#....", "####.", "....#", "....#", "#...#", ".###."},
	'6': {".###.", "#....", "#....", "####.", "#...#", "#...#", ".###."},
	'7': {"#####", "....#", "...#.", "..#..", ".#...", ".#...", ".#..."},
	'8': {".###.", "#...#", "#...#", ".###.", "#...#", "#...#", ".###."},
	'9': {".###.", "#...#", "#...#", ".####", "....#", "....#", ".###."},
}

// font5x7 holds the compiled glyphs: bit 4 of each byte is the leftmost
// pixel of that row.
var font5x7 = map[byte][7]byte{}

func init() {
	for ch, rows := range fontSrc {
		var g [7]byte
		for r, row := range rows {
			var b byte
			for c := 0; c < 5 && c < len(row); c++ {
				if row[c] == '#' {
					b |= 1 << (4 - c)
				}
			}
			g[r] = b
		}
		font5x7[ch] = g
	}
}

// DrawText draws s left-to-right with the 5x7 font scaled by scale.
// Unknown characters (including lowercase) render as a blank advance.
func (p *Pixmap) DrawText(x, y int, s string, scale int, c uint16) {
	cx := x
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if g, ok := font5x7[ch]; ok {
			for r := 0; r < 7; r++ {
				for col := 0; col < 5; col++ {
					if g[r]&(1<<(4-col)) != 0 {
						p.FillRect(cx+col*scale, y+r*scale, scale, scale, c)
					}
				}
			}
		}
		cx += (5 + 1) * scale
	}
}

// TextWidth returns the pixel width of s rendered at the given scale.
func TextWidth(s string, scale int) int {
	if len(s) == 0 {
		return 0
	}
	return len(s)*(5+1)*scale - scale
}
