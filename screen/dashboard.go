package screen

// DashboardData is everything the single-page dashboard shows.
// Rates are bytes/sec; UpHist/DownHist are oldest-to-newest samples
// for the sparklines. BaiduOk/GoogleOk drive the connectivity icons.
type DashboardData struct {
	RouterIP string // e.g. "192.168.1.1" (bottom-right, replaces time)
	MemUsed  uint64 // bytes
	MemTotal uint64 // bytes
	CPUFrac  float64
	UpRate   uint64 // bytes/sec
	DownRate uint64 // bytes/sec
	UpHist   []float64
	DownHist []float64
	BaiduOk  bool
	GoogleOk bool
	NikkiOk  bool // true when nikki (mihomo) process is running
}

var (
	colBg     uint16 = Black
	colFg     uint16 = White
	colDim    uint16 = RGB565(0x80, 0x80, 0x80)
	colAccent uint16 = Green
	colUp     uint16 = RGB565(0x00, 0xC0, 0xFF) // cyan-ish for upload
	colDown   uint16 = RGB565(0xFF, 0xA0, 0x00) // orange-ish for download
	colGood   uint16 = Green
	colBad    uint16 = Red
	colBarBg  uint16 = RGB565(0x20, 0x20, 0x30)
	colBaidu  uint16 = RGB565(0x30, 0x60, 0xFF) // baidu blue
	colGoogle uint16 = RGB565(0xFF, 0x40, 0x40) // google red (placeholder)
	colGray   uint16 = RGB565(0x60, 0x60, 0x60)
)

// DrawDashboard renders the single-page watch-style dashboard for the
// 320x240 landscape panel. All blocks are vertically centered with
// balanced spacing (8px label-bar, 12px block gaps).
//
// Blocks:
//	MEM/CPU: 28px (label 14 + gap 8 + bar 6)
//	gap: 12px
//	Chart: 70px
//	gap: 12px
//	Icons: 22px (centered)
//	Total: 134px, centered in 240px.
//
// Merged chart: UP cyan + DOWN orange, shared scale.
// Time is deliberately inconspicuous.
func DrawDashboard(p *Pixmap, d *DashboardData) {
	p.Fill(colBg)

	// ---- Vertical centering ----
	const contentH = 134
	startY := (p.H - contentH) / 2
	if startY < 0 {
		startY = 0
	}

	// ---- Memory + CPU thin bars ----
	// Label (14px) + 8px gap + bar (6px).
	const barH = 6
	barY := startY
	const labelBarGap = 8

	// MEM (left half) - labeled RAM
	p.DrawText(12, barY, "RAM", 2, colDim)
	memFrac := 0.0
	if d.MemTotal > 0 {
		memFrac = float64(d.MemUsed) / float64(d.MemTotal)
	}
	memPct := sprintfPct(memFrac)
	p.DrawText(12+130-TextWidth(memPct, 2), barY, memPct, 2, colFg)
	p.DrawBar(12, barY+14+labelBarGap, 130, barH, memFrac, colAccent, colBarBg, colDim, false)

	// CPU (right half)
	p.DrawText(170, barY, "CPU", 2, colDim)
	cpuPct := sprintfPct(d.CPUFrac)
	p.DrawText(170+130-TextWidth(cpuPct, 2), barY, cpuPct, 2, colFg)
	p.DrawBar(170, barY+14+labelBarGap, 130, barH, d.CPUFrac, colUp, colBarBg, colDim, false)

	// ---- UP/DOWN text (left) + merged chart (right) ----
	// Block gap 12px after MEM/CPU block (ends at startY+28).
	midY := startY + 28 + 12
	const chartH = 70
	// UP text (left, top)
	p.DrawIcon(12, midY, arrowUpIcon, 2, colUp)
	iw, _ := IconSize(arrowUpIcon, 2)
	p.DrawText(12+iw+6, midY+2, "UP", 2, colDim)
	upRate := formatRate(d.UpRate)
	p.DrawText(12, midY+18, upRate, 2, colFg)
	// DOWN text (left, bottom)
	const downTextDy = 35
	p.DrawIcon(12, midY+downTextDy, arrowDownIcon, 2, colDown)
	p.DrawText(12+iw+6, midY+downTextDy+2, "DOWN", 2, colDim)
	downRate := formatRate(d.DownRate)
	p.DrawText(12, midY+downTextDy+18, downRate, 2, colFg)

	// Merged chart (right). Right edge aligned with CPU bar (x=300).
	sx := 110
	sw := 300 - sx // 190, right edge at x=300
	p.DrawBar(sx, midY, sw, chartH, 0, colBg, colBg, colDim, true)
	p.DrawDualSparkline(sx+2, midY+2, sw-4, chartH-4, d.UpHist, d.DownHist, colUp, colDown)

	// ---- Connectivity icons (bottom-left) + time (bottom-right) ----
	// 12px gap after chart (ends at midY+70).
	cy := midY + 70 + 12 + 11 // 11 = radius, center the 22px icon
	const iconD = 22
	const iconStep = 34
	ix := 12 + iconD/2
	// Baidu (color if ok, grayscale if not)
	p.DrawIconImage(ix, cy, "baidu", iconD, d.BaiduOk)
	ix += iconStep
	// Google (color if ok, grayscale if not)
	p.DrawIconImage(ix, cy, "google", iconD, d.GoogleOk)
	ix += iconStep
	// Nikki/Clash: only show when running (hide when stopped).
	if d.NikkiOk {
		p.DrawIconImage(ix, cy, "clash", iconD, true)
	}

	// Router IP (bottom-right, replaces time).
	// Vertically centered with icons.
	ipw := TextWidth(d.RouterIP, 2)
	p.DrawText(p.W-12-ipw, cy-7, d.RouterIP, 2, colDim)
}

func sprintfPct(frac float64) string {
	pct := int(frac*100 + 0.5)
	return itoa(pct) + "%"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func formatBytes(b uint64) string {
	if b >= 1<<30 {
		return itoa(int(b>>30)) + "G"
	}
	if b >= 1<<20 {
		return itoa(int(b>>20)) + "M"
	}
	if b >= 1<<10 {
		return itoa(int(b>>10)) + "K"
	}
	return itoa(int(b))
}

func formatRate(bps uint64) string {
	// Network rates shown in bits/sec, e.g. "3.5Mb/s".
	bits := bps * 8
	if bits >= 1000*1000 {
		whole := bits / (1000 * 1000)
		frac := (bits - whole*1000*1000) * 10 / (1000 * 1000)
		return itoa(int(whole)) + "." + itoa(int(frac)) + "Mb/s"
	}
	if bits >= 1000 {
		whole := bits / 1000
		frac := (bits - whole*1000) * 10 / 1000
		return itoa(int(whole)) + "." + itoa(int(frac)) + "Kb/s"
	}
	return itoa(int(bits)) + "b/s"
}
