include $(TOPDIR)/rules.mk

LUCI_TITLE:=LuCI support for XGP v3 screen
LUCI_DEPENDS:=+xgp-screen-core
LUCI_PKGARCH:=all

define Package/luci-app-xgp-screen/conffiles
/etc/config/xgp_screen
endef

include $(TOPDIR)/feeds/luci/luci.mk

# call BuildPackage - OpenWrt buildroot only
