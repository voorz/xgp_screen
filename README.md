# xgp_screen

Real-time system dashboard for the XGP v3 GC9307 TFT display (320×240),
packaged as LuCI app + Go daemon `.ipk` files.

![dashboard](dash.png)

## Features

- RAM/CPU usage bars
- Up/down throughput chart (live)
- Internet connectivity icons (Baidu/Google)
- Router LAN IP display
- 1 Hz refresh, procd-managed daemon

## Install

One-line install:

```sh
curl -fsSL https://raw.githubusercontent.com/voorz/xgp_screen/main/install.sh | sh
```

Or manually:

```sh
cd /tmp
wget -O core.ipk https://github.com/voorz/xgp_screen/releases/download/v0.1.0/xgp-screen-core_0.1.0-13_aarch64_generic.ipk
wget -O luci.ipk https://github.com/voorz/xgp_screen/releases/download/v0.1.0/luci-app-xgp-screen_0.1.0-13_all.ipk
opkg install core.ipk luci.ipk
```

## License

See [LICENSE](LICENSE).
