#!/bin/sh
# Fast-iteration ipk builder for xgp_screen.
#
# A full firmware image build takes ~2h; these two packages are a static
# Go binary plus Lua/config files, so the ipks are assembled directly
# (an ipk is just an ar archive: debian-binary + control.tar.gz +
# data.tar.gz). No OpenWrt buildroot/SDK needed.
#
# Usage:
#   ./tools/mkipk.sh [version] [release]
#   ./tools/mkipk.sh 0.1.0 2        # -> dist/*0.1.0-2*.ipk
#
# Then on the device:
#   scp dist/*.ipk root@<dev>:/tmp/
#   ssh root@<dev> 'opkg install /tmp/xgp-screen-core_*.ipk /tmp/luci-app-xgp-screen_*.ipk'
set -eu

VERSION="${1:-0.1.0}"
RELEASE="${2:-1}"
ARCH="aarch64_generic"   # rockchip/armv8 generic; adjust if your target differs

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="$ROOT/dist"
STAGE="$ROOT/.ipk-stage"

command -v ar >/dev/null || { echo "need 'ar' (binutils)" >&2; exit 1; }
command -v go >/dev/null || { echo "need 'go' on PATH" >&2; exit 1; }

rm -rf "$DIST" "$STAGE"
mkdir -p "$DIST" "$STAGE"

echo "==> building Go binary (linux/arm64, static)"
(
	cd "$ROOT"
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -trimpath -ldflags="-s -w" -o "$STAGE/xgp-screen" ./cmd/xgp-screen
)
file "$STAGE/xgp-screen" | grep -q "ARM aarch64" || { echo "not an aarch64 binary?" >&2; exit 1; }

# mkipk <name> <version> <arch> <stagedir> <description> [depends]
# stagedir holds the data payload; stagedir.postinst (optional) becomes postinst.
mkipk() {
	name="$1"; ver="$2"; arch="$3"; src="$4"; desc="$5"; deps="${6:-}"
	pkgdir="$STAGE/pkg-$name"
	ctrl="$pkgdir/CONTROL"
	data="$pkgdir/DATA"
	mkdir -p "$ctrl" "$data"
	cp -r "$src"/. "$data"/
	rm -f "$data/postinst"   # control script must not land in data
	{
		echo "Package: $name"
		echo "Version: $ver"
		echo "Architecture: $arch"
		echo "Maintainer: voorz"
		echo "Source: https://github.com/voorz/xgp_screen"
		echo "Description: $desc"
		[ -n "$deps" ] && echo "Depends: $deps"
	} > "$ctrl/control"
	files="./control"
	if [ -f "$src.postinst" ]; then
		cp "$src.postinst" "$ctrl/postinst"
		chmod 755 "$ctrl/postinst"
		files="$files ./postinst"
	fi
	# shellcheck disable=SC2086
	( cd "$ctrl" && tar --owner=0 --group=0 -czf "$STAGE/control.tar.gz" $files )
	( cd "$data" && tar --owner=0 --group=0 -czf "$STAGE/data.tar.gz" . )
	printf '2.0\n' > "$STAGE/debian-binary"
	( cd "$STAGE" && ar rcs "$DIST/${name}_${ver}_${arch}.ipk" debian-binary control.tar.gz data.tar.gz )
	rm -rf "$pkgdir" "$STAGE/control.tar.gz" "$STAGE/data.tar.gz" "$STAGE/debian-binary"
	echo "==> ${name}_${ver}_${arch}.ipk"
}

# --- xgp-screen-core: the Go renderer ---
S1="$STAGE/s1"; mkdir -p "$S1/usr/bin"
cp "$STAGE/xgp-screen" "$S1/usr/bin/"
mkipk "xgp-screen-core" "$VERSION-$RELEASE" "$ARCH" "$S1" \
	"Go framebuffer renderer for the XGP v3 GC9307 panel (240x320 RGB565)."

# --- luci-app-xgp-screen: the LuCI frontend (arch-independent) ---
S2="$STAGE/s2"
mkdir -p "$S2/usr/lib/lua/luci/controller" \
	"$S2/usr/lib/lua/luci/model/cbi" \
	"$S2/etc/config"
cp "$ROOT/openwrt/luci-app-xgp-screen/luasrc/controller/xgp_screen.lua" \
	"$S2/usr/lib/lua/luci/controller/"
cp "$ROOT/openwrt/luci-app-xgp-screen/luasrc/model/cbi/xgp_screen.lua" \
	"$S2/usr/lib/lua/luci/model/cbi/"
cp "$ROOT/openwrt/luci-app-xgp-screen/root/etc/config/xgp_screen" \
	"$S2/etc/config/"
# clear LuCI caches after install so the new app shows up immediately
cat > "$S2.postinst" <<'EOF'
#!/bin/sh
rm -rf /tmp/luci-indexcache /tmp/luci-modulecache
exit 0
EOF
mkipk "luci-app-xgp-screen" "$VERSION-$RELEASE" "all" "$S2" \
	"LuCI support for XGP v3 screen." "luci-base"

rm -rf "$STAGE"
echo
echo "done. install with:"
echo "  scp $DIST/*.ipk root@<dev>:/tmp/"
echo "  ssh root@<dev> 'opkg install /tmp/xgp-screen-core_*.ipk /tmp/luci-app-xgp-screen_*.ipk && /usr/bin/xgp-screen'"
