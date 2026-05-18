# chromesnap

基于 [chromedp](https://github.com/chromedp/chromedp) 构建的 headless Chrome 截图工具，同时提供 Go 开发库。

## 环境要求

需要在系统中安装 Chrome 或 Chromium。chromesnap 会自动查找，也可以通过 `--chrome` 手动指定路径。

## 安装

```bash
go install github.com/normal-coder/chromesnap/cmd/chromesnap@latest
```

## CLI 使用

### 单页截图

```bash
# 基础用法
chromesnap https://example.com

# 全页截图
chromesnap https://example.com -F -o full.png

# 移动设备模拟
chromesnap https://example.com -e iPhone-15 -o mobile.png

# 暗色模式，输出 JPEG
chromesnap https://example.com -D -f jpeg -o dark.jpg

# 截取特定元素
chromesnap https://example.com -s "#hero" -o hero.png

# 等待元素出现后再延迟截图
chromesnap https://example.com -w "#app" -d 1s -o app.png

# 截图前注入 JS 和 CSS
chromesnap https://example.com --js "document.querySelector('.banner').remove()" --css "body{zoom:0.8}"

# 输出到 stdout，管道传递给其他命令
chromesnap https://example.com --stdout | convert - result.webp
```

### 批量截图

```bash
# 从文件读取 URL（每行一个，# 开头为注释）
chromesnap batch urls.txt -o ./shots/

# 直接传入 URL 列表
chromesnap batch --urls "https://a.com,https://b.com" -o ./shots/

# 自定义并发数，失败继续，输出 JSON 汇总
chromesnap batch urls.txt -o ./shots/ -p 5 --continue-on-error -j

# 自定义文件命名，支持 {index}、{host}、{ts}
chromesnap batch urls.txt -o ./shots/ -n "{index}_{host}"
```

### 全局参数

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--chrome` | | 自动查找 | Chrome/Chromium 可执行路径 |
| `--remote` | | | 远程 CDP 地址 `ws://localhost:9222` |
| `--no-headless` | | false | 关闭无头模式（调试用） |
| `--proxy` | | | 代理服务器地址 |
| `--ignore-cert-errors` | | false | 忽略 TLS 证书错误 |
| `--timeout` | `-t` | `30s` | 单页超时时间 |
| `--quiet` | `-q` | false | 静默模式，不输出日志 |

### 截图参数

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--output` | `-o` | `screenshot.png` | 输出文件路径 |
| `--format` | `-f` | `png` | 输出格式：`png` / `jpeg` / `webp` / `pdf` |
| `--quality` | | `85` | JPEG/WebP 质量（1–100） |
| `--stdout` | | false | 输出到 stdout |
| `--json` | `-j` | false | 将 JSON 元数据输出到 stderr |
| `--width` | `-W` | `1920` | 视口宽度 |
| `--height` | `-H` | `1080` | 视口高度 |
| `--device` | `-e` | | 设备预设 |
| `--dpr` | | `1.0` | Device Pixel Ratio |
| `--dark-mode` | `-D` | false | 启用 `prefers-color-scheme: dark` |
| `--full-page` | `-F` | false | 截取完整页面（含滚动区域） |
| `--selector` | `-s` | | 截取指定 CSS 选择器元素 |
| `--clip` | | | 截取坐标区域 `x,y,width,height` |
| `--wait-for` | `-w` | | 等待 CSS 选择器出现后截图 |
| `--wait-network` | | false | 等待网络空闲后截图 |
| `--delay` | `-d` | | 页面加载后额外等待时间 |
| `--js` | | | 页面加载后执行的 JavaScript |
| `--css` | | | 注入的 CSS 样式 |
| `--basic-auth` | | | HTTP Basic Auth，格式 `user:pass` |
| `--header` | | | 自定义请求头（可多次使用） |
| `--cookie` | `-c` | | 设置 Cookie（可多次使用） |
| `--user-agent` | `-A` | | 自定义 User-Agent |

### 设备预设列表

| 名称 | 分辨率 | DPR |
|------|--------|-----|
| `iPhone-15` | 390×844 | 3 |
| `iPhone-15-Pro-Max` | 430×932 | 3 |
| `iPad-Pro` | 1024×1366 | 2 |
| `Pixel-7` | 412×915 | 2.625 |
| `Galaxy-S23` | 360×780 | 3 |
| `MacBook-Air` | 1280×800 | 2 |
| `Desktop-1080p` | 1920×1080 | 1 |
| `Desktop-4K` | 3840×2160 | 1 |

### Shell 自动补全

```bash
# Zsh
chromesnap completion zsh > ~/.zsh/completions/_chromesnap

# Bash
chromesnap completion bash > /etc/bash_completion.d/chromesnap
```

## Go 开发库

```bash
go get normalcoder.com/chromesnap
```

### 快速开始

```go
import "normalcoder.com/chromesnap/snap"

// 返回 []byte
buf, err := snap.Capture("https://example.com")

// 直接保存到文件
err := snap.CaptureToFile("https://example.com", "result.png")
```

### 完整参数控制

```go
buf, err := snap.Capture("https://example.com",
    snap.WithViewport(1440, 900),
    snap.WithFullPage(),
    snap.WithFormat(snap.FormatJPEG),
    snap.WithQuality(90),
    snap.WithDarkMode(),
    snap.WithWaitFor("#app"),
    snap.WithDelay(1 * time.Second),
    snap.WithJS("document.querySelector('.cookie-banner').remove()"),
    snap.WithHeader("Authorization", "Bearer TOKEN"),
)
```

### 设备预设

```go
buf, err := snap.Capture("https://example.com",
    snap.WithDevice(snap.DeviceiPhone15),
)
```

### 复用 Browser 实例（批量高性能场景）

```go
browser, err := snap.NewBrowser(
    snap.WithChromePath("/usr/bin/google-chrome"),
)
defer browser.Close()

buf1, err := browser.Capture("https://a.com", snap.WithFullPage())
buf2, err := browser.Capture("https://b.com", snap.WithDevice(snap.DeviceiPhone15))

results, err := browser.CaptureAll([]string{
    "https://a.com",
    "https://b.com",
}, snap.WithConcurrency(5))
```

## License

MIT
