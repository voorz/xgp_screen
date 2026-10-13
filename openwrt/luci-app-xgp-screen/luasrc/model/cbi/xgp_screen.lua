m = Map("xgp_screen", translate("XGP Screen"),
	translate("XGP v3 GC9307 屏幕显示控制。渲染核心为 /usr/bin/xgp-screen（Go，二进制），"..
	"直接写 /dev/fb0（240x320 RGB565），不需要 X 或 LVGL。"))

s = m:section(TypedSection, "config", translate("设置"))
s.anonymous = true
s.addremove = false

s:option(Flag, "enabled", translate("启用")).default = "1"

local d = s:option(DummyValue, "_test", translate("屏幕测试"))
d.rawhtml = true
d.value = string.format(
	'<a href="%s" target="_blank">%s</a>',
	luci.dispatcher.build_url("admin", "services", "xgp_screen", "test"),
	translate("在屏幕上显示 HELLO WORLD 测试画面"))

return m
