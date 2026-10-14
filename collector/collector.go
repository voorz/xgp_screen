// Package collector gathers the live data shown on the dashboard:
// memory usage, network rates, and connectivity status.
package collector

import (
	"bufio"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// MemInfo holds memory usage in bytes.
type MemInfo struct {
	Used  uint64
	Total uint64
}

// ReadMem parses /proc/meminfo. Used = Total - Available (falls back to
// MemFree if MemAvailable is missing, e.g. very old kernels).
func ReadMem() (MemInfo, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemInfo{}, err
	}
	defer f.Close()

	var total, avail, free uint64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		val, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		val *= 1024 // kB -> bytes
		switch fields[0] {
		case "MemTotal:":
			total = val
		case "MemAvailable:":
			avail = val
		case "MemFree:":
			free = val
		}
	}
	if err := sc.Err(); err != nil {
		return MemInfo{}, err
	}
	if avail == 0 {
		avail = free
	}
	var used uint64
	if total > avail {
		used = total - avail
	}
	return MemInfo{Used: used, Total: total}, nil
}

// NetRates holds per-second byte rates.
type NetRates struct {
	Rx uint64 // download bytes/sec
	Tx uint64 // upload bytes/sec
}

// netSnapshot is a point-in-time counter reading.
type netSnapshot struct {
	rx, tx uint64
	at     time.Time
}

// NetMonitor tracks interface byte counters and computes rates.
// It sums all non-loopback interfaces.
type NetMonitor struct {
	prev netSnapshot
	hist []float64 // recent Rx rates (for sparkline)
	histTx []float64
}

// NewNetMonitor creates a monitor and takes the initial snapshot.
func NewNetMonitor() *NetMonitor {
	m := &NetMonitor{}
	m.prev = readNetCounters()
	return m
}

// Sample returns the current Rx/Tx rates based on the delta since the
// last Sample (or construction). First call may return zeros.
func (m *NetMonitor) Sample() NetRates {
	cur := readNetCounters()
	dt := cur.at.Sub(m.prev.at).Seconds()
	var r NetRates
	if dt > 0 {
		if cur.rx >= m.prev.rx {
			r.Rx = uint64(float64(cur.rx-m.prev.rx) / dt)
		}
		if cur.tx >= m.prev.tx {
			r.Tx = uint64(float64(cur.tx-m.prev.tx) / dt)
		}
	}
	m.prev = cur
	// Keep last 40 samples for sparklines.
	m.hist = append(m.hist, float64(r.Rx))
	m.histTx = append(m.histTx, float64(r.Tx))
	if len(m.hist) > 40 {
		m.hist = m.hist[1:]
		m.histTx = m.histTx[1:]
	}
	return r
}

// Histories returns the recent Rx/Tx rate histories (oldest first).
func (m *NetMonitor) Histories() (rx, tx []float64) {
	return m.hist, m.histTx
}

// readNetCounters sums rx/tx bytes over all non-loopback interfaces
// from /proc/net/dev.
func readNetCounters() netSnapshot {
	snap := netSnapshot{at: time.Now()}
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return snap
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		// Format: "eth0: 12345 ... 67890 ..."
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)
		snap.rx += rx
		snap.tx += tx
	}
	return snap
}

// CheckConnectivity probes whether a URL is reachable within timeout.
// Used for Baidu (real internet) and Google (proxy/VPN environment).
func CheckConnectivity(url string, timeout time.Duration) bool {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Head(url)
	if err != nil {
		// Fall back to GET; some endpoints reject HEAD.
		resp, err = client.Get(url)
		if err != nil {
			return false
		}
	}
	defer resp.Body.Close()
	return resp.StatusCode < 400
}

// cpuSnapshot is a point-in-time CPU counter reading.
type cpuSnapshot struct {
	idle, total uint64
	at          time.Time
}

// CPUMonitor tracks /proc/stat and computes usage percentage.
type CPUMonitor struct {
	prev cpuSnapshot
}

// NewCPUMonitor creates a monitor and takes the initial snapshot.
func NewCPUMonitor() *CPUMonitor {
	m := &CPUMonitor{}
	m.prev = readCPUCounters()
	return m
}

