#!/bin/sh
# xgp_screen one-line installer for OpenWrt/ImmortalWrt.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/voorz/xgp_screen/main/install.sh | sh
#
# Environment:
#   VERSION  Specific version to install (e.g. "0.1.0-13"), default: latest
#   REPO     GitHub repo (default: voorz/xgp_screen)
set -eu

REPO="${REPO:-voorz/xgp_screen}"
VERSION="${VERSION:-latest}"

log() { printf '[xgp-install] %s\n' "$*"; }
err() { printf '[xgp-install] ERROR: %s\n' "$*" >&2; }

# Must run on OpenWrt.
if [ ! -f /etc/openwrt_release ]; then
  err "This script is for OpenWrt/ImmortalWrt only."
  exit 1
fi

# Must be root.
if [ "$(id -u)" -ne 0 ]; then
  err "Please run as root."
  exit 1
fi

# Remove legacy packages from the original author (conflict with the new Go version).
# Legacy init script name is zz_xgp_screen, package name is xgp-v3-screen.
if [ -x /etc/init.d/zz_xgp_screen ]; then
  log "Stopping legacy daemon..."
  /etc/init.d/zz_xgp_screen stop 2>/dev/null || true
  /etc/init.d/zz_xgp_screen disable 2>/dev/null || true
fi
for pkg in xgp-v3-screen luci-app-xgp-v3-screen-brightness; do
  if opkg list-installed 2>/dev/null | grep -q "^$pkg "; then
    log "Removing legacy package: $pkg..."
    opkg remove "$pkg" || true
  fi
done

# Download tool.
if command -v curl >/dev/null 2>&1; then
  DL="curl -fsSL -o"
elif command -v wget >/dev/null 2>&1; then
  DL="wget -O"
else
  err "Need curl or wget."
  exit 1
fi

# Resolve version.
if [ "$VERSION" = "latest" ]; then
  log "Resolving latest version..."
  # The release tag is v0.1.0; assets carry the PKG_RELEASE suffix.
  TAG="v0.1.0"
else
  # VERSION like "0.1.0-13" -> tag v0.1.0
  TAG="v$(echo "$VERSION" | cut -d- -f1)"
fi

# Find the newest PKG_RELEASE if not specified.
if [ "$VERSION" = "latest" ]; then
  log "Fetching release assets for $TAG..."
  API="https://api.github.com/repos/$REPO/releases/tags/$TAG"
  if command -v curl >/dev/null 2>&1; then
    ASSETS=$(curl -fsSL "$API" | grep '"name"' | cut -d'"' -f4)
  else
    ASSETS=$(wget -qO- "$API" | grep '"name"' | cut -d'"' -f4)
  fi
  CORE=$(echo "$ASSETS" | grep '^xgp-screen-core_' | sort -V | tail -1)
  LUCI=$(echo "$ASSETS" | grep '^luci-app-xgp-screen_' | sort -V | tail -1)
else
  CORE="xgp-screen-core_${VERSION}_aarch64_generic.ipk"
  LUCI="luci-app-xgp-screen_${VERSION}_all.ipk"
fi

if [ -z "$CORE" ] || [ -z "$LUCI" ]; then
  err "Could not find packages for version $VERSION."
  exit 1
fi

BASE="https://github.com/$REPO/releases/download/$TAG"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

log "Downloading $CORE..."
$DL "$TMPDIR/$CORE" "$BASE/$CORE"
log "Downloading $LUCI..."
$DL "$TMPDIR/$LUCI" "$BASE/$LUCI"

log "Installing..."
opkg install "$TMPDIR/$CORE" "$TMPDIR/$LUCI"

log "Verifying..."
if /etc/init.d/xgp-screen enabled; then
  log "Service enabled."
else
  /etc/init.d/xgp-screen enable
fi
/etc/init.d/xgp-screen restart

if pgrep -f "xgp-screen" >/dev/null 2>&1; then
  log "xgp-screen is running. Done."
else
  err "Daemon did not start. Check: logread | grep xgp-screen"
  exit 1
fi
