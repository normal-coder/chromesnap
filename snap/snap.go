package snap

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Capture takes a screenshot of the given URL and returns the image bytes.
func Capture(rawURL string, opts ...Option) ([]byte, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	return capture(rawURL, o)
}

// CaptureToFile takes a screenshot and writes it to the given file path.
func CaptureToFile(rawURL, filePath string, opts ...Option) error {
	buf, err := Capture(rawURL, opts...)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, buf, 0o644)
}

func capture(rawURL string, o *Options) ([]byte, error) {
	allocCtx, cancelAlloc, err := newAllocator(o)
	if err != nil {
		return nil, err
	}
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	ctx, cancelTimeout := context.WithTimeout(ctx, o.Timeout)
	defer cancelTimeout()

	var buf []byte
	tasks, err := buildTasks(rawURL, o, &buf)
	if err != nil {
		return nil, err
	}

	if err := chromedp.Run(ctx, tasks...); err != nil {
		return nil, fmt.Errorf("capture %s: %w", rawURL, err)
	}
	return buf, nil
}

func newAllocator(o *Options) (context.Context, context.CancelFunc, error) {
	if o.RemoteDebuggerURL != "" {
		ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), o.RemoteDebuggerURL)
		return ctx, cancel, nil
	}

	chromePath, err := FindChrome(o.ChromePath)
	if err != nil {
		return nil, nil, err
	}

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.WindowSize(int(o.Width), int(o.Height)),
	)

	if !o.Headless {
		allocOpts = append(allocOpts, chromedp.Flag("headless", false))
	}
	if o.IgnoreCertErrors {
		allocOpts = append(allocOpts, chromedp.Flag("ignore-certificate-errors", true))
	}
	if o.Proxy != "" {
		allocOpts = append(allocOpts, chromedp.Flag("proxy-server", o.Proxy))
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	return ctx, cancel, nil
}

func buildTasks(rawURL string, o *Options, buf *[]byte) (chromedp.Tasks, error) {
	var tasks chromedp.Tasks

	// Viewport / device emulation
	width, height, dpr := resolveViewport(o)
	tasks = append(tasks, chromedp.EmulateViewport(width, height, chromedp.EmulateScale(dpr)))

	// Dark mode
	if o.DarkMode {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetEmulatedMedia().WithFeatures([]*emulation.MediaFeature{
				{Name: "prefers-color-scheme", Value: "dark"},
			}).Do(ctx)
		}))
	}

	// Basic auth via URL embedding
	if o.BasicAuth != "" {
		parts := strings.SplitN(o.BasicAuth, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("--basic-auth must be in user:pass format")
		}
		u, err := url.Parse(rawURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		u.User = url.UserPassword(parts[0], parts[1])
		rawURL = u.String()
	}

	// Custom headers
	if len(o.Headers) > 0 {
		headers := make(network.Headers, len(o.Headers))
		for k, v := range o.Headers {
			headers[k] = v
		}
		tasks = append(tasks, network.SetExtraHTTPHeaders(headers))
	}

	// User-Agent
	if o.UserAgent != "" {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetUserAgentOverride(o.UserAgent).Do(ctx)
		}))
	}

	// Navigate
	tasks = append(tasks, chromedp.Navigate(rawURL))

	// Cookies (set after navigation so domain is known)
	for _, cookie := range o.Cookies {
		c := parseCookieString(cookie)
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			return network.SetCookie(c.name, c.value).
				WithDomain(c.domain).
				Do(ctx)
		}))
	}

	// Wait for network idle
	if o.WaitForNetwork {
		tasks = append(tasks, chromedp.ActionFunc(waitForNetworkIdle))
	}

	// Wait for selector
	if o.WaitForSelector != "" {
		tasks = append(tasks, chromedp.WaitVisible(o.WaitForSelector))
	}

	// Inject CSS
	if o.CSS != "" {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			var res string
			return chromedp.Evaluate(fmt.Sprintf(
				`(function(){var s=document.createElement('style');s.textContent=%q;document.head.appendChild(s)})()`,
				o.CSS,
			), &res).Do(ctx)
		}))
	}

	// Execute JS
	if o.JS != "" {
		var res any
		tasks = append(tasks, chromedp.Evaluate(o.JS, &res))
	}

	// Delay
	if o.Delay > 0 {
		tasks = append(tasks, chromedp.Sleep(o.Delay))
	}

	// Screenshot
	tasks = append(tasks, screenshotAction(o, buf))

	return tasks, nil
}