// Sample returns CPU usage as 0.0..1.0 based on the delta since the
// last Sample (or construction). First call may return 0.
func (m *CPUMonitor) Sample() float64 {
	cur := readCPUCounters()
	var frac float64
	dIdle := cur.idle - m.prev.idle
	dTotal := cur.total - m.prev.total
	// Handle counter wrap (unlikely but safe).
	if cur.idle >= m.prev.idle && cur.total > m.prev.total && dTotal > 0 {
		frac = 1 - float64(dIdle)/float64(dTotal)
		if frac < 0 {
			frac = 0
		}
		if frac > 1 {
			frac = 1
		}
	}
	m.prev = cur
	return frac
}

// readCPUCounters parses the aggregate "cpu" line from /proc/stat.
// Fields: user nice system idle iowait irq softirq steal guest guest_nice.
func readCPUCounters() cpuSnapshot {
	snap := cpuSnapshot{at: time.Now()}
	f, err := os.Open("/proc/stat")
	if err != nil {
		return snap
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)[1:] // skip "cpu"
		var vals []uint64
		for _, f := range fields {
			v, _ := strconv.ParseUint(f, 10, 64)
			vals = append(vals, v)
		}
		if len(vals) < 4 {
			break
		}
		var total uint64
		for _, v := range vals {
			total += v
		}
		// idle = idle + iowait (index 3 and 4).
		idle := vals[3]
		if len(vals) > 4 {
			idle += vals[4]
		}
		snap.idle, snap.total = idle, total
		break
	}
	return snap
}

// GetRouterIP returns the router's LAN IPv4 address (br-lan preferred,
// falls back to the first non-loopback private IPv4).
func GetRouterIP() string {
	// Try br-lan first (OpenWrt default LAN interface).
	if ip := ifaceIPv4("br-lan"); ip != "" {
		return ip
	}
	// Fallback: first non-loopback IPv4.
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if v4 := ip.To4(); v4 != nil {
				return v4.String()
			}
		}
	}
	return ""
}

// ifaceIPv4 returns the first IPv4 address of the named interface.
func ifaceIPv4(name string) string {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return ""
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if v4 := ip.To4(); v4 != nil && !v4.IsLoopback() {
			return v4.String()
		}
	}
	return ""
}

// IsNikkiRunning reports whether nikki (mihomo) is running by scanning
// /proc for a process whose cmdline contains both "mihomo" and "nikki".
// This matches the nikki daemon: /usr/bin/mihomo -d /etc/nikki/run.
func IsNikkiRunning() bool {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Only numeric PIDs.
		pid := e.Name()
		isNum := true
		for _, c := range pid {
			if c < '0' || c > '9' {
				isNum = false
				break
			}
		}
		if !isNum {
			continue
		}
		data, err := os.ReadFile("/proc/" + pid + "/cmdline")
		if err != nil {
			continue
		}
		// cmdline is NUL-separated.
		cmd := strings.ReplaceAll(string(data), "\x00", " ")
		if strings.Contains(cmd, "mihomo") && strings.Contains(cmd, "nikki") {
			return true
		}
	}
	return false
}

// ReadCPUTemp returns the CPU temperature in Celsius.
// Reads /sys/class/thermal/thermal_zone0 (cpu-thermal), temp in millidegrees.
func ReadCPUTemp() float64 {
	data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0
	}
	milli, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return float64(milli) / 1000.0
}

// FanState holds the pwm-fan cooling device state.
type FanState struct {
	Cur int // current state (0 = off/min)
	Max int // max state
}

// ReadFanState reads /sys/class/thermal/cooling_device0 (pwm-fan).
// Returns Cur=-1 if the device is absent.
func ReadFanState() FanState {
	fs := FanState{Cur: -1, Max: 0}
	cur, err := os.ReadFile("/sys/class/thermal/cooling_device0/cur_state")
	if err != nil {
		return fs
	}
	max, err := os.ReadFile("/sys/class/thermal/cooling_device0/max_state")
	if err != nil {
		return fs
	}
	c, _ := strconv.Atoi(strings.TrimSpace(string(cur)))
	m, _ := strconv.Atoi(strings.TrimSpace(string(max)))
	fs.Cur, fs.Max = c, m
	return fs
}
