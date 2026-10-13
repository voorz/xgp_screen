# xgp_screen

A LuCI application for the XGP v3 GC9307 TFT display. Provides a real-time
system dashboard rendered natively to the framebuffer, packaged as
independently installable `.ipk` files for fast iteration without rebuilding
firmware.

## Architecture

```
┌─────────────────────────────────────────────────┐
│  LuCI Web UI (luci-app-xgp-screen)              │
│  JavaScript view + menu.d entry                 │
└──────────────────────┬──────────────────────────┘
                       │ UCI config
┌──────────────────────▼──────────────────────────┐
│  Go Daemon (xgp-screen-core)                    │
│  ┌─────────────┐  ┌──────────────────────────┐  │
│  │ collector   │  │ screen (framebuffer)     │  │
│  │ - /proc     │→ │ - 320×240 RGB565        │  │
│  │ - net rates │  │ - TTF text (DejaVu Sans)│  │
│  │ - probes    │  │ - charts, bars, icons   │  │
│  └─────────────┘  └──────────────────────────┘  │
│  procd service, respawn, syslog                 │
└─────────────────────────────────────────────────┘
```

- **Device rendering**: Go draws directly to `/dev/fb0` (RGB565). No
  browser, CSS, or HTML on the device — everything is pixel-rendered.
- **Web preview**: The LuCI page provides configuration; a browser-based
  WYSIWYG preview uses standard CSS/charts aligned to the same layout spec.
- **Layout spec**: A shared JSON/widget schema drives both the device
  renderer and the web preview, enabling flexible configuration.

## Features

- **Single-page dashboard** (320×240 landscape, black background):
  - RAM/CPU thin progress bars (name left, percentage right)
  - Merged up/down throughput chart (dual-line, grid, area fill)
  - Connectivity icons (Baidu/Google; color when reachable, grayscale when not)
  - Router LAN IP (bottom-right)
- **Real-time data** (1 Hz refresh):
  - Memory from `/proc/meminfo`
  - CPU from `/proc/stat`
  - Network rates from `/proc/net/dev` (all non-loopback interfaces)
  - HTTP connectivity probes (Baidu for internet, Google for VPN environment)
- **Procd service**: auto-start on boot, respawn on crash, logs to syslog.
- **Vector fonts**: Embedded DejaVu Sans with anti-aliasing (no bitmap fonts).
- **SVG icons**: Brand icons rasterized at build time, embedded in the binary.

## Hardware

- Panel: Galaxycore GC9307, 240×320 portrait
- Driver: `kmod-fb-tft-gc9307` with `rotate=270` → 320×240 landscape framebuffer
- Interface: `/dev/fb0`, 16bpp RGB565

The driver exposes only the raw framebuffer. Fonts, logos, and all display
content are provided by this package.

## Installation

Download from the
[releases page](https://github.com/voorz/xgp_screen/releases):

```sh
cd /tmp
wget -O xgp-screen-core_0.1.0-13_aarch64_generic.ipk \
  https://github.com/voorz/xgp_screen/releases/download/v0.1.0/xgp-screen-core_0.1.0-13_aarch64_generic.ipk
wget -O luci-app-xgp-screen_0.1.0-13_all.ipk \
  https://github.com/voorz/xgp_screen/releases/download/v0.1.0/luci-app-xgp-screen_0.1.0-13_all.ipk
opkg install xgp-screen-core_0.1.0-13_aarch64_generic.ipk \
  luci-app-xgp-screen_0.1.0-13_all.ipk
```

The daemon starts automatically via procd. Verify:

```sh
ps | grep xgp-screen | grep -v grep
logread | grep xgp-screen
```

## Packages

| Package | Description |
|---------|-------------|
| `xgp-screen-core` | Go daemon binary + procd init script |
| `luci-app-xgp-screen` | LuCI JS view, menu entry, rpcd ACL |

Built with `tools/mkipk.sh`:

```sh
./tools/mkipk.sh 0.1.0 13
```

Produces OpenWrt-compatible `.ipk` files (gzip tar with `debian-binary`,
`data.tar.gz`, `control.tar.gz` in order; root:root, 644/755).

## Development

Requirements: Go 1.21+.

```sh
# Verify
gofmt -l .
go vet ./...
go build ./...

# Preview the dashboard (writes dash.png)
go run ./tools/preview

# Cross-compile for the router (ARM64, static)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o xgp-screen ./cmd/xgp-screen

# Run once (useful over SSH)
./xgp-screen -once
```

### Project layout

```
cmd/xgp-screen/    Daemon entrypoint (flags: -once, -dev, -interval)
collector/         /proc parsers, rate calculation, connectivity probes
screen/            Framebuffer backend, Pixmap, widgets, dashboard layout
  icons/           Embedded PNG icons (generated from SVG)
  DejaVuSans.ttf   Embedded vector font
openwrt/           Packaging: init script, LuCI files, control templates
tools/
  mkipk.sh         .ipk builder
  preview/         Dashboard PNG renderer (no hardware needed)
  svg2png/         SVG → PNG converter (for icons)
```

## Design principles

- **One page, no interaction**: The panel has no touch input; the dashboard
  shows everything at a glance, inspired by sports watches and commercial
  routers.
- **Content over chrome**: Throughput charts are the hero; time is
  deliberately inconspicuous (replaced by the router IP).
- **Black background**: Reduces eye strain and matches the panel's aesthetics.
- **Fast iteration**: The `.ipk` installs in seconds; no firmware rebuild needed.

## License

See [LICENSE](LICENSE).
