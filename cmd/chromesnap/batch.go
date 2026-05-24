package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"normalcoder.com/chromesnap/snap"
)

func newBatchCmd(gf *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch [FILE]",
		Short: "Capture screenshots of multiple URLs",
		Long: `Capture screenshots for a list of URLs.

FILE is a text file with one URL per line.
Alternatively, use --urls to pass a comma-separated list.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatch(cmd, args, gf)
		},
		SilenceUsage: true,
	}

	f := cmd.Flags()
	f.StringP("urls", "u", "", "comma-separated URLs (alternative to FILE)")
	f.StringP("output-dir", "o", "./screenshots", "directory to save screenshots")
	f.StringP("format", "f", "png", "output format: png, jpeg, webp, pdf")
	f.Int("quality", 85, "JPEG/WebP quality (1-100)")
	f.StringP("name-pattern", "n", "{index}_{host}", "file naming pattern: {index}, {host}, {ts}")
	f.BoolP("json", "j", false, "print JSON summary to stdout when done")
	f.IntP("width", "W", 1920, "viewport width")
	f.IntP("height", "H", 1080, "viewport height")
	f.StringP("device", "e", "", "device preset")
	f.BoolP("full-page", "F", false, "capture full scrollable page")
	f.BoolP("dark-mode", "D", false, "enable dark mode")
	f.IntP("concurrency", "p", 3, "number of parallel captures")
	f.Bool("continue-on-error", false, "continue batch if a URL fails")

	return cmd
}

func runBatch(cmd *cobra.Command, args []string, gf *globalFlags) error {
	f := cmd.Flags()

	urlsFlag, _ := f.GetString("urls")
	outputDir, _ := f.GetString("output-dir")
	format, _ := f.GetString("format")
	quality, _ := f.GetInt("quality")
	namePattern, _ := f.GetString("name-pattern")
	jsonOut, _ := f.GetBool("json")
	width, _ := f.GetInt("width")
	height, _ := f.GetInt("height")
	device, _ := f.GetString("device")
	fullPage, _ := f.GetBool("full-page")
	darkMode, _ := f.GetBool("dark-mode")
	concurrency, _ := f.GetInt("concurrency")
	continueOnErr, _ := f.GetBool("continue-on-error")

	var urls []string
	if urlsFlag != "" {
		for u := range strings.SplitSeq(urlsFlag, ",") {
			if u = strings.TrimSpace(u); u != "" {
				urls = append(urls, u)
			}
		}
	} else if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("reading URL file: %w", err)
		}
		for line := range strings.SplitSeq(string(data), "\n") {
			if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
				urls = append(urls, line)
			}
		}
	} else {
		return fmt.Errorf("provide a URL file or --urls flag")
	}

	if len(urls) == 0 {
		return fmt.Errorf("no URLs found")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	var baseOpts []snap.Option
	baseOpts = append(baseOpts, snap.WithViewport(int64(width), int64(height)))
	baseOpts = append(baseOpts, snap.WithTimeout(gf.timeout))
	baseOpts = append(baseOpts, snap.WithFormat(snap.Format(format)))
	if quality != 85 {
		baseOpts = append(baseOpts, snap.WithQuality(quality))
	}
	if device != "" {
		baseOpts = append(baseOpts, snap.WithDevice(snap.Device(device)))
	}
	if fullPage {
		baseOpts = append(baseOpts, snap.WithFullPage())
	}
	if darkMode {
		baseOpts = append(baseOpts, snap.WithDarkMode())
	}
	if gf.chrome != "" {
		baseOpts = append(baseOpts, snap.WithChromePath(gf.chrome))
	}
	if gf.remote != "" {
		baseOpts = append(baseOpts, snap.WithRemoteDebugger(gf.remote))
	}
	if gf.proxy != "" {
		baseOpts = append(baseOpts, snap.WithProxy(gf.proxy))
	}
	if gf.ignoreCertErrors {
		baseOpts = append(baseOpts, snap.WithIgnoreCertErrors())
	}
	if gf.noHeadless {
		baseOpts = append(baseOpts, snap.WithHeadless(false))
	}
	baseOpts = append(baseOpts, snap.WithConcurrency(concurrency))

	browser, err := snap.NewBrowser(baseOpts...)
	if err != nil {
		return fmt.Errorf("launching browser: %w", err)
	}
	defer browser.Close()

	type result struct {
		Index   int    `json:"index"`
		URL     string `json:"url"`
		File    string `json:"file,omitempty"`
		Error   string `json:"error,omitempty"`
		Elapsed string `json:"elapsed"`
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

	done := make(chan result, len(urls))
	sem := make(chan struct{}, concurrency)
	var stopped atomic.Bool

	for range concurrency {
		go func() {
			for j := range jobs {
				if stopped.Load() {
					done <- result{Index: j.index + 1, URL: j.url, Error: "skipped", Elapsed: "0s"}
					continue
				}
				sem <- struct{}{}
				start := time.Now()
				fileName := expandPattern(namePattern, j.index+1, j.url, format)
				filePath := filepath.Join(outputDir, fileName)

				data, captureErr := browser.Capture(j.url)
				elapsed := time.Since(start).Round(time.Millisecond).String()

				r := result{Index: j.index + 1, URL: j.url, Elapsed: elapsed}
				if captureErr != nil {
					r.Error = captureErr.Error()
					if !gf.quiet {
						fmt.Fprintf(os.Stderr, "[%d/%d] ERROR %s: %v\n", j.index+1, len(urls), j.url, captureErr)
					}
					if !continueOnErr {
						stopped.Store(true)
					}
				} else if writeErr := os.WriteFile(filePath, data, 0o644); writeErr != nil {
					r.Error = writeErr.Error()
					if !gf.quiet {
						fmt.Fprintf(os.Stderr, "[%d/%d] ERROR %s: %v\n", j.index+1, len(urls), j.url, writeErr)
					}
					if !continueOnErr {
						stopped.Store(true)
					}
				} else {
					r.File = filePath
					if !gf.quiet {
						fmt.Fprintf(os.Stderr, "[%d/%d] OK    %s → %s (%s)\n", j.index+1, len(urls), j.url, filePath, elapsed)
					}
				}
				done <- r
				<-sem
			}
		}()
	}

	results := make([]result, len(urls))
	var hasError bool
	var firstErr error
	for range urls {
		r := <-done
		results[r.Index-1] = r
		if r.Error != "" && r.Error != "skipped" {
			hasError = true
			if firstErr == nil {
				firstErr = fmt.Errorf("batch failed at URL %d: %s", r.Index, r.Error)
			}
		}
	}

	if jsonOut {
		enc := jsonEncoder(os.Stdout)
		_ = enc.Encode(results)
	}

	if hasError {
		if !continueOnErr {
			return firstErr
		}
		return fmt.Errorf("one or more URLs failed")
	}
	return nil
}

func expandPattern(pattern string, index int, rawURL, format string) string {
	host := rawURL
	if _, after, ok := strings.Cut(rawURL, "://"); ok {
		host = after
	}
	if before, _, ok := strings.Cut(host, "/"); ok {
		host = before
	}
	host = strings.ReplaceAll(host, ":", "_")

	s := strings.ReplaceAll(pattern, "{index}", fmt.Sprintf("%04d", index))
	s = strings.ReplaceAll(s, "{host}", host)
	s = strings.ReplaceAll(s, "{ts}", fmt.Sprintf("%d", time.Now().Unix()))
	return s + "." + format
}
