package screen

// Bitmap icons drawn from string maps ('#' = filled). All icons are
// monochrome; the caller picks the color.

// arrowUpIcon / arrowDownIcon: 8x8 traffic direction arrows.
var arrowUpIcon = []string{
	"...##...",
	"..####..",
	".######.",
	"########",
	"...##...",
	"...##...",
	"...##...",
	"...##...",
}

var arrowDownIcon = []string{
	"...##...",
	"...##...",
	"...##...",
	"...##...",
	"########",
	".######.",
	"..####..",
	"...##...",
}

// checkIcon: 8x8 checkmark (ok status).
var checkIcon = []string{
	"........",
	".....##.",
	"....##..",
	"...##...",
	"##.##...",
	".####...",
	"..##....",
	"........",
}

// crossIcon: 8x8 cross (fail status).
var crossIcon = []string{
	"##....##",
	".##..##.",
	"..####..",
	"...##...",
	"...##...",
	"..####..",
	".##..##.",
	"##....##",
}

// chipIcon: 12x10 memory chip with side pins.
var chipIcon = []string{
	"..########..",
	"..########..",
	"############",
	"##.######.##",
	"##.######.##",
	"##.######.##",
	"##.######.##",
	"############",
	"..########..",
	"..########..",
}

// globeIcon: 10x10 network globe (connectivity).
var globeIcon = []string{
	"...####...",
	".########.",
	"##########",
	"##########",
	"##########",
	"##########",
	"##########",
	"##########",
	".########.",
	"...####...",
}

// DrawIcon paints a string-map icon at (x, y) in color c.
// Scale enlarges each bitmap pixel to scale x scale device pixels.
func (p *Pixmap) DrawIcon(x, y int, icon []string, scale int, c uint16) {
	if scale < 1 {
		scale = 1
	}
	for ry, row := range icon {
		for rx, ch := range row {
			if ch != '#' {
				continue
			}
			p.FillRect(x+rx*scale, y+ry*scale, scale, scale, c)
		}
	}
}

// IconSize returns the device pixel size of an icon at the given scale.
func IconSize(icon []string, scale int) (w, h int) {
	if len(icon) == 0 {
		return 0, 0
	}
	return len(icon[0]) * scale, len(icon) * scale
}
