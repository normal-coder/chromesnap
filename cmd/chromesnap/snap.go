package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"normalcoder.com/chromesnap/snap"
)

type snapFlags struct {
	output   string
	format   string
	quality  int
	stdout   bool
	jsonMeta bool

	width    int
	height   int
	device   string
	dpr      float64
	darkMode bool

	fullPage bool
	selector string
	clip     string

	waitFor              string
	waitNetwork          bool
	networkIdleThreshold time.Duration
	networkIdleTimeout   time.Duration
	delay                time.Duration
	js          string
	css         string

	basicAuth string
	headers   []string
	cookies   []string
	userAgent string
}

func newSnapCmd(gf *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "snap [URL]",
		Short:        "Capture a screenshot of a single URL",
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sf := snapFlagsFrom(cmd)
			return runSnap(args[0], gf, sf)
		},
		SilenceUsage: true,
	}
	addSnapFlags(cmd)
	return cmd
}

func addSnapFlags(cmd *cobra.Command) {
	f := cmd.Flags()

	// Output
	f.StringP("output", "o", "screenshot.png", "output file path")
	f.StringP("format", "f", "png", "output format: png, jpeg, webp, pdf")
	f.Int("quality", 85, "JPEG/WebP quality (1-100)")
	f.Bool("stdout", false, "write image to stdout instead of file")
	f.BoolP("json", "j", false, "print JSON metadata to stderr after capture")

	// Viewport
	f.IntP("width", "W", 1920, "viewport width")
	f.IntP("height", "H", 1080, "viewport height")
	f.StringP("device", "e", "", "device preset (e.g. iPhone-15, iPad-Pro)")
	f.Float64("dpr", 1.0, "device pixel ratio")
	f.BoolP("dark-mode", "D", false, "enable prefers-color-scheme: dark")

	// Capture area
	f.BoolP("full-page", "F", false, "capture full scrollable page")
	f.StringP("selector", "s", "", "capture a specific CSS selector element")
	f.String("clip", "", "capture region x,y,width,height")

	// Wait & interaction
	f.StringP("wait-for", "w", "", "wait for CSS selector before capturing")
	f.Bool("wait-network", false, "wait for network idle before capturing")
	f.Duration("network-idle-threshold", 500*time.Millisecond, "silence window to consider network idle")
	f.Duration("network-idle-timeout", 10*time.Second, "max time to wait for network idle")
	f.DurationP("delay", "d", 0, "extra wait after page load (e.g. 2s, 500ms)")
	f.String("js", "", "execute JavaScript after page load")
	f.String("css", "", "inject CSS into the page")

	// Auth & request
	f.String("basic-auth", "", "HTTP basic auth in user:pass format")
	f.StringArray("header", nil, "custom request header (repeatable)")
	f.StringArrayP("cookie", "c", nil, "set cookie (repeatable)")
	f.StringP("user-agent", "A", "", "custom User-Agent string")
}

func snapFlagsFrom(cmd *cobra.Command) *snapFlags {
	f := cmd.Flags()
	sf := &snapFlags{}
	sf.output, _ = f.GetString("output")
	sf.format, _ = f.GetString("format")
	sf.quality, _ = f.GetInt("quality")
	sf.stdout, _ = f.GetBool("stdout")
	sf.jsonMeta, _ = f.GetBool("json")
	sf.width, _ = f.GetInt("width")
	sf.height, _ = f.GetInt("height")
	sf.device, _ = f.GetString("device")
	sf.dpr, _ = f.GetFloat64("dpr")
	sf.darkMode, _ = f.GetBool("dark-mode")
	sf.fullPage, _ = f.GetBool("full-page")
	sf.selector, _ = f.GetString("selector")
	sf.clip, _ = f.GetString("clip")
	sf.waitFor, _ = f.GetString("wait-for")
	sf.waitNetwork, _ = f.GetBool("wait-network")
	sf.networkIdleThreshold, _ = f.GetDuration("network-idle-threshold")
	sf.networkIdleTimeout, _ = f.GetDuration("network-idle-timeout")
	sf.delay, _ = f.GetDuration("delay")
	sf.js, _ = f.GetString("js")
	sf.css, _ = f.GetString("css")
	sf.basicAuth, _ = f.GetString("basic-auth")
	sf.headers, _ = f.GetStringArray("header")
	sf.cookies, _ = f.GetStringArray("cookie")
	sf.userAgent, _ = f.GetString("user-agent")
	return sf
}

