package snap

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
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
