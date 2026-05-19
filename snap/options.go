package snap

import "time"

// Format represents the screenshot output format.
type Format string

const (
	FormatPNG  Format = "png"
	FormatJPEG Format = "jpeg"
	FormatWebP Format = "webp"
	FormatPDF  Format = "pdf"
)

// Device represents a predefined device preset.
type Device string

const (
	DeviceiPhone15       Device = "iPhone-15"
	DeviceiPhone15ProMax Device = "iPhone-15-Pro-Max"
	DeviceiPadPro        Device = "iPad-Pro"
	DevicePixel7         Device = "Pixel-7"
	DeviceGalaxyS23      Device = "Galaxy-S23"
	DeviceMacBookAir     Device = "MacBook-Air"
	DeviceDesktop1080p   Device = "Desktop-1080p"
	DeviceDesktop4K      Device = "Desktop-4K"
)

type deviceSpec struct {
	width  int64
	height int64
	dpr    float64
}

var devicePresets = map[Device]deviceSpec{
	DeviceiPhone15:       {390, 844, 3},
	DeviceiPhone15ProMax: {430, 932, 3},
	DeviceiPadPro:        {1024, 1366, 2},
	DevicePixel7:         {412, 915, 2.625},
	DeviceGalaxyS23:      {360, 780, 3},
	DeviceMacBookAir:     {1280, 800, 2},
	DeviceDesktop1080p:   {1920, 1080, 1},
	DeviceDesktop4K:      {3840, 2160, 1},
}

// Options holds all configuration for a screenshot capture.
type Options struct {
	// Viewport
	Width  int64
	Height int64
	DPR    float64
	Device Device

	// Output
	Format  Format
	Quality int

	// Capture area
	FullPage bool
	Selector string
	Clip     *ClipRect

	// Wait & interaction
	WaitForSelector string
	WaitForNetwork  bool
	Delay           time.Duration
	Timeout         time.Duration
	JS              string
	CSS             string

	// Theme
	DarkMode bool

	// Request
	BasicAuth string
	Headers   map[string]string
	Cookies   []string
	UserAgent string
	Proxy     string

	// Browser
	ChromePath        string
	RemoteDebuggerURL string
	Headless          bool
	IgnoreCertErrors  bool

	// Batch
	Concurrency int
}

// ClipRect defines a rectangular region to capture.
type ClipRect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func defaultOptions() *Options {
	return &Options{
		Width:   1920,
		Height:  1080,
		DPR:     1.0,
		Format:  FormatPNG,
		Quality: 85,
		Timeout: 30 * time.Second,
		Headless: true,
	}
}

// Option is a functional option for configuring a screenshot capture.
type Option func(*Options)

func WithViewport(width, height int64) Option {
	return func(o *Options) {
		o.Width = width
		o.Height = height
	}
}

func WithDPR(dpr float64) Option {
	return func(o *Options) { o.DPR = dpr }
}

func WithDevice(d Device) Option {
	return func(o *Options) { o.Device = d }
}

func WithFormat(f Format) Option {
	return func(o *Options) { o.Format = f }
}

func WithQuality(q int) Option {
	return func(o *Options) { o.Quality = q }
}

func WithFullPage() Option {
	return func(o *Options) { o.FullPage = true }
}

func WithSelector(sel string) Option {
	return func(o *Options) { o.Selector = sel }
}

func WithClip(x, y, w, h float64) Option {
	return func(o *Options) { o.Clip = &ClipRect{x, y, w, h} }
}

func WithWaitFor(sel string) Option {
	return func(o *Options) { o.WaitForSelector = sel }
}

func WithWaitForNetwork() Option {
	return func(o *Options) { o.WaitForNetwork = true }
}

func WithDelay(d time.Duration) Option {
	return func(o *Options) { o.Delay = d }
}

func WithTimeout(d time.Duration) Option {
	return func(o *Options) { o.Timeout = d }
}

func WithJS(js string) Option {
	return func(o *Options) { o.JS = js }
}

func WithCSS(css string) Option {
	return func(o *Options) { o.CSS = css }
}

func WithDarkMode() Option {
	return func(o *Options) { o.DarkMode = true }
}

func WithBasicAuth(userpass string) Option {
	return func(o *Options) { o.BasicAuth = userpass }
}

func WithHeader(key, value string) Option {
	return func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers[key] = value
	}
}

func WithCookie(cookie string) Option {
	return func(o *Options) { o.Cookies = append(o.Cookies, cookie) }
}

func WithUserAgent(ua string) Option {
	return func(o *Options) { o.UserAgent = ua }
}

func WithProxy(proxy string) Option {
	return func(o *Options) { o.Proxy = proxy }
}

func WithChromePath(path string) Option {
	return func(o *Options) { o.ChromePath = path }
}

func WithRemoteDebugger(url string) Option {
	return func(o *Options) { o.RemoteDebuggerURL = url }
}

func WithHeadless(v bool) Option {
	return func(o *Options) { o.Headless = v }
}

func WithIgnoreCertErrors() Option {
	return func(o *Options) { o.IgnoreCertErrors = true }
}

func WithConcurrency(n int) Option {
	return func(o *Options) { o.Concurrency = n }
}