func runSnap(rawURL string, gf *globalFlags, sf *snapFlags) error {
	opts, err := buildSnapOptions(gf, sf)
	if err != nil {
		return err
	}

	start := time.Now()
	buf, err := snap.Capture(rawURL, opts...)
	if err != nil {
		return err
	}
	elapsed := time.Since(start)

	if sf.stdout {
		_, err = os.Stdout.Write(buf)
		return err
	}

	if err := os.WriteFile(sf.output, buf, 0o644); err != nil {
		return err
	}

	if !gf.quiet {
		fmt.Fprintf(os.Stderr, "saved %s (%d bytes, %s)\n", sf.output, len(buf), elapsed.Round(time.Millisecond))
	}

	if sf.jsonMeta {
		meta := map[string]any{
			"url":     rawURL,
			"file":    sf.output,
			"bytes":   len(buf),
			"elapsed": elapsed.Round(time.Millisecond).String(),
			"format":  sf.format,
		}
		_ = jsonEncoder(os.Stderr).Encode(meta)
	}

	return nil
}

func buildSnapOptions(gf *globalFlags, sf *snapFlags) ([]snap.Option, error) {
	var opts []snap.Option

	opts = append(opts, snap.WithViewport(int64(sf.width), int64(sf.height)))
	opts = append(opts, snap.WithTimeout(gf.timeout))

	if sf.device != "" {
		opts = append(opts, snap.WithDevice(snap.Device(sf.device)))
	}
	if sf.dpr != 1.0 {
		opts = append(opts, snap.WithDPR(sf.dpr))
	}
	if sf.darkMode {
		opts = append(opts, snap.WithDarkMode())
	}
	if sf.fullPage {
		opts = append(opts, snap.WithFullPage())
	}
	if sf.selector != "" {
		opts = append(opts, snap.WithSelector(sf.selector))
	}
	if sf.clip != "" {
		c, err := parseClip(sf.clip)
		if err != nil {
			return nil, err
		}
		opts = append(opts, snap.WithClip(c[0], c[1], c[2], c[3]))
	}
	if sf.waitFor != "" {
		opts = append(opts, snap.WithWaitFor(sf.waitFor))
	}
	if sf.waitNetwork {
		opts = append(opts, snap.WithWaitForNetwork())
		if sf.networkIdleThreshold > 0 {
			opts = append(opts, snap.WithNetworkIdleThreshold(sf.networkIdleThreshold))
		}
		if sf.networkIdleTimeout > 0 {
			opts = append(opts, snap.WithNetworkIdleTimeout(sf.networkIdleTimeout))
		}
	}
	if sf.delay > 0 {
		opts = append(opts, snap.WithDelay(sf.delay))
	}
	if sf.js != "" {
		opts = append(opts, snap.WithJS(sf.js))
	}
	if sf.css != "" {
		opts = append(opts, snap.WithCSS(sf.css))
	}
	if sf.basicAuth != "" {
		opts = append(opts, snap.WithBasicAuth(sf.basicAuth))
	}
	for _, h := range sf.headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header %q: must be Key: Value", h)
		}
		opts = append(opts, snap.WithHeader(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])))
	}
	for _, c := range sf.cookies {
		opts = append(opts, snap.WithCookie(c))
	}
	if sf.userAgent != "" {
		opts = append(opts, snap.WithUserAgent(sf.userAgent))
	}

	opts = append(opts, snap.WithFormat(snap.Format(sf.format)))
	opts = append(opts, snap.WithQuality(sf.quality))

	if gf.chrome != "" {
		opts = append(opts, snap.WithChromePath(gf.chrome))
	}
	if gf.remote != "" {
		opts = append(opts, snap.WithRemoteDebugger(gf.remote))
	}
	if gf.noHeadless {
		opts = append(opts, snap.WithHeadless(false))
	}
	if gf.proxy != "" {
		opts = append(opts, snap.WithProxy(gf.proxy))
	}
	if gf.ignoreCertErrors {
		opts = append(opts, snap.WithIgnoreCertErrors())
	}

	return opts, nil
}

func parseClip(s string) ([4]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return [4]float64{}, fmt.Errorf("--clip must be x,y,width,height")
	}
	var vals [4]float64
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return [4]float64{}, fmt.Errorf("--clip invalid value %q: %w", p, err)
		}
		vals[i] = v
	}
	return vals, nil
}
