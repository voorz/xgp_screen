#!/bin/sh
# Fast-iteration ipk builder for xgp_screen.
#
# A full firmware image build takes ~2h; these two packages are a static
# Go binary plus Lua/config files, so the ipks are assembled directly.
# IMPORTANT: OpenWrt .ipk is a gzipped TAR (not an ar archive!):
#   tar ./debian-binary ./data.tar.gz ./control.tar.gz | gzip
# (see OpenWrt's scripts/ipkg-build). No OpenWrt buildroot/SDK needed.
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
	rm -f "$data/postinst" "$data/prerm"   # control scripts must not land in data
	# Normalize like official ipkg-build: root:root, 644 files / 755 dirs,
	# executables stay executable. (Without this, the builder's uid/gid and
	# umask leak into the package — e.g. 660 root:nogroup — and non-root
	# readers like uhttpd may fail to read the files.)
	find "$data" -type d -exec chmod 755 {} +
	find "$data" -type f -exec chmod 644 {} +
	[ -f "$data/usr/bin/xgp-screen" ] && chmod 755 "$data/usr/bin/xgp-screen"
	[ -f "$data/bin/set-xgp-screen-brightness" ] && chmod 755 "$data/bin/set-xgp-screen-brightness"
	find "$data/etc/init.d" -type f -exec chmod 755 {} + 2>/dev/null || true
	{
		echo "Package: $name"
		echo "Version: $ver"
		echo "Architecture: $arch"
		echo "Maintainer: voorz"
		echo "Source: https://github.com/voorz/xgp_screen"
		echo "Description: $desc"
		[ -n "$deps" ] && echo "Depends: $deps"
	} > "$ctrl/control"
	# shellcheck disable=SC2086
	( cd "$data" && tar --format=gnu --numeric-owner --owner=0 --group=0 --sort=name -cf - . | gzip -n > "$STAGE/data.tar.gz" )
	# Installed-Size, like official ipkg-build
	installed_size=$(zcat < "$STAGE/data.tar.gz" | wc -c)
	echo "Installed-Size: $installed_size" >> "$ctrl/control"
	files="./control"
	if [ -f "$src.postinst" ]; then
		cp "$src.postinst" "$ctrl/postinst"
		chmod 755 "$ctrl/postinst"
		files="$files ./postinst"
	fi
	if [ -f "$src.prerm" ]; then
		cp "$src.prerm" "$ctrl/prerm"
		chmod 755 "$ctrl/prerm"
		files="$files ./prerm"
	fi
	# shellcheck disable=SC2086
	( cd "$ctrl" && tar --format=gnu --numeric-owner --owner=0 --group=0 --sort=name -cf - $files | gzip -n > "$STAGE/control.tar.gz" )
	printf '2.0\n' > "$STAGE/debian-binary"
	# Official member order: debian-binary, data.tar.gz, control.tar.gz,
	# whole thing gzipped (NOT an ar archive).
	( cd "$STAGE" && tar --format=gnu --numeric-owner --owner=0 --group=0 --sort=name -cf - \
		./debian-binary ./data.tar.gz ./control.tar.gz | gzip -n > \
		"$DIST/${name}_${ver}_${arch}.ipk" )
	rm -rf "$pkgdir" "$STAGE/control.tar.gz" "$STAGE/data.tar.gz" "$STAGE/debian-binary"
	echo "==> ${name}_${ver}_${arch}.ipk"
}

# --- xgp-screen-core: the Go renderer ---
S1="$STAGE/s1"; mkdir -p "$S1/usr/bin" "$S1/etc/init.d"
cp "$STAGE/xgp-screen" "$S1/usr/bin/"
cp "$ROOT/openwrt/xgp-screen-core/root/etc/init.d/xgp-screen" "$S1/etc/init.d/"
# postinst: enable and start the daemon (fresh install only; on upgrade the
# service is already enabled and procd restarts it).
cat > "$S1.postinst" <<'EOF'
#!/bin/sh
[ -n "${IPKG_INSTROOT}" ] || {
/etc/init.d/xgp-screen enable
/etc/init.d/xgp-screen restart
}
exit 0
EOF
# prerm: stop the daemon on remove.
cat > "$S1.prerm" <<'EOF'
#!/bin/sh
[ -n "${IPKG_INSTROOT}" ] || {
/etc/init.d/xgp-screen stop
/etc/init.d/xgp-screen disable
}
exit 0
EOF
mkipk "xgp-screen-core" "$VERSION-$RELEASE" "$ARCH" "$S1" \
	"Go framebuffer renderer for the XGP v3 GC9307 panel (240x320 RGB565)."

# --- luci-app-xgp-screen: the LuCI frontend (arch-independent) ---
# Native JS view + menu.d (no Lua controller): matches the ucode-based LuCI,
# sidesteps the Lua compatibility layer entirely.
S2="$STAGE/s2"
mkdir -p "$S2/www/luci-static/resources/view/xgp-screen" \
	"$S2/usr/share/luci/menu.d" \
	"$S2/usr/share/rpcd/acl.d" \
	"$S2/etc/config"
cp "$ROOT/openwrt/luci-app-xgp-screen/htdocs/luci-static/resources/view/xgp-screen/overview.js" \
	"$S2/www/luci-static/resources/view/xgp-screen/"
cp "$ROOT/openwrt/luci-app-xgp-screen/root/usr/share/luci/menu.d/luci-app-xgp-screen.json" \
	"$S2/usr/share/luci/menu.d/"
cp "$ROOT/openwrt/luci-app-xgp-screen/root/usr/share/rpcd/acl.d/luci-app-xgp-screen.json" \
	"$S2/usr/share/rpcd/acl.d/"
cp "$ROOT/openwrt/luci-app-xgp-screen/root/etc/config/xgp_screen" \
	"$S2/etc/config/"
# postinst: drop files from the old Lua-controller based package (<= 0.1.0-1)
# and clear LuCI caches so the new view shows up immediately.
cat > "$S2.postinst" <<'EOF'
#!/bin/sh
rm -f /usr/lib/lua/luci/controller/xgp_screen.lua \
      /usr/lib/lua/luci/model/cbi/xgp_screen.lua
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
