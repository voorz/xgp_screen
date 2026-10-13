# xgp_screen

XGP v3 屏幕显示方案：固件只负责屏幕驱动（`kmod-fb-tft-gc9307`，`/dev/fb0`，
240x320 RGB565），显示内容全部在用户态完成——Go 二进制渲染核心 +
luci-app 后台编排。驱动和内容之间没有私有协议，只有标准 framebuffer 接口。

```
固件层（immortalwrt-xgpv3）        本仓库
┌─────────────────────┐      ┌──────────────────────────────┐
│ kmod-fb-tft-gc9307  │      │ cmd/xgp-screen  (Go, 静态)   │──▶ /dev/fb0
│ /dev/fb0 240x320    │◀─────│ screen/         (pixmap+字体) │
│ RGB565, SPI 推送     │ mmap │ openwrt/luci-app-xgp-screen  │──▶ 后台编排
└─────────────────────┘      └──────────────────────────────┘
```

## 里程碑 1：HELLO WORLD

`cmd/xgp-screen` 在屏幕上画出 HELLO WORLD + 分辨率自检行（`240X320` /
`BPP16`，取自真实 ioctl，不是写死的），画完即退出，画面保留。

渲染代码与硬件解耦（`screen.Pixmap`），可用 `tools/preview` 在无硬件
环境验证：`go run ./tools/preview` 生成 `hello.png`。

## 设备端测试（不需要重新编译固件）

### Level 0：驱动是否正常（30 秒）

```sh
ls -l /dev/fb0
cat /sys/class/graphics/fb0/virtual_size   # 期望 240,320
echo 0 > /sys/class/graphics/fb0/blank     # 解除 blank
# 整屏刷红，验证驱动+排线+背光通路：
python3 -c "import sys; sys.stdout.buffer.write(b'\x00\xf8'*76800)" > /dev/fb0
```

屏幕变红 ⇒ 驱动层 OK，问题一定在应用层。屏幕没反应 ⇒ 先查
`dmesg | grep -i gc9307` 和背光（`set-xgp-screen-brightness`）。

### Level 1：跑 Go 二进制

```sh
# 在开发机上交叉编译（本仓库根目录）：
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o xgp-screen ./cmd/xgp-screen
scp xgp-screen root@<设备IP>:/tmp/
ssh root@<设备IP> '/tmp/xgp-screen'
```

期望输出类似：

```
OK fb /dev/fb0: 240x320 bpp=16 line=480 smem=153600
OK unblank
OK draw: HELLO WORLD should now be on the panel
```

此时屏幕应显示：深蓝底、白色 HELLO / WORLD、绿色 `240X320` / `BPP16`、
白色边框。

### 如果还是不显示，把下面贴回来

1. 上面命令的完整输出；
2. `cat /sys/class/graphics/fb0/{virtual_size,bits_per_pixel,blank}`；
3. `dmesg | grep -i -E 'gc9307|fb0' | tail -5`；
4. Level 0 刷红那步屏幕有没有变红；
5. 背光状态（`cat /sys/class/backlight/*/brightness` 或 brightness 脚本）。

## 排障清单（hello world 不显示的常见原因）

| 现象 | 原因 | 本仓库如何规避 |
|---|---|---|
| 花屏/偏色 | 按 RGB888 或 32bpp 写进了 RGB565 fb | ioctl 实测 bpp，非 16 直接报错退出 |
| 全黑 | fb 被 blank | 程序内写 sysfs unblank |
| 全黑 | 背光没开 | Level 0 先验证背光通路 |
| 图像撕裂/错位 | 硬编码分辨率、忽略 line_length stride | 全部来自 ioctl |
| `cannot execute binary file` | 二进制架构不对 / CGO 动态链接 | `GOARCH=arm64 CGO_ENABLED=0` 纯静态 |
| 无字只有底色 | 字体数据缺失 | 内嵌手绘 5x7 点阵字体，零依赖 |
| 程序静默失败 | open/mmap 没判错 | 每一步都打印 OK/WARN/FAIL |

## 固件集成（后面做）

`openwrt/` 下有两个包：`xgp-screen-core`（Go 二进制，
`golang-package.mk`）和 `luci-app-xgp-screen`（`luci.mk`，依赖前者）。
集成到 immortalwrt-xgpv3 构建时，把这两个目录放进 `zz-packages/` 即可；
luci-app 在 `服务 → XGP Screen`，有点开即跑的屏幕测试入口。

## 路线图

- M1（本提交）：framebuffer hello world + 自检 ✓
- M2：JSON 布局规约 + widget 白名单（信号/时钟/流量/二维码）
- M3：Go 守护进程（文件监听/ubus 触发重绘）+ 数据采集（复用 modem_info.py 思路）
- M4：luci-app 可视化编排 + 浏览器实时预览（HTML/CSS 只做预览皮，设备端原生渲染）
