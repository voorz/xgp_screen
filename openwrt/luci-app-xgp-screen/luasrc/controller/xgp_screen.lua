module("luci.controller.xgp_screen", package.seeall)

function index()
	entry({"admin", "services", "xgp_screen"},
		cbi("xgp_screen"), _("XGP Screen"), 60).dependent = false
	entry({"admin", "services", "xgp_screen", "test"},
		call("action_test")).leaf = true
end

-- Runs the Go renderer once and returns its diagnostic output as text.
-- The drawn frame stays on the panel after the process exits.
function action_test()
	local util = require("luci.util")
	luci.http.prepare_content("text/plain; charset=utf-8")
	local out = util.exec("/usr/bin/xgp-screen 2>&1") or "(no output)"
	luci.http.write(out)
end
