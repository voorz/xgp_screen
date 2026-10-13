package screen

// Linux framebuffer backend: opens /dev/fb0 (or any fb device), queries
// its geometry via ioctl, and mmaps the video memory into a Pixmap.
//
// Contract with the GC9307 fbtft driver (240x320, RGB565):
//   - the driver exposes the standard framebuffer interface only;
//   - userspace draws into the mmap'd region; fbtft's deferred-I/O
//     worker pushes dirty pages to the panel over SPI by itself.
// No SPI timing or panel init is done here — that is the driver's job.
import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	fbIOGetVScreenInfo = 0x4600 // FBIOGET_VSCREENINFO
	fbIOGetFScreenInfo = 0x4602 // FBIOGET_FSCREENINFO
)

// FBInfo describes the framebuffer geometry we will draw into.
type FBInfo struct {
	XRes, YRes int
	BPP        int
	LineLength int // bytes per row
	SmemLen    int // total mapped bytes
}

// FB is an open, mmap'd framebuffer device.
type FB struct {
	f    *os.File
	data []byte // mmap'd video memory
	Pix  *Pixmap
	Info FBInfo
}

// Open maps the framebuffer device and returns it ready for drawing.
func Open(path string) (*FB, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	ok := true
	defer func() {
		if !ok {
			f.Close()
		}
	}()

	var vinfo [160]byte
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(),
		fbIOGetVScreenInfo, uintptr(unsafe.Pointer(&vinfo[0]))); errno != 0 {
		return nil, fmt.Errorf("FBIOGET_VSCREENINFO: %v", errno)
	}
	var finfo [128]byte
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(),
		fbIOGetFScreenInfo, uintptr(unsafe.Pointer(&finfo[0]))); errno != 0 {
		return nil, fmt.Errorf("FBIOGET_FSCREENINFO: %v", errno)
	}

	info := FBInfo{
		XRes:       int(binary.LittleEndian.Uint32(vinfo[0:4])),
		YRes:       int(binary.LittleEndian.Uint32(vinfo[4:8])),
		BPP:        int(binary.LittleEndian.Uint32(vinfo[24:28])),
		LineLength: int(binary.LittleEndian.Uint32(finfo[48:52])),
		SmemLen:    int(binary.LittleEndian.Uint32(finfo[24:28])),
	}
	if info.XRes <= 0 || info.YRes <= 0 || info.BPP != 16 ||
		info.LineLength <= 0 || info.SmemLen <= 0 {
		return nil, fmt.Errorf("unsupported geometry %+v (need RGB565)", info)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, info.SmemLen,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap: %w", err)
	}

	// Alias the mmap'd bytes as little-endian uint16 pixels (ARM is LE).
	n := len(data) / 2
	pix := unsafe.Slice((*uint16)(unsafe.Pointer(&data[0])), n)
	stride := info.LineLength / 2

	ok = false
	return &FB{
		f:    f,
		data: data,
		Pix:  &Pixmap{W: info.XRes, H: info.YRes, Stride: stride, Pix: pix},
		Info: info,
	}, nil
}

// Unblank clears a kernel-blanked framebuffer (best effort).
func Unblank() error {
	err := os.WriteFile("/sys/class/graphics/fb0/blank", []byte("0"), 0644)
	if err != nil {
		return fmt.Errorf("unblank fb0: %w", err)
	}
	return nil
}

// Close unmaps and closes the device. The last drawn frame stays on the
// panel — the driver keeps displaying video memory after we exit.
func (fb *FB) Close() error {
	merr := syscall.Munmap(fb.data)
	cerr := fb.f.Close()
	if merr != nil {
		return merr
	}
	return cerr
}
