package snap

import (
	"bytes"
	"io"
	"log"
	"testing"
)

func testLogger(w io.Writer) *log.Logger {
	return log.New(w, "", 0)
}

// --- resolveViewport ---

func TestResolveViewport_Defaults(t *testing.T) {
	o := defaultOptions()
	w, h, dpr := resolveViewport(o)
	if w != 1920 || h != 1080 || dpr != 1.0 {
		t.Errorf("got %dx%d dpr=%g, want 1920x1080 dpr=1.0", w, h, dpr)
	}
}

func TestResolveViewport_DeviceOverride(t *testing.T) {
	o := defaultOptions()
	o.Device = DeviceiPhone15
	w, h, dpr := resolveViewport(o)
	if w != 390 || h != 844 || dpr != 3 {
		t.Errorf("got %dx%d dpr=%g, want 390x844 dpr=3", w, h, dpr)
	}
}

func TestResolveViewport_UnknownDevice_FallsBack(t *testing.T) {
	o := defaultOptions()
	o.Device = "NonExistent"
	o.Width = 800
	o.Height = 600
	o.DPR = 2.0
	w, h, dpr := resolveViewport(o)
	if w != 800 || h != 600 || dpr != 2.0 {
		t.Errorf("got %dx%d dpr=%g, want 800x600 dpr=2.0", w, h, dpr)
	}
}

func TestResolveViewport_AllDevices(t *testing.T) {
	for _, dev := range []Device{
		DeviceiPhone15, DeviceiPhone15ProMax, DeviceiPadPro,
		DevicePixel7, DeviceGalaxyS23, DeviceMacBookAir,
		DeviceDesktop1080p, DeviceDesktop4K,
	} {
		o := defaultOptions()
		o.Device = dev
		spec := devicePresets[dev]
		w, h, dpr := resolveViewport(o)
		if w != spec.width || h != spec.height || dpr != spec.dpr {
			t.Errorf("device %s: got %dx%d dpr=%g, want %dx%d dpr=%g",
				dev, w, h, dpr, spec.width, spec.height, spec.dpr)
		}
	}
}

// --- parseCookieString ---

func TestParseCookieString_NameValue(t *testing.T) {
	c := parseCookieString("foo=bar")
	if c.name != "foo" || c.value != "bar" {
		t.Errorf("got name=%q value=%q, want name=foo value=bar", c.name, c.value)
	}
}

func TestParseCookieString_WithDomain(t *testing.T) {
	c := parseCookieString("foo=bar; domain=.example.com")
	if c.domain != ".example.com" {
		t.Errorf("domain = %q, want .example.com", c.domain)
	}
}

func TestParseCookieString_WithPath(t *testing.T) {
	c := parseCookieString("foo=bar; path=/api")
	if c.path != "/api" {
		t.Errorf("path = %q, want /api", c.path)
	}
}

func TestParseCookieString_WithDomainAndPath(t *testing.T) {
	c := parseCookieString("foo=bar; domain=.example.com; path=/api")
	if c.domain != ".example.com" {
		t.Errorf("domain = %q, want .example.com", c.domain)
	}
	if c.path != "/api" {
		t.Errorf("path = %q, want /api", c.path)
	}
}

func TestParseCookieString_Empty(t *testing.T) {
	c := parseCookieString("")
	// Should not panic; name/value will be empty
	if c.name != "" {
		t.Errorf("name = %q, want empty", c.name)
	}
}

func TestParseCookieString_NoValue(t *testing.T) {
	c := parseCookieString("foo")
	if c.name != "foo" {
		t.Errorf("name = %q, want foo", c.name)
	}
	if c.value != "" {
		t.Errorf("value = %q, want empty", c.value)
	}
}

func TestParseCookieString_ValueWithEquals(t *testing.T) {
	c := parseCookieString("token=abc=def")
	if c.name != "token" {
		t.Errorf("name = %q, want token", c.name)
	}
	if c.value != "abc=def" {
		t.Errorf("value = %q, want abc=def", c.value)
	}
}

// --- vLogger ---

func TestVLogger_Nil_NoOp(t *testing.T) {
	var v *vLogger
	// Should not panic
	v.printf(1, "hello %s", "world")
}

func TestVLogger_LevelFilter(t *testing.T) {
	var buf bytes.Buffer
	v := &vLogger{
		level: 1,
		log:   testLogger(&buf),
	}
	v.printf(2, "should not appear")
	if buf.Len() != 0 {
		t.Errorf("expected no output for level 2 when logger level=1, got %q", buf.String())
	}
}

func TestVLogger_Output(t *testing.T) {
	var buf bytes.Buffer
	v := &vLogger{
		level: 2,
		log:   testLogger(&buf),
	}
	v.printf(1, "hello %s", "world")
	if buf.Len() == 0 {
		t.Error("expected output for level 1 when logger level=2")
	}
	if !bytes.Contains(buf.Bytes(), []byte("hello world")) {
		t.Errorf("output = %q, want to contain 'hello world'", buf.String())
	}
}

func TestVLogger_MatchesMinLevel(t *testing.T) {
	var buf bytes.Buffer
	v := &vLogger{
		level: 2,
		log:   testLogger(&buf),
	}
	v.printf(2, "exact level")
	if buf.Len() == 0 {
		t.Error("expected output when level matches minLevel")
	}
}

// --- chromedpCtxOpts ---

func TestChromedpCtxOpts_NilOptions(t *testing.T) {
	opts := chromedpCtxOpts(nil)
	if opts != nil {
		t.Errorf("expected nil for nil options, got %d opts", len(opts))
	}
}

func TestChromedpCtxOpts_Level0(t *testing.T) {
	o := defaultOptions()
	o.Verbose = 0
	opts := chromedpCtxOpts(o)
	if opts != nil {
		t.Errorf("expected nil for level 0, got %d opts", len(opts))
	}
}

func TestChromedpCtxOpts_Level1(t *testing.T) {
	o := defaultOptions()
	o.Verbose = 1
	opts := chromedpCtxOpts(o)
	if opts != nil {
		t.Errorf("expected nil for level 1, got %d opts", len(opts))
	}
}

func TestChromedpCtxOpts_Level2(t *testing.T) {
	o := defaultOptions()
	o.Verbose = 2
	opts := chromedpCtxOpts(o)
	if len(opts) != 3 {
		t.Errorf("got %d opts, want 3 for level 2", len(opts))
	}
}

func TestChromedpCtxOpts_Level2_WithCustomWriter(t *testing.T) {
	o := defaultOptions()
	o.Verbose = 2
	var buf dummyWriter
	o.LogOutput = &buf
	opts := chromedpCtxOpts(o)
	if len(opts) != 3 {
		t.Errorf("got %d opts, want 3 for level 2 with custom writer", len(opts))
	}
}

// --- newVLogger ---

func TestNewVLogger_NilOptions(t *testing.T) {
	v := newVLogger(nil)
	if v != nil {
		t.Error("expected nil for nil options")
	}
}

func TestNewVLogger_Verbose0(t *testing.T) {
	o := defaultOptions()
	o.Verbose = 0
	v := newVLogger(o)
	if v != nil {
		t.Error("expected nil for verbose 0")
	}
}

func TestNewVLogger_Verbose1(t *testing.T) {
	o := defaultOptions()
	o.Verbose = 1
	v := newVLogger(o)
	if v == nil {
		t.Fatal("expected non-nil for verbose 1")
	}
	if v.level != 1 {
		t.Errorf("level = %d, want 1", v.level)
	}
}
