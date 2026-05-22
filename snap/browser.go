package snap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/chromedp/chromedp"
)

var macOSChromePaths = []string{
	"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	"/Applications/Chromium.app/Contents/MacOS/Chromium",
	"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
	"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
	"/usr/local/bin/chromium",
}

var linuxChromePaths = []string{
	"/usr/bin/google-chrome",
	"/usr/bin/google-chrome-stable",
	"/usr/bin/chromium-browser",
	"/usr/bin/chromium",
	"/snap/bin/chromium",
}

var chromeCommandNames = []string{
	"google-chrome",
	"google-chrome-stable",
	"chromium-browser",
	"chromium",
}

// FindChrome returns the path to a Chrome/Chromium executable.
// Resolution order:
//  1. Explicit path argument (from --chrome flag)
//  2. CHROME_PATH environment variable
//  3. Platform-specific well-known paths
//  4. PATH lookup by command name
func FindChrome(explicitPath string) (string, error) {
	if explicitPath != "" {
		if err := checkExec(explicitPath); err != nil {
			return "", fmt.Errorf("--chrome %q: %w", explicitPath, err)
		}
		return explicitPath, nil
	}

	if p := os.Getenv("CHROME_PATH"); p != "" {
		if err := checkExec(p); err != nil {
			return "", fmt.Errorf("CHROME_PATH=%q: %w", p, err)
		}
		return p, nil
	}

	var fixedPaths []string
	if runtime.GOOS == "darwin" {
		fixedPaths = macOSChromePaths
	} else {
		fixedPaths = linuxChromePaths
	}

	for _, p := range fixedPaths {
		if checkExec(p) == nil {
			return p, nil
		}
	}

	for _, name := range chromeCommandNames {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}

	return "", chromeNotFoundError()
}

func checkExec(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("%q is not executable", path)
	}
	return nil
}

func chromeNotFoundError() error {
	var hint string
	if runtime.GOOS == "darwin" {
		hint = "  macOS:  brew install --cask google-chrome\n"
	} else {
		hint = "  Linux:  sudo apt install google-chrome-stable\n" +
			"          sudo apt install chromium-browser\n"
	}
	return fmt.Errorf(
		"Chrome/Chromium not found.\n\n"+
			"Install options:\n%s\n"+
			"Or specify path manually:\n"+
			"  chromesnap --chrome /path/to/chrome <url>\n"+
			"  export CHROME_PATH=/path/to/chrome",
		hint,
	)
}

// Browser is a reusable Chrome instance. Use NewBrowser when capturing
// multiple pages to avoid the overhead of launching a new process per capture.
type Browser struct {
	allocCtx context.Context
	cancel   context.CancelFunc
	base     *Options
}

// NewBrowser launches a Chrome instance that can be reused across captures.
func NewBrowser(opts ...Option) (*Browser, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	allocCtx, cancel, err := newAllocator(o)
	if err != nil {
		return nil, err
	}
	return &Browser{allocCtx: allocCtx, cancel: cancel, base: o}, nil
}

// Close shuts down the underlying Chrome process.
func (b *Browser) Close() {
	b.cancel()
}

// Capture takes a screenshot of rawURL in a new tab, reusing the existing
// Chrome process. Per-call options override the Browser's base options.
func (b *Browser) Capture(rawURL string, opts ...Option) ([]byte, error) {
	o := *b.base
	for _, opt := range opts {
		opt(&o)
	}

	ctx, cancel := chromedp.NewContext(b.allocCtx)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, o.Timeout)
	defer cancelTimeout()

	var buf []byte
	tasks, err := buildTasks(rawURL, &o, &buf)
	if err != nil {
		return nil, err
	}
	if err := chromedp.Run(ctx, tasks...); err != nil {
		return nil, fmt.Errorf("capture %s: %w", rawURL, err)
	}
	return buf, nil
}

// CaptureResult holds the outcome of a single URL in a batch capture.
type CaptureResult struct {
	URL   string
	Data  []byte
	Error error
}

// CaptureAll captures all URLs concurrently, reusing the same Chrome process.
// Concurrency is controlled via WithConcurrency (default 3).
func (b *Browser) CaptureAll(urls []string, opts ...Option) ([]CaptureResult, error) {
	o := *b.base
	for _, opt := range opts {
		opt(&o)
	}

	concurrency := o.Concurrency
	if concurrency <= 0 {
		concurrency = 3
	}

	type job struct {
		index int
		url   string
	}

	jobs := make(chan job, len(urls))
	for i, u := range urls {
		jobs <- job{i, u}
	}
	close(jobs)

	results := make([]CaptureResult, len(urls))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for range concurrency {
		wg.Go(func() {
			for j := range jobs {
				sem <- struct{}{}
				data, err := b.Capture(j.url, opts...)
				results[j.index] = CaptureResult{URL: j.url, Data: data, Error: err}
				<-sem
			}
		})
	}

	wg.Wait()

	var errCount int
	for _, r := range results {
		if r.Error != nil {
			errCount++
		}
	}
	if errCount > 0 {
		return results, fmt.Errorf("%d of %d captures failed", errCount, len(urls))
	}
	return results, nil
}
