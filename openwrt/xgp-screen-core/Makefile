include $(TOPDIR)/rules.mk

PKG_NAME:=xgp-screen-core
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

PKG_SOURCE_PROTO:=git
PKG_SOURCE_URL:=https://github.com/voorz/xgp_screen.git
PKG_SOURCE_VERSION:=v0.1.0
PKG_MIRROR_HASH:=skip

PKG_LICENSE:=MIT
PKG_LICENSE_FILES:=LICENSE
PKG_MAINTAINER:=voorz

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/xgp-screen-core
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=XGP v3 screen renderer core (Go)
  DEPENDS:=$(GO_ARCH_DEPENDS)
endef

define Package/xgp-screen-core/description
  Go framebuffer renderer for the XGP v3 GC9307 panel (240x320, RGB565).
  Draws directly to /dev/fb0; no X, no LVGL needed.
endef

GO_PKG:=github.com/voorz/xgp_screen/cmd/xgp-screen

define Package/xgp-screen-core/install
	$(call GoPackage/Package/Install/Bin,$(PKG_INSTALL_DIR))
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_INSTALL_DIR)/usr/bin/xgp-screen $(1)/usr/bin/
endef

$(eval $(call GoBinPackage,xgp-screen-core))
$(eval $(call BuildPackage,xgp-screen-core))
