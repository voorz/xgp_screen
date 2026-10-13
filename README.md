<div align="center">

# xgp_screen

![Release](https://img.shields.io/badge/release-v0.1.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-OpenWrt-lightgrey)

Real-time system dashboard for the XGP v3 GC9307 TFT display.

</div>

![dashboard](dash.png)

## Architecture

```
LuCI Web UI (JS view)
        |
        v (UCI)
Go Daemon (procd)
  collector -> screen -> /dev/fb0 320x240
```

## Features

- RAM/CPU usage bars
- Live up/down throughput chart
- Connectivity icons (Baidu/Google)
- Router LAN IP display
- 1 Hz refresh, auto-start on boot

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/voorz/xgp_screen/main/install.sh | sh
```

## License

See LICENSE file.
