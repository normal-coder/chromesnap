package snap

import (
	"testing"
	"time"
)

func TestDefaultOptions(t *testing.T) {
	o := defaultOptions()
	if o.Width != 1920 {
		t.Errorf("Width = %d, want 1920", o.Width)
	}
	if o.Height != 1080 {
		t.Errorf("Height = %d, want 1080", o.Height)
	}
	if o.DPR != 1.0 {
		t.Errorf("DPR = %f, want 1.0", o.DPR)
	}
	if o.Format != FormatPNG {
		t.Errorf("Format = %q, want %q", o.Format, FormatPNG)
	}
	if o.Quality != 85 {
		t.Errorf("Quality = %d, want 85", o.Quality)
	}
	if o.Timeout != 30*time.Second {
		t.Errorf("Timeout = %s, want 30s", o.Timeout)
	}
	if !o.Headless {
		t.Error("Headless should default to true")
	}
	if o.NetworkIdleThreshold != 500*time.Millisecond {
		t.Errorf("NetworkIdleThreshold = %s, want 500ms", o.NetworkIdleThreshold)
	}
	if o.NetworkIdleTimeout != 10*time.Second {
		t.Errorf("NetworkIdleTimeout = %s, want 10s", o.NetworkIdleTimeout)
	}
}

func TestWithViewport(t *testing.T) {
	o := defaultOptions()
	WithViewport(800, 600)(o)
	if o.Width != 800 || o.Height != 600 {
		t.Errorf("got %dx%d, want 800x600", o.Width, o.Height)
	}
}

func TestWithDevice(t *testing.T) {
	o := defaultOptions()
	WithDevice(DeviceiPhone15)(o)
	if o.Device != DeviceiPhone15 {
		t.Errorf("Device = %q, want %q", o.Device, DeviceiPhone15)
	}
}

func TestWithFormat(t *testing.T) {
	o := defaultOptions()
	WithFormat(FormatJPEG)(o)
	if o.Format != FormatJPEG {
		t.Errorf("Format = %q, want %q", o.Format, FormatJPEG)
	}
}

func TestWithQuality(t *testing.T) {
	o := defaultOptions()
	WithQuality(50)(o)
	if o.Quality != 50 {
		t.Errorf("Quality = %d, want 50", o.Quality)
	}
}

func TestWithFullPage(t *testing.T) {
	o := defaultOptions()
	if o.FullPage {
		t.Fatal("FullPage should default to false")
	}
	WithFullPage()(o)
	if !o.FullPage {
		t.Error("FullPage should be true after WithFullPage")
	}
}

func TestWithSelector(t *testing.T) {
	o := defaultOptions()
	WithSelector("#app")(o)
	if o.Selector != "#app" {
		t.Errorf("Selector = %q, want %q", o.Selector, "#app")
	}
}

func TestWithClip(t *testing.T) {
	o := defaultOptions()
	WithClip(10, 20, 300, 400)(o)
	if o.Clip == nil {
		t.Fatal("Clip should not be nil")
	}
	if o.Clip.X != 10 || o.Clip.Y != 20 || o.Clip.Width != 300 || o.Clip.Height != 400 {
		t.Errorf("Clip = %+v, want {10,20,300,400}", o.Clip)
	}
}

func TestWithWaitFor(t *testing.T) {
	o := defaultOptions()
	WithWaitFor("#loaded")(o)
	if o.WaitForSelector != "#loaded" {
		t.Errorf("WaitForSelector = %q, want %q", o.WaitForSelector, "#loaded")
	}
}

func TestWithWaitForNetwork(t *testing.T) {
	o := defaultOptions()
	WithWaitForNetwork()(o)
	if !o.WaitForNetwork {
		t.Error("WaitForNetwork should be true after WithWaitForNetwork")
	}
}

func TestWithNetworkIdleThreshold(t *testing.T) {
	o := defaultOptions()
	WithNetworkIdleThreshold(1 * time.Second)(o)
	if o.NetworkIdleThreshold != 1*time.Second {
		t.Errorf("NetworkIdleThreshold = %s, want 1s", o.NetworkIdleThreshold)
	}
}

func TestWithNetworkIdleTimeout(t *testing.T) {
	o := defaultOptions()
	WithNetworkIdleTimeout(20 * time.Second)(o)
	if o.NetworkIdleTimeout != 20*time.Second {
		t.Errorf("NetworkIdleTimeout = %s, want 20s", o.NetworkIdleTimeout)
	}
}

func TestWithDelay(t *testing.T) {
	o := defaultOptions()
	WithDelay(2 * time.Second)(o)
	if o.Delay != 2*time.Second {
		t.Errorf("Delay = %s, want 2s", o.Delay)
	}
}

func TestWithTimeout(t *testing.T) {
	o := defaultOptions()
	WithTimeout(60 * time.Second)(o)
	if o.Timeout != 60*time.Second {
		t.Errorf("Timeout = %s, want 60s", o.Timeout)
	}
}

func TestWithDarkMode(t *testing.T) {
	o := defaultOptions()
	WithDarkMode()(o)
	if !o.DarkMode {
		t.Error("DarkMode should be true after WithDarkMode")
	}
}