func screenshotAction(o *Options, buf *[]byte) chromedp.Action {
	switch o.Format {
	case FormatPDF:
		return chromedp.ActionFunc(func(ctx context.Context) error {
			data, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				Do(ctx)
			if err != nil {
				return err
			}
			*buf = data
			return nil
		})
	case FormatJPEG:
		return chromedp.ActionFunc(func(ctx context.Context) error {
			return captureScreenshot(ctx, o, buf, page.CaptureScreenshotFormatJpeg, int64(o.Quality))
		})
	case FormatWebP:
		return chromedp.ActionFunc(func(ctx context.Context) error {
			return captureScreenshot(ctx, o, buf, page.CaptureScreenshotFormatWebp, int64(o.Quality))
		})
	default: // PNG
		return chromedp.ActionFunc(func(ctx context.Context) error {
			return captureScreenshot(ctx, o, buf, page.CaptureScreenshotFormatPng, 0)
		})
	}
}

func captureScreenshot(ctx context.Context, o *Options, buf *[]byte, format page.CaptureScreenshotFormat, quality int64) error {
	params := page.CaptureScreenshot().WithFormat(format)
	if quality > 0 {
		params = params.WithQuality(quality)
	}

	if o.FullPage {
		// Scroll to get full page dimensions then capture
		var pageW, pageH float64
		if err := chromedp.Evaluate(`(function(){return [document.documentElement.scrollWidth, document.documentElement.scrollHeight]})()`, &[]any{&pageW, &pageH}); err == nil {
			params = params.WithClip(&page.Viewport{
				X: 0, Y: 0,
				Width: pageW, Height: pageH,
				Scale: 1,
			})
		}
	} else if o.Selector != "" {
		var rect map[string]float64
		err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(
			`(function(){var r=document.querySelector(%q).getBoundingClientRect();return {x:r.left,y:r.top,w:r.width,h:r.height}})()`,
			o.Selector,
		), &rect))
		if err != nil {
			return fmt.Errorf("selector %q not found: %w", o.Selector, err)
		}
		params = params.WithClip(&page.Viewport{
			X: rect["x"], Y: rect["y"],
			Width: rect["w"], Height: rect["h"],
			Scale: 1,
		})
	} else if o.Clip != nil {
		params = params.WithClip(&page.Viewport{
			X: o.Clip.X, Y: o.Clip.Y,
			Width: o.Clip.Width, Height: o.Clip.Height,
			Scale: 1,
		})
	}

	data, err := params.Do(ctx)
	if err != nil {
		return err
	}
	*buf = data
	return nil
}

func resolveViewport(o *Options) (width, height int64, dpr float64) {
	width, height, dpr = o.Width, o.Height, o.DPR
	if o.Device != "" {
		if spec, ok := devicePresets[o.Device]; ok {
			width, height, dpr = spec.width, spec.height, spec.dpr
		}
	}
	return
}

type cookieParsed struct {
	name, value, domain string
}

func parseCookieString(s string) cookieParsed {
	var c cookieParsed
	parts := strings.Split(s, ";")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if i == 0 {
			kv := strings.SplitN(part, "=", 2)
			c.name = strings.TrimSpace(kv[0])
			if len(kv) > 1 {
				c.value = strings.TrimSpace(kv[1])
			}
		} else {
			lower := strings.ToLower(part)
			if strings.HasPrefix(lower, "domain=") {
				c.domain = strings.TrimSpace(part[7:])
			}
		}
	}
	return c
}

func waitForNetworkIdle(ctx context.Context) error {
	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}