# chromesnap

Headless Chrome screenshot tool and Go library, built on [chromedp](https://github.com/chromedp/chromedp).

## Requirements

Chrome or Chromium must be installed on your system. chromesnap will find it automatically, or you can specify the path with `--chrome`.

## Installation

```bash
go install github.com/normal-coder/chromesnap/cmd/chromesnap@latest
```

## CLI

### Single page

```bash
# Basic
chromesnap https://example.com

# Full page
chromesnap https://example.com -F -o full.png

# Mobile device
chromesnap https://example.com -e iPhone-15 -o mobile.png

# Dark mode, JPEG
chromesnap https://example.com -D -f jpeg -o dark.jpg

# Capture a specific element
chromesnap https://example.com -s "#hero" -o hero.png

# Wait for element, then delay
chromesnap https://example.com -w "#app" -d 1s -o app.png

# Inject JS and CSS before capture
chromesnap https://example.com --js "document.querySelector('.banner').remove()" --css "body{zoom:0.8}"

# Pipe to another command
chromesnap https://example.com --stdout | convert - result.webp
```

### Batch

```bash
# From a file (one URL per line, # for comments)
chromesnap batch urls.txt -o ./shots/

# From inline list
chromesnap batch --urls "https://a.com,https://b.com" -o ./shots/

# Custom concurrency, continue on error, JSON summary
chromesnap batch urls.txt -o ./shots/ -p 5 --continue-on-error -j

# Custom file naming: {index}, {host}, {ts}
chromesnap batch urls.txt -o ./shots/ -n "{index}_{host}"
```

### Global flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--chrome` | | auto | Chrome/Chromium executable path |
| `--remote` | | | Remote CDP URL `ws://localhost:9222` |
| `--no-headless` | | false | Disable headless mode |
| `--proxy` | | | Proxy server URL |
| `--ignore-cert-errors` | | false | Ignore TLS certificate errors |
| `--timeout` | `-t` | `30s` | Per-page timeout |
| `--quiet` | `-q` | false | Suppress log output |
| `--verbose` | `-v` | | Verbose diagnostics (`-v`: snap events, `-vv`: + chromedp CDP debug) |

### Snap flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `screenshot.png` | Output file path |
| `--format` | `-f` | `png` | `png` / `jpeg` / `webp` / `pdf` |
| `--quality` | | `85` | JPEG/WebP quality (1–100) |
| `--stdout` | | false | Write to stdout |
| `--json` | `-j` | false | Print JSON metadata to stderr |
| `--width` | `-W` | `1920` | Viewport width |
| `--height` | `-H` | `1080` | Viewport height |
| `--device` | `-e` | | Device preset |
| `--dpr` | | `1.0` | Device pixel ratio |
| `--dark-mode` | `-D` | false | `prefers-color-scheme: dark` |
| `--full-page` | `-F` | false | Capture full scrollable page |
| `--selector` | `-s` | | Capture a CSS selector element |
| `--clip` | | | Capture region `x,y,width,height` |
| `--wait-for` | `-w` | | Wait for CSS selector |
| `--wait-network` | | false | Wait for network idle |
| `--delay` | `-d` | | Extra wait after load |
| `--js` | | | Execute JavaScript after load |
| `--css` | | | Inject CSS into page |
| `--basic-auth` | | | HTTP basic auth `user:pass` |
| `--header` | | | Custom request header (repeatable) |
| `--cookie` | `-c` | | Set cookie (repeatable) |
| `--user-agent` | `-A` | | Custom User-Agent |

### Device presets

| Name | Resolution | DPR |
|------|-----------|-----|
| `iPhone-15` | 390×844 | 3 |
| `iPhone-15-Pro-Max` | 430×932 | 3 |
| `iPad-Pro` | 1024×1366 | 2 |
| `Pixel-7` | 412×915 | 2.625 |
| `Galaxy-S23` | 360×780 | 3 |
| `MacBook-Air` | 1280×800 | 2 |
| `Desktop-1080p` | 1920×1080 | 1 |
| `Desktop-4K` | 3840×2160 | 1 |

### Shell completion

```bash
# Zsh
chromesnap completion zsh > ~/.zsh/completions/_chromesnap

# Bash
chromesnap completion bash > /etc/bash_completion.d/chromesnap
```

## Go library

```bash
go get normalcoder.com/chromesnap
```

### Quick start

```go
import "normalcoder.com/chromesnap/snap"

// Returns []byte
buf, err := snap.Capture("https://example.com")

// Save to file
err := snap.CaptureToFile("https://example.com", "result.png")
```

### With options

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

### Device presets

```go
buf, err := snap.Capture("https://example.com",
    snap.WithDevice(snap.DeviceiPhone15),
)
```

### Reuse browser across captures

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