func TestWithBasicAuth(t *testing.T) {
	o := defaultOptions()
	WithBasicAuth("user:pass")(o)
	if o.BasicAuth != "user:pass" {
		t.Errorf("BasicAuth = %q, want %q", o.BasicAuth, "user:pass")
	}
}

func TestWithHeader(t *testing.T) {
	o := defaultOptions()
	WithHeader("X-Custom", "value1")(o)
	WithHeader("X-Other", "value2")(o)
	if len(o.Headers) != 2 {
		t.Fatalf("Headers count = %d, want 2", len(o.Headers))
	}
	if o.Headers["X-Custom"] != "value1" {
		t.Errorf("Headers[X-Custom] = %q, want %q", o.Headers["X-Custom"], "value1")
	}
	if o.Headers["X-Other"] != "value2" {
		t.Errorf("Headers[X-Other] = %q, want %q", o.Headers["X-Other"], "value2")
	}
}

func TestWithCookie(t *testing.T) {
	o := defaultOptions()
	WithCookie("foo=bar")(o)
	WithCookie("baz=qux")(o)
	if len(o.Cookies) != 2 {
		t.Fatalf("Cookies count = %d, want 2", len(o.Cookies))
	}
	if o.Cookies[0] != "foo=bar" {
		t.Errorf("Cookies[0] = %q, want %q", o.Cookies[0], "foo=bar")
	}
	if o.Cookies[1] != "baz=qux" {
		t.Errorf("Cookies[1] = %q, want %q", o.Cookies[1], "baz=qux")
	}
}

func TestWithUserAgent(t *testing.T) {
	o := defaultOptions()
	WithUserAgent("CustomAgent/1.0")(o)
	if o.UserAgent != "CustomAgent/1.0" {
		t.Errorf("UserAgent = %q, want %q", o.UserAgent, "CustomAgent/1.0")
	}
}

func TestWithProxy(t *testing.T) {
	o := defaultOptions()
	WithProxy("http://proxy:8080")(o)
	if o.Proxy != "http://proxy:8080" {
		t.Errorf("Proxy = %q, want %q", o.Proxy, "http://proxy:8080")
	}
}

func TestWithChromePath(t *testing.T) {
	o := defaultOptions()
	WithChromePath("/usr/bin/chrome")(o)
	if o.ChromePath != "/usr/bin/chrome" {
		t.Errorf("ChromePath = %q, want %q", o.ChromePath, "/usr/bin/chrome")
	}
}

func TestWithRemoteDebugger(t *testing.T) {
	o := defaultOptions()
	WithRemoteDebugger("ws://localhost:9222")(o)
	if o.RemoteDebuggerURL != "ws://localhost:9222" {
		t.Errorf("RemoteDebuggerURL = %q, want %q", o.RemoteDebuggerURL, "ws://localhost:9222")
	}
}

func TestWithHeadless(t *testing.T) {
	o := defaultOptions()
	WithHeadless(false)(o)
	if o.Headless {
		t.Error("Headless should be false after WithHeadless(false)")
	}
}

func TestWithIgnoreCertErrors(t *testing.T) {
	o := defaultOptions()
	WithIgnoreCertErrors()(o)
	if !o.IgnoreCertErrors {
		t.Error("IgnoreCertErrors should be true after WithIgnoreCertErrors")
	}
}

func TestWithConcurrency(t *testing.T) {
	o := defaultOptions()
	WithConcurrency(5)(o)
	if o.Concurrency != 5 {
		t.Errorf("Concurrency = %d, want 5", o.Concurrency)
	}
}

func TestWithVerbose(t *testing.T) {
	o := defaultOptions()
	WithVerbose(2)(o)
	if o.Verbose != 2 {
		t.Errorf("Verbose = %d, want 2", o.Verbose)
	}
}

func TestWithLogOutput(t *testing.T) {
	o := defaultOptions()
	// Use a simple writer to verify assignment
	var buf dummyWriter
	WithLogOutput(&buf)(o)
	if o.LogOutput != &buf {
		t.Error("LogOutput should be set to the provided writer")
	}
}

func TestWithDPR(t *testing.T) {
	o := defaultOptions()
	WithDPR(2.5)(o)
	if o.DPR != 2.5 {
		t.Errorf("DPR = %f, want 2.5", o.DPR)
	}
}

func TestWithJS(t *testing.T) {
	o := defaultOptions()
	WithJS("document.title")(o)
	if o.JS != "document.title" {
		t.Errorf("JS = %q, want %q", o.JS, "document.title")
	}
}

func TestWithCSS(t *testing.T) {
	o := defaultOptions()
	WithCSS("body { background: red }")(o)
	if o.CSS != "body { background: red }" {
		t.Errorf("CSS = %q, want %q", o.CSS, "body { background: red }")
	}
}

// dummyWriter is a minimal io.Writer for testing WithLogOutput.
type dummyWriter struct{}

func (d *dummyWriter) Write(p []byte) (n int, err error) { return len(p), nil }
